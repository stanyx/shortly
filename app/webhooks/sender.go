package webhooks

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	bolt "go.etcd.io/bbolt"
)

// Sender ...
type Sender struct {
	Cache  *bolt.DB
	Logger *log.Logger
}

// Send ...
func (s *Sender) Send(url string, payload interface{}) {

	body, err := json.Marshal(payload)
	if err != nil {
		s.Logger.Println(err)
		return
	}

	r, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		s.Logger.Println(err)
		return
	}

	cl := &http.Client{Timeout: time.Second * 10}

	resp, err := cl.Do(r)
	if err != nil {
		s.Logger.Println(err)
		return
	}

	defer resp.Body.Close()

	s.Logger.Println("SEND webhook - initiated", r.URL, payload)

	if !(resp.StatusCode >= 200 && resp.StatusCode < 300) {
		s.Logger.Println("SEND webhook - failed, status_code: ", resp.StatusCode)
		return
	}
}

var DefaultSender = &Sender{}

// Send ...
func Send(hookName string) func(int64, interface{}) {
	return func(accountID int64, payload interface{}) {
		go func() {
			_ = DefaultSender.Cache.View(func(tx *bolt.Tx) error {
				b := tx.Bucket([]byte("webhooks"))
				hookUrl := b.Get([]byte(strconv.Itoa(int(accountID)) + ":" + hookName))
				if len(hookUrl) > 0 {
					DefaultSender.Send(string(hookUrl), payload)
				}
				return nil
			})
		}()
	}
}
