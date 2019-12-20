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
	"sync"
	"syscall"
	"time"

	"shortly/config"
	"shortly/handlers"
	"shortly/server"
	"shortly/storage"

	_ "github.com/lib/pq"
)

var (
	db    *sql.DB
	cache = sync.Map{}
)

func LoadCacheFromDatabase() error {

	rows, err := db.Query("select short_url, full_url from urls")
	if err != nil {
		return err
	}

	for rows.Next() {
		var shortURL, fullURL string
		err := rows.Scan(&shortURL, &fullURL)
		if err != nil {
			return err
		}
		cache.Store(shortURL, fullURL)
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

	db, err = storage.StartDB(connString)
	if err != nil {
		logger.Fatal(err)
	}

	if err := LoadCacheFromDatabase(); err != nil {
		logger.Fatal(err)
	}

	handlers.GetURLList(db, cache, logger)
	handlers.CreateShortURL(db, cache, logger)
	handlers.RemoveShortURL(db, cache, logger)
	handlers.RedirectToFullURL(db, cache, logger)

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
		if err := db.Close(); err != nil {
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
