package urls

import (
	"log"
	"database/sql"
)

type IUrlsRepository interface {
	GetAllUrls() ([]UrlPair, error)
	CreateUrl(short, long string) error
	GetUserUrls(userID int64) ([]UrlPair, error)
	GetUserUrlsCount(userID int64) (int, error)
}

type UrlsRepository struct {
	DB     *sql.DB
	Logger *log.Logger
}

func (repo *UrlsRepository) GetAllUrls() ([]UrlPair, error) {

	rows, err := repo.DB.Query("select short_url, full_url from urls")
	if err != nil {
		return nil, err
	}

	var urls []UrlPair

	for rows.Next() {
		var shortURL, fullURL string
		err := rows.Scan(&shortURL, &fullURL)
		if err != nil {
			return nil, err
		}
		urls = append(urls, UrlPair{Short: shortURL, Long: fullURL})
	}

	return urls, nil
}

func (repo *UrlsRepository) CreateUrl(short, long string) error {
	_, err := repo.DB.Exec("INSERT INTO urls (short_url, full_url) VALUES ($1, $2)", short, long)
	return err
}

func (repo *UrlsRepository) GetUserUrls(userID int64) ([]UrlPair, error) {

	rows, err := repo.DB.Query("select short_url, full_url from urls where user_id = $1", userID)
	if err != nil {
		return nil, err
	}

	var urls []UrlPair

	for rows.Next() {
		var shortURL, fullURL string
		err := rows.Scan(&shortURL, &fullURL)
		if err != nil {
			return nil, err
		}
		urls = append(urls, UrlPair{Short: shortURL, Long: fullURL})
	}

	return urls, nil
}

func (repo *UrlsRepository) GetUserUrlsCount(userID int64) (int, error) {

	var count int
	err := repo.DB.QueryRow("select count(*) from urls where user_id = $1", userID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}