package api

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"shortly/cache"
	"shortly/db"
	"shortly/utils"
)

func GetURLList(database *sql.DB, logger *log.Logger) {

	http.HandleFunc("/api/v1/urls", func(w http.ResponseWriter, r *http.Request) {

		// TODO - rate limiting
		rows, err := db.GetAllUrls(database)
		if err != nil {
			logger.Println(err)
			// TODO - логгирование асинхронное
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte("<!DOCTYPE html><html><body><ul>Urls:\n"))

		for _, r := range rows {
			_, _ = w.Write([]byte(fmt.Sprintf("<li>%s - %s</li>\n", r.Short, r.Long)))
		}
		_, _ = w.Write([]byte("</ul></body></html>\n"))
	})

}

func CreateShortURL(db *sql.DB, urlCache cache.UrlCache, logger *log.Logger) {

	http.HandleFunc("/api/v1/urls/create", func(w http.ResponseWriter, r *http.Request) {

		// TODO - rate limiting
		if r.Method != "POST" {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		urlArg := r.URL.Query()["url"]
		if len(urlArg) != 1 {
			http.Error(w, "invalid number of query values for parameter <url>, must be 1", http.StatusBadRequest)
			return
		}

		fullURL := urlArg[0]

		validFullURL, err := url.Parse(fullURL)
		if err != nil {
			http.Error(w, "url has incorrect format", http.StatusBadRequest)
			return
		}

		shortURL := utils.RandomString(5)
		_, err = db.Exec("INSERT INTO urls (short_url, full_url) VALUES ($1, $2)", shortURL, validFullURL.String())
		if err != nil {
			logger.Println(err)
			// TODO - логгирование асинхронное
			http.Error(w, "internal error", http.StatusInternalServerError)
		} else {
			urlCache.Store(shortURL, validFullURL.String())
			// TODO - определять хост
			_, _ = w.Write([]byte("http://localhost:5000/" + shortURL))
		}
	})

}

func RemoveShortURL(db *sql.DB, urlCache cache.UrlCache, logger *log.Logger) {

	http.HandleFunc("/api/v1/urls/remove", func(w http.ResponseWriter, r *http.Request) {

		// TODO - rate limiting
		if r.Method != "DELETE" {
			http.Error(w, "method not allowed", http.StatusBadRequest)
			return
		}

		urlArg := r.URL.Query()["url"]
		if len(urlArg) != 1 {
			http.Error(w, "invalid number of query values for parameter <url>, must be 1", http.StatusBadRequest)
			return
		}

		shortURL := urlArg[0]
		_, err := db.Exec("DELETE FROM urls WHERE short_url = $1", shortURL)
		if err != nil {
			logger.Println(err)
			// TODO - логгирование асинхронное
			http.Error(w, "internal error", http.StatusInternalServerError)
		} else {
			urlCache.Delete(shortURL)
			_, _ = w.Write([]byte("removed"))
		}
	})
}

func RedirectToFullURL(db *sql.DB, urlCache cache.UrlCache, logger *log.Logger) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		shortURL := strings.TrimPrefix(r.URL.Path, "/")

		if cacheURLValue, ok := urlCache.Load(shortURL); ok {

			fullURL, ok := cacheURLValue.(string)
			if !ok {
				http.Error(w, "url is not a string", http.StatusBadRequest)
				return
			}

			if !(strings.HasPrefix(fullURL, "http") || strings.HasPrefix(fullURL, "https")) {
				fullURL = "https://" + fullURL
			}

			validURL, err := url.Parse(fullURL)
			if err != nil {
				http.Error(w, "url has incorrect format", http.StatusBadRequest)
				return
			}

			http.Redirect(w, r, validURL.String(), http.StatusSeeOther)
		} else {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("not found"))
		}
	})
}
