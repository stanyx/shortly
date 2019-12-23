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