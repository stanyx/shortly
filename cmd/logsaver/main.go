// Queue consumer for saving redirects
package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/adjust/rmq"
	_ "github.com/lib/pq"

	"shortly/config"
	"shortly/storage"
)

type linkRedirect struct {
	ShortUrl string
	LongUrl  string
	Headers  map[string]interface{}
	IPAddr   string
	Country  string
	Referer  string
}

type Consumer struct {
	db *sql.DB
}

func NewConsumer(db *sql.DB) *Consumer {
	return &Consumer{
		db: db,
	}
}

func (consumer *Consumer) Consume(delivery rmq.Delivery) {
	// handle delivery and call Ack() or Reject() on it

	var msg linkRedirect
	err := json.Unmarshal([]byte(delivery.Payload()), &msg)
	if err != nil {
		log.Println("error on unmarshal delivery", err)
		delivery.Reject()
		return
	}

	headers, err := json.Marshal(msg.Headers)
	if err != nil {
		log.Println("error on marshal headers", err)
		delivery.Reject()
		return
	}

	_, err = consumer.db.Exec(`
		insert into redirect_log(short_url, long_url, headers, country, ip_addr, referer, timestamp) 
		values ($1, $2, $3, $4, $5, $6, now())
	`,
		msg.ShortUrl,
		msg.LongUrl,
		string(headers),
		msg.Country,
		msg.IPAddr,
		msg.Referer,
	)
	if err != nil {
		log.Println("error on save", err)
		delivery.Reject()
	}

	delivery.Ack()
}

func main() {

	var configPath string
	flag.StringVar(&configPath, "config", "./config/config.yaml", "path to config file")
	flag.Parse()

	appConfig, err := config.ReadConfig(configPath)
	if err != nil {
		log.Fatal(err)
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

	db, err := storage.StartDB(dbConnString)
	if err != nil {
		log.Fatal(err)
	}

	queueConn := rmq.OpenConnection("shortly", "tcp", ":6379", 1)
	taskQueue := queueConn.OpenQueue("redirects")

	log.Println("start consuming")
	taskQueue.StartConsuming(10, time.Second)

	consumer := NewConsumer(db)
	taskQueue.AddConsumer("logsaver", consumer)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT)

	done := make(chan struct{}, 1)
	go func() {
		<-ch
		if err := db.Close(); err != nil {
			log.Println("error closing database", err)
		}
		taskQueue.StopConsuming()
		close(done)
	}()

	<-done
	log.Println("stop consuming")
}
