package webhooks

import (
	"database/sql"
	"log"
	"strconv"

	"github.com/lib/pq"
	bolt "go.etcd.io/bbolt"
)

type Repository interface {
	GetWebhooks(accountID int64) ([]Webhook, error)
	GetWebhookByID(accountID int64, id int64) (*Webhook, error)
	CreateWebhook(accountID int64, m Webhook) (int64, error)
	UpdateWebhook(accountID int64, m Webhook) error
	DeleteWebhook(accountID int64, id int64) error
	EnableWebhook(accountID int64, id int64) error
	DisableWebhook(accountID int64, id int64) error
}

type WebhooksRepository struct {
	DB     *sql.DB
	Cache  *bolt.DB
	Logger *log.Logger
}

func (r *WebhooksRepository) InitCache() error {

	ws, err := r.GetWebhooks(0)
	if err != nil {
		return err
	}

	return r.Cache.Update(func(tx *bolt.Tx) error {
		for _, w := range ws {
			b := tx.Bucket([]byte("webhooks"))
			for _, event := range w.Events {
				if err := b.Put([]byte(strconv.Itoa(int(w.AccountID))+":"+event), []byte(w.URL)); err != nil {
					return err
				}
			}
		}
		return nil
	})

}

func (r *WebhooksRepository) GetWebhooks(accountID int64) ([]Webhook, error) {
	var list []Webhook

	query := "select id, name, description, events, url, active from webhooks"
	var queryArgs []interface{}

	if accountID > 0 {
		query += " where account_id = $1"
		queryArgs = append(queryArgs, accountID)
	}

	rows, err := r.DB.Query(query, queryArgs)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var m Webhook
		err := rows.Scan(&m.ID, &m.Name, &m.Description, pq.Array(&m.Events), &m.URL, &m.Active)
		if err != nil {
			return nil, err
		}
		list = append(list, m)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	defer rows.Close()

	return list, nil
}

func (r *WebhooksRepository) GetWebhookByID(accountID int64, id int64) (*Webhook, error) {

	rows := r.DB.QueryRow(
		"select id, name, description, events, url, active from webhooks where account_id = $1 and id = $2",
		accountID, id)

	var m Webhook
	err := rows.Scan(&m.ID, &m.Name, &m.Description, pq.Array(&m.Events), &m.URL, &m.Active)
	if err != nil {
		return nil, err
	}

	return &m, nil
}

func (r *WebhooksRepository) CreateWebhook(accountID int64, m Webhook) (int64, error) {
	var rowID int64

	err := r.DB.QueryRow(`
		insert into "webhooks" (name, description, events, url, account_id) VALUES ( $1, $2, $3, $4, $5 )
		returning id
	`, m.Name, m.Description, pq.Array(m.Events), m.URL, accountID).Scan(&rowID)

	if err != nil {
		return 0, err
	}

	err = r.Cache.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("webhooks"))
		for _, event := range m.Events {
			if err := b.Put([]byte(strconv.Itoa(int(accountID))+":"+event), []byte(m.URL)); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return 0, err
	}

	return rowID, nil
}

func (r *WebhooksRepository) UpdateWebhook(accountID int64, m Webhook) error {
	_, err := r.DB.Exec(`
		update "webhooks" set name = $1, description = $2, events = $3, url = $4 where id = $5 and account_id = $6
	`, m.Name, m.Description, pq.Array(m.Events), m.URL, m.ID, accountID)
	if err != nil {
		return err
	}

	err = r.Cache.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("webhooks"))
		for _, event := range m.Events {
			if err := b.Put([]byte(strconv.Itoa(int(accountID))+":"+event), []byte(m.URL)); err != nil {
				return err
			}
		}
		return nil
	})

	return err
}

func (r *WebhooksRepository) DeleteWebhook(accountID int64, id int64) error {
	_, err := r.DB.Exec(`
		delete from "webhooks" where id = $1 and account_id = $2
	`, id, accountID)

	if err != nil {
		return err
	}

	err = r.Cache.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("webhooks"))
		for _, event := range []string{"link__created", "link__deleted", "link__redirect"} {
			if err := b.Delete([]byte(strconv.Itoa(int(accountID)) + ":" + event)); err != nil {
				return err
			}
		}
		return nil
	})

	return err
}

func (r *WebhooksRepository) EnableWebhook(accountID int64, id int64) error {

	var events []string
	var webhookURL string

	err := r.DB.QueryRow(`
		update "webhooks" set active = true where id = $1 and account_id = $2
		returning events, url
	`, id, accountID).Scan(pq.Array(&events))

	if err != nil {
		return err
	}

	err = r.Cache.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("webhooks"))
		for _, event := range events {
			if err := b.Put([]byte(strconv.Itoa(int(accountID))+":"+event), []byte(webhookURL)); err != nil {
				return err
			}
		}
		return nil
	})

	return err
}

func (r *WebhooksRepository) DisableWebhook(accountID int64, id int64) error {

	var events []string
	err := r.DB.QueryRow(`
		update "webhooks" set active = false where id = $1 and account_id = $2
		returning events
	`, id, accountID).Scan(pq.Array(&events))

	if err != nil {
		return err
	}

	err = r.Cache.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("webhooks"))
		for _, event := range events {
			if err := b.Delete([]byte(strconv.Itoa(int(accountID)) + ":" + event)); err != nil {
				return err
			}
		}
		return nil
	})

	return err
}
