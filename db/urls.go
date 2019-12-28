package db

import (
	"database/sql"
)

type UrlPair struct {
	Short string
	Long string
}

func GetAllUrls(db *sql.DB) ([]UrlPair, error) {

	rows, err := db.Query("select short_url, full_url from urls")
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

func GetUserUrls(db *sql.DB, userID int64) ([]UrlPair, error) {

	rows, err := db.Query("select short_url, full_url from urls where user_id = $1", userID)
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

func GetUserUrlsCount(db *sql.DB, userID int64) (int, error) {

	var count int
	err := db.QueryRow("select count(*) from urls where user_id = $1", userID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}