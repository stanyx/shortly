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
	"syscall"
	"time"

	_ "github.com/lib/pq"
	bolt "go.etcd.io/bbolt"
	"github.com/kr/pretty"

	"shortly/db"
	"shortly/api"
	"shortly/cache"
	"shortly/config"
	"shortly/server"
	"shortly/storage"

	"shortly/app/billing"
	"shortly/app/users"
	"shortly/app/urls"
)

func LoadCacheFromDatabase(database *sql.DB, urlCache cache.UrlCache) error {

	rows, err := db.GetAllUrls(database)
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

	err = LoadCacheFromDatabase(database, urlCache)
	if err != nil {
		logger.Fatal(err)
	}

	// ** route declaration

	// public api

	urlsRepository := &urls.UrlsRepository{DB: database, Logger: logger}

	http.Handle("/api/v1/urls", api.GetURLList(urlsRepository, logger))
	http.Handle("/api/v1/urls/create", api.CreateShortURL(urlsRepository, urlCache, logger))

	api.Redirect(urlCache, logger)

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

	auth := api.AuthMiddleware(appConfig.Auth)
	//   billing api
	billingRepository := &billing.BillingRepository{DB: database}
	billingLimiter := &billing.BillingLimiter{
		Repo:   billingRepository, 
		DB:     billingDataStorage,
		UrlDB:  database,
		Logger: logger,
	}
	if err := billingLimiter.LoadData(); err != nil {
		logger.Fatal(err)
	}
	urlBillingLimit := api.BillingLimitMiddleware("url_limit", billingLimiter, logger)

	http.Handle("/api/v1/billing/plans",     api.ListBillingPlans(billingRepository, logger))
	http.Handle("/api/v1/billing/apply",     auth(api.ApplyBillingPlan(billingRepository, billingLimiter, logger)))

	http.Handle("/api/v1/users/urls",        auth(api.GetUserURLList(urlsRepository, logger)))
	http.Handle("/api/v1/users/urls/create", auth(urlBillingLimit(api.CreateUserShortURL(database, urlCache, billingLimiter, logger))))
	http.Handle("/api/v1/users/urls/delete", auth(api.RemoveUserShortURL(database, urlCache, logger)))

	// users api
	usersRepository := &users.UsersRepository{DB: database}
	api.RegisterUser(usersRepository, logger)
	api.LoginUser(usersRepository, logger, appConfig.Auth)

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
