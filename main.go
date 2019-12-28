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

	"github.com/kr/pretty"

	"shortly/api"
	"shortly/cache"
	"shortly/config"
	"shortly/db"
	"shortly/server"
	"shortly/storage"

	"shortly/app/billing"

	_ "github.com/lib/pq"
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

	logger := log.New(os.Stdout, "", log.LstdFlags)

	serverConfig := server.ParseServerOptions()
	appConfig, err := config.ReadConfig(serverConfig.ConfigPath)
	if err != nil {
		logger.Fatal(err)
	}

	dbConfig := appConfig.Database

	connString := fmt.Sprintf("host=%v port=%v user=%v password=%v dbname=%v sslmode=%v",
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.User,
		dbConfig.Password,
		dbConfig.Database,
		dbConfig.SSLMode,
	)

	database, err := storage.StartDB(connString)
	if err != nil {
		logger.Fatal(err)
	}

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
	}

	err = LoadCacheFromDatabase(database, urlCache)
	if err != nil {
		logger.Fatal(err)
	}

	// route declaration
	api.GetURLList(database, logger)
	api.CreateShortURL(database, urlCache, logger)
	api.RemoveShortURL(database, urlCache, logger)
	api.RedirectToFullURL(database, urlCache, logger)

	billingRepository := &billing.BillingRepository{DB: database}
	api.ListBillingPlans(billingRepository, logger)
	api.ApplyBillingPlan(billingRepository, logger)

	// server running
	srv := http.Server{Addr: fmt.Sprintf(":%v", serverConfig.Port)}

	go func() {
		logger.Printf("starting web server at port: %v\n", serverConfig.Port)
		if err := srv.ListenAndServe(); err != nil {
			logger.Fatalf("server stop unexpectedly, cause: %+v", err)
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
		doneCh <- struct{}{}
	}()

	<-doneCh

	logger.Println("server exit normally")
}
