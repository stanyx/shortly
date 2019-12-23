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

	"shortly/cache"
	"shortly/config"
	"shortly/handlers"
	"shortly/server"
	"shortly/storage"
	"shortly/db"

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

	urlCache := cache.NewMemoryCache()
	err = LoadCacheFromDatabase(database, urlCache)
	if err != nil {
		logger.Fatal(err)
	}

	handlers.GetURLList(database, logger)
	handlers.CreateShortURL(database, urlCache, logger)
	handlers.RemoveShortURL(database, urlCache, logger)
	handlers.RedirectToFullURL(database, urlCache, logger)

	// запуск сервера
	srv := http.Server{Addr: fmt.Sprintf(":%v", serverConfig.Port)}

	go func() {
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
