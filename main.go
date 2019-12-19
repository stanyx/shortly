package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"shortly/handlers"

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

	flag.Parse()

	// TODO - read from config
	var err error
	connString := "host=localhost port=5432 user=shortly_user password=1 dbname=shortly sslmode=disable"

	logger := log.New(os.Stdout, "", log.LstdFlags)

	db, err = sql.Open("postgres", connString)
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

	srv := http.Server{Addr: ":5000"}

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
