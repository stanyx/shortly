package webhooks

import (
	"database/sql"
	"log"

	"github.com/lib/pq"
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
	Logger *log.Logger
}

func (r *WebhooksRepository) GetWebhooks(accountID int64) ([]Webhook, error) {
	var list []Webhook

	rows, err := r.DB.Query("select id, name, description, events, url, active from webhooks where account_id = $1")
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
		insert into "webhooks" (name, description, events, url, account_id) VALUES ( $1, $2, $3, $4 )
		returning id
	`, m.Name, m.Description, pq.Array(m.Events), m.URL, accountID).Scan(&rowID)

	if err != nil {
		return 0, err
	}

	return rowID, nil
}

func (r *WebhooksRepository) UpdateWebhook(accountID int64, m Webhook) error {
	_, err := r.DB.Exec(`
		update "webhooks" set name = $1, description = $2, events = $3, url = $4 where id = $5 and account_id = $6
	`, m.Name, m.Description, pq.Array(m.Events), m.URL, m.ID, accountID)
	return err
}

func (r *WebhooksRepository) DeleteWebhook(accountID int64, id int64) error {
	_, err := r.DB.Exec(`
		delete from "webhooks" where id = $1 and account_id = $2
	`, id, accountID)
	return err
}

func (r *WebhooksRepository) EnableWebhook(accountID int64, id int64) error {
	_, err := r.DB.Exec(`
		update "webhooks" set active = true where id = $1 and account_id = $2
	`, id, accountID)
	return err
}

func (r *WebhooksRepository) DisableWebhook(accountID int64, id int64) error {
	_, err := r.DB.Exec(`
		update "webhooks" set active = false where id = $1 and account_id = $2
	`, id, accountID)
	return err
}
