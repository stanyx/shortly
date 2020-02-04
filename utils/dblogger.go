package utils

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/adjust/rmq"

	"shortly/config"
)

type linkRedirect struct {
	ShortUrl string
	LongUrl  string
	Headers  map[string]interface{}
	IPAddr   string
	Country  string
}

// DbLogger ...
type DbLogger interface {
	Push([]byte) error
}

// RMQLogger ...
type RMQLogger struct {
	queue rmq.Queue
}

// NewRMQLogger ...
func NewRMQLogger(dbName string, queueName string, conf config.RedisConfig) *RMQLogger {
	conn := rmq.OpenConnection(dbName, "tcp", fmt.Sprintf("%v:%v", conf.Host, conf.Port), 1)
	queue := conn.OpenQueue(queueName)
	return &RMQLogger{
		queue: queue,
	}
}

func (l *RMQLogger) Push(body []byte) error {
	l.queue.Publish(string(body))
	return nil
}

// SyncLogger ...
type SyncLogger struct {
	db *sql.DB
}

// NewSyncLogger ...
func NewSyncLogger(db *sql.DB) *SyncLogger {
	return &SyncLogger{
		db: db,
	}
}

func (l *SyncLogger) Push(body []byte) error {

	var msg linkRedirect
	err := json.Unmarshal([]byte(body), &msg)
	if err != nil {
		return err
	}

	headers, err := json.Marshal(msg.Headers)
	if err != nil {
		return err
	}

	_, err = l.db.Exec(`
		insert into redirect_log(short_url, long_url, headers, country, ip_addr, timestamp) 
		values ($1, $2, $3, $4, $5, now())
	`,
		msg.ShortUrl,
		msg.LongUrl,
		string(headers),
		msg.Country,
		msg.IPAddr,
	)

	return err
}
