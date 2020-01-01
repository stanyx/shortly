package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	bolt "go.etcd.io/bbolt"
	"github.com/kr/pretty"

	"shortly/api"
	"shortly/cache"
	"shortly/config"
	"shortly/server"
	"shortly/storage"

	"shortly/app/billing"
	"shortly/app/rbac"
	"shortly/app/users" //TODO - rename to accounts
	"shortly/app/urls"
	"shortly/app/data"
)

func LoadCacheFromDatabase(repo *urls.UrlsRepository, urlCache cache.UrlCache) error {

	rows, err := repo.GetAllUrls()
	if err != nil {
		return err
	}

	for _, r := range rows {
		urlCache.Store(r.Short, r.Long)
	}

	return nil
}

func main() {

	rand.Seed(time.Now().UnixNano())

	logger := log.New(os.Stdout, "", log.LstdFlags | log.Llongfile)

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

	billingDataStorage, err := bolt.Open(appConfig.Billing.Dir + "/billing.db", 0666, nil)
	if err != nil {
		logger.Fatal(err)
	}

	linksStorage, err := bolt.Open(appConfig.LinkDB.Dir + "/links.db", 0666, nil)
	if err != nil {
		logger.Fatal(err)
	}

	urlsRepository := &urls.UrlsRepository{DB: database, Logger: logger}

	billingRepository := &billing.BillingRepository{DB: database}
	billingLimiter := &billing.BillingLimiter{
		Repo:    billingRepository, 
		DB:      billingDataStorage,
		UrlRepo: urlsRepository,
		Logger:  logger,
	}
	if err := billingLimiter.LoadData(); err != nil {
		logger.Fatal(err)
	}

	urlBillingLimit := api.BillingLimitMiddleware("url_limit", billingLimiter, logger)

	historyDB := &data.HistoryDB{DB: linksStorage, Limiter: billingLimiter}

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
		urlDataStorage, err := bolt.Open(cacheConfig.BoltDB.Dir + "/urls.db", 0666, nil)
		if err != nil {
			logger.Fatal(err)
		}
		urlCache, err = cache.NewBoltDBCache(urlDataStorage, logger)
		if err != nil {
			logger.Fatal(err)
		}
	}

	err = LoadCacheFromDatabase(urlsRepository, urlCache)
	if err != nil {
		logger.Fatal(err)
	}

	// ** route declaration

	// public api

	fs := http.FileServer(http.Dir("static"))
    http.Handle("/static", http.StripPrefix("/static", fs))

	http.Handle("/", api.Redirect(historyDB, urlCache, logger))

	http.Handle("/api/v1/urls", api.GetURLList(urlsRepository, logger))
	http.Handle("/api/v1/urls/create", api.CreateShortURL(urlsRepository, urlCache, logger))

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
			return fmt.Errorf("create buket error, cause: %+v", err)
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

	http.Handle("/api/v1/billing/plans", api.ListBillingPlans(billingRepository, logger))

	http.Handle("/api/v1/billing/apply", auth(
		rbac.NewPermission("/api/v1/billing/apply", "apply_billingplan", "POST"), 
		api.ApplyBillingPlan(billingRepository, billingLimiter, appConfig.Billing.Payment, logger),
	))

	// TODO remove to routes.go
	// links api
	http.Handle("/api/v1/users/urls", auth(
		rbac.NewPermission("/api/v1/users/urls", "read_urls", "GET"), 
		api.GetUserURLList(urlsRepository, logger),
	))

	http.Handle("/api/v1/users/urls/clicks", auth(
		rbac.NewPermission("/api/v1/users/urls/clicks", "get_links_clicks", "GET"), 
		api.GetClicksData(historyDB, logger),
	))

	http.Handle("/api/v1/users/urls/add_group", auth(
		rbac.NewPermission("/api/v1/users/urls/add_group", "add_url_to_group", "POST"), 
		api.AddUrlToGroup(urlsRepository, logger),
	))

	http.Handle("/api/v1/users/urls/delete_group", auth(
		rbac.NewPermission("/api/v1/users/urls/delete_group", "delete_url_to_group", "DELETE"), 
		api.DeleteUrlFromGroup(urlsRepository, logger),
	))

	// account api
	usersRepository := &users.UsersRepository{DB: database}

	http.Handle("/api/v1/registration", api.RegisterAccount(usersRepository, logger))
	http.Handle("/api/v1/login", api.Login(usersRepository, logger, appConfig.Auth))

	http.Handle("/api/v1/users/urls/create", auth(
		rbac.NewPermission("/api/v1/users/urls/create", "create_url", "POST"), 
		urlBillingLimit(api.CreateUserShortURL(historyDB, database, urlCache, billingLimiter, logger)),
	))
		
	http.Handle("/api/v1/users/urls/delete", auth(
		rbac.NewPermission("/api/v1/users/urls/delete", "delete_url", "DELETE"), 
		api.RemoveUserShortURL(database, urlCache, logger),
	))

	http.Handle("/api/v1/users/groups/create", auth(
		rbac.NewPermission("/api/v1/users/groups/create", "create_group", "POST"), 
		api.AddGroup(usersRepository, logger),
	))

	http.Handle("/api/v1/users/groups/delete", auth(
		rbac.NewPermission("/api/v1/users/groups/delete", "delete_group", "DELETE"), 
		api.DeleteGroup(usersRepository, logger),
	))

	http.Handle("/api/v1/users/groups/add_user", auth(
		rbac.NewPermission("/api/v1/users/groups/add_user", "add_group_user", "POST"), 
		api.AddUserToGroup(usersRepository, logger),
	))

	http.Handle("/api/v1/users/groups/delete_user", auth(
		rbac.NewPermission("/api/v1/users/groups/delete_user", "delete_group_user", "DELETE"), 
		api.DeleteUserFromGroup(usersRepository, logger),
	))

	api.RbacRoutes(auth, permissionRegistry, usersRepository, rbacRepository, logger)

	serverPort := os.Getenv("PORT")
	if serverPort == "" {
		serverPort = fmt.Sprintf("%v", serverConfig.Port)
	}

	var srv *http.Server
	// server running
	go func() {
		logger.Printf("starting web server at port: %v, tls: %v\n", serverConfig.Port, appConfig.Server.UseTLS)
		if appConfig.Server.UseTLS {
			srv = &http.Server{Addr: fmt.Sprintf(":%v", serverPort)}
			if err := srv.ListenAndServeTLS("./server.crt", "./server.key"); err != nil {
				logger.Fatalf("server stop unexpectedly, cause: %+v", err)
			}
		} else {
			srv = &http.Server{Addr: fmt.Sprintf(":%v", serverPort)}
			if err := srv.ListenAndServe(); err != nil {
				logger.Fatalf("server stop unexpectedly, cause: %+v", err)
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
