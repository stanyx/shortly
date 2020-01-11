package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	"github.com/kr/pretty"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger"
	bolt "go.etcd.io/bbolt"
	"golang.org/x/time/rate"

	"shortly/api"
	"shortly/cache"
	"shortly/config"
	"shortly/server"
	"shortly/storage"
	"shortly/utils"

	"shortly/app/accounts"
	"shortly/app/billing"
	"shortly/app/campaigns"
	"shortly/app/data"
	"shortly/app/links"
	"shortly/app/rbac"
	"shortly/app/tags"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func LoadCacheFromDatabase(repo *links.LinksRepository, urlCache cache.UrlCache) error {

	rows, err := repo.GetAllLinks()
	if err != nil {
		return err
	}

	for _, r := range rows {
		urlCache.Store(r.Short, r.Long)
	}

	return nil
}

func RunMigrations(database *sql.DB) error {

	driver, err := postgres.WithInstance(database, &postgres.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance("file://storage/migrations", "postgres", driver)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil {
		return err
	}

	return nil
}

// @title Shotly API
// @version 1.0
// @description Url shortener web application.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host store.swagger.io
// @BasePath /api/v1
func main() {

	rand.Seed(time.Now().UnixNano())

	logger := log.New(os.Stdout, "", log.LstdFlags|log.Llongfile)

	serverConfig := server.ParseServerOptions()
	appConfig, err := config.ReadConfig(serverConfig.ConfigPath)
	if err != nil {
		logger.Fatal(err)
	}

	dbConfig := appConfig.Database

	dbConnString := os.Getenv("DATABASE_URL")

	if dbConnString == "" {
		dbConnString = fmt.Sprintf("host=%v port=%v user=%v password=%v dbname=%v sslmode=%v",
			dbConfig.Host,
			dbConfig.Port,
			dbConfig.User,
			dbConfig.Password,
			dbConfig.Database,
			dbConfig.SSLMode,
		)
	}

	// connection to databases

	database, err := storage.StartDB(dbConnString)
	if err != nil {
		logger.Fatal(err)
	}

	if err := RunMigrations(database); err != nil && err.Error() != "no change" {
		logger.Fatal(err)
	}

	billingDataStorage, err := bolt.Open(appConfig.Billing.Dir+"/billing.db", 0666, nil)
	if err != nil {
		logger.Fatal(err)
	}

	linksStorage, err := bolt.Open(appConfig.LinkDB.Dir+"/links.db", 0666, nil)
	if err != nil {
		logger.Fatal(err)
	}

	linksRepository := &links.LinksRepository{DB: database, Logger: logger}

	billingRepository := &billing.BillingRepository{DB: database}
	billingLimiter := &billing.BillingLimiter{
		Repo:    billingRepository,
		DB:      billingDataStorage,
		UrlRepo: linksRepository,
		Logger:  logger,
	}
	if err := billingLimiter.LoadData(); err != nil {
		logger.Fatal(err)
	}

	urlBillingLimit := api.BillingLimitMiddleware("url_limit", billingLimiter, logger)

	historyDB := &data.HistoryDB{DB: linksStorage, Limiter: billingLimiter}
	campaignsRepository := &campaigns.Repository{DB: database, HistoryDB: historyDB, Logger: logger}

	// cache initialization

	var urlCache cache.UrlCache

	cacheConfig := appConfig.Cache

	pretty.Printf("\n\nCONFIG\n======\n\n%# v\n\n", appConfig)

	switch cacheConfig.CacheType {
	case "memory", "":
		log.Println("CACHE: use MEMORY")
		urlCache = cache.NewMemoryCache()
	case "memcached":
		log.Println("CACHE: use MEMCACHED")
		urlCache = cache.NewMemcachedCache(cacheConfig.Memcached.ServerList, logger)
	case "boltdb":
		log.Println("CACHE: use BOLT_DB")
		urlDataStorage, err := bolt.Open(cacheConfig.BoltDB.Dir+"/urls.db", 0666, nil)
		if err != nil {
			logger.Fatal(err)
		}
		urlCache, err = cache.NewBoltDBCache(urlDataStorage, logger)
		if err != nil {
			logger.Fatal(err)
		}
	}

	err = LoadCacheFromDatabase(linksRepository, urlCache)
	if err != nil {
		logger.Fatal(err)
	}

	// ** route declaration

	// public api

	serverPort := os.Getenv("PORT")
	if serverPort == "" {
		serverPort = fmt.Sprintf("%v", serverConfig.Port)
	}

	r := chi.NewRouter()

	r.Use(utils.RateLimit(func(w http.ResponseWriter, r *http.Request) *rate.Limiter {

		claims, _ := api.ParseToken(w, r, appConfig.Auth)

		accountID := claims.AccountID
		if accountID == 0 {
			return rate.NewLimiter(rate.Every(rate.InfDuration), 1000)
		}

		option, _ := billingLimiter.GetOptionValue("rate_limit", accountID)

		val := "10,100"
		if option != nil {
			val = option.Value
		}

		parts := strings.Split(val, ",")
		rateV, _ := strconv.ParseInt(parts[0], 0, 64)
		if rateV == 0 {
			rateV = 10
		}
		burstV, _ := strconv.ParseInt(parts[1], 0, 64)
		if burstV == 0 {
			burstV = 1000
		}

		return rate.NewLimiter(rate.Every(time.Duration(rateV)*time.Second), int(burstV))
	}))

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL(fmt.Sprintf("http://localhost:%d/swagger/doc.json", serverConfig.Port)), //The url pointing to API definition"
	))

	fs := http.FileServer(http.Dir("static"))
	fsHandler := http.StripPrefix("/static", fs)

	r.Get("/static/*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fsHandler.ServeHTTP(w, r)
	}))

	r.Get("/health", utils.HealthCheck(
		[]utils.HealthChecker{
			utils.HealthCheckFunc(func(_ context.Context) error {
				return database.Ping()
			}),
			utils.HealthCheckFunc(func(_ context.Context) error {
				return urlCache.Ping()
			}),
		},
		logger,
	))

	totalLinkCreatedPromMiddleware := utils.PrometheusMiddleware("totalLinksCreated", "TODO description")
	r.Get("/api/v1/links", api.GetURLList(linksRepository, logger))
	r.Post("/api/v1/links", totalLinkCreatedPromMiddleware(
		api.CreateLink(linksRepository, urlCache, logger)))

	// storage metadata preparation
	err = billingDataStorage.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("billing"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})

	if err != nil {
		logger.Fatal(err)
	}

	err = historyDB.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("details"))
		if err != nil {
			return fmt.Errorf("create bucket error, cause: %+v", err)
		}
		return nil
	})

	if err != nil {
		logger.Fatal(err)
	}

	// private api (with authorized access)
	enforcer, err := rbac.NewEnforcer(database, appConfig.Casbin)
	if err != nil {
		logger.Fatal(err)
	}
	permissionRegistry := make(map[string]rbac.Permission)
	auth := api.AuthMiddleware(enforcer, appConfig.Auth, permissionRegistry)

	rbacRepository := &rbac.RbacRepository{
		DB:       database,
		Logger:   logger,
		Enforcer: enforcer,
	}

	//   billing api

	r.Get("/api/v1/billing/plans", api.ListBillingPlans(billingRepository, logger))

	applyBillingPlansPromMiddleware := utils.PrometheusMiddleware("billingApplied", "TODO description")
	r.Post("/api/v1/billing/apply", auth(
		rbac.NewPermission("/api/v1/billing/apply", "apply_billingplan", "POST"),
		applyBillingPlansPromMiddleware(
			api.ApplyBillingPlan(billingRepository, billingLimiter, appConfig.Billing.Payment, logger),
		),
	))

	// TODO remove to routes.go

	tagsRepository := &tags.TagsRepository{
		DB:     database,
		Logger: logger,
	}

	r.Post("/api/v1/tags/create", auth(
		rbac.NewPermission("/api/v1/tags/create", "create_tag", "POST"),
		api.AddTagToLink(tagsRepository, logger),
	))

	r.Delete("/api/v1/tags/delete", auth(
		rbac.NewPermission("/api/v1/tags/delete", "delete_tag", "POST"),
		api.DeleteTagFromLink(tagsRepository, logger),
	))

	// links api
	r.Get("/api/v1/users/links", auth(
		rbac.NewPermission("/api/v1/users/links", "read_links", "GET"),
		api.GetUserURLList(linksRepository, logger),
	))

	r.Get("/api/v1/users/links/clicks", auth(
		rbac.NewPermission("/api/v1/users/links/clicks", "get_links_clicks", "GET"),
		api.GetClicksData(historyDB, logger),
	))

	r.Post("/api/v1/users/links/add_group", auth(
		rbac.NewPermission("/api/v1/users/links/add_group", "add_link_to_group", "POST"),
		api.AddUrlToGroup(linksRepository, logger),
	))

	r.Delete("/api/v1/users/links/delete_group", auth(
		rbac.NewPermission("/api/v1/users/links/delete_group", "delete_link_from_group", "DELETE"),
		api.DeleteUrlFromGroup(linksRepository, logger),
	))

	// account api
	usersRepository := &accounts.UsersRepository{DB: database}

	r.Post("/api/v1/registration", api.RegisterAccount(usersRepository, logger))
	r.Post("/api/v1/accounts/users", api.AddUser(usersRepository, logger))
	r.Post("/api/v1/login", api.Login(usersRepository, logger, appConfig.Auth))

	r.Post("/api/v1/users/links/create", auth(
		rbac.NewPermission("/api/v1/users/links/create", "create_link", "POST"),
		urlBillingLimit(api.CreateUserLink(linksRepository, historyDB, urlCache, billingLimiter, logger)),
	))

	r.Post("/api/v1/users/links/upload", auth(
		rbac.NewPermission("/api/v1/users/links/upload", "upload_links", "POST"),
		api.UploadLinksInBulk(billingLimiter, linksRepository, historyDB, urlCache, logger),
	))

	r.Delete("/api/v1/users/links/delete", auth(
		rbac.NewPermission("/api/v1/users/links/delete", "delete_link", "DELETE"),
		api.DeleteUserLink(linksRepository, urlCache, billingLimiter, logger),
	))

	r.Post("/api/v1/users/groups/create", auth(
		rbac.NewPermission("/api/v1/users/groups/create", "create_group", "POST"),
		api.AddGroup(usersRepository, logger),
	))

	r.Delete("/api/v1/users/groups/delete", auth(
		rbac.NewPermission("/api/v1/users/groups/delete", "delete_group", "DELETE"),
		api.DeleteGroup(usersRepository, logger),
	))

	r.Post("/api/v1/users/groups/add_user", auth(
		rbac.NewPermission("/api/v1/users/groups/add_user", "add_group_user", "POST"),
		api.AddUserToGroup(usersRepository, logger),
	))

	r.Delete("/api/v1/users/groups/delete_user", auth(
		rbac.NewPermission("/api/v1/users/groups/delete_user", "delete_group_user", "DELETE"),
		api.DeleteUserFromGroup(usersRepository, logger),
	))

	api.RbacRoutes(r, auth, permissionRegistry, usersRepository, rbacRepository, logger)
	api.CampaignRoutes(r, auth, campaignsRepository, logger)

	totalRedirectsPromMiddleware := utils.PrometheusMiddleware("totalRedirects", "TODO description")

	var dbLogger utils.DbLogger
	if appConfig.RedirectLogger.Mode == "sync" && appConfig.RedirectLogger.Storage == "postgres" {
		dbLogger = utils.NewSyncLogger(database)
	} else if appConfig.RedirectLogger.Storage == "redis" || appConfig.RedirectLogger.Storage == "" {
		dbLogger = utils.NewRMQLogger("shortly", "redirects", appConfig.RedirectLogger.Redis)
	} else {
		logger.Fatal("incorrect config params for redirect logger")
	}
	r.Get("/metrics", promhttp.Handler().(http.HandlerFunc))
	r.Get("/*", totalRedirectsPromMiddleware(api.Redirect(dbLogger, historyDB, urlCache, logger)))

	var srv *http.Server
	// server running
	go func() {
		srv = &http.Server{
			Addr:    fmt.Sprintf(":%v", serverPort),
			Handler: r,
		}
		logger.Printf("starting web server at port: %v, tls: %v\n", serverConfig.Port, appConfig.Server.UseTLS)
		if appConfig.Server.UseTLS {
			if err := srv.ListenAndServeTLS("./server.crt", "./server.key"); err != nil && err != http.ErrServerClosed {
				logger.Printf("server stop unexpectedly, cause: %+v\n", err)
			}
		} else {
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Printf("server stop unexpectedly, cause: %+v\n", err)
			}
		}
	}()

	shutdownCh := make(chan os.Signal)
	doneCh := make(chan struct{})

	signal.Notify(shutdownCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-shutdownCh
		if err := database.Close(); err != nil {
			logger.Printf("database close error, cause: %+v", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			logger.Printf("server shutdown error, cause: %+v", err)
		}

		billingDataStorage.Close()

		doneCh <- struct{}{}
	}()

	<-doneCh

	logger.Println("server exit normally")
}
