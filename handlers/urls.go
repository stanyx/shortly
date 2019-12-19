package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"shortly/utils"
)

func GetURLList(db *sql.DB, cache sync.Map, logger *log.Logger) {

	http.HandleFunc("/v1/urls", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte("<!DOCTYPE html><html><body><ul>Urls:\n"))
		cache.Range(func(shortUrl interface{}, fullUrl interface{}) bool {
			w.Write([]byte(fmt.Sprintf("<li>%s - %s</li>\n", shortUrl.(string), fullUrl.(string))))
			return true
		})
		w.Write([]byte("</ul></body></html>\n"))
	})

}

func CreateShortURL(db *sql.DB, cache sync.Map, logger *log.Logger) {

	http.HandleFunc("/v1/urls/create", func(w http.ResponseWriter, r *http.Request) {

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
			cache.Store(shortURL, validFullURL.String())
			// TODO - определять хост
			w.Write([]byte("http://localhost:5000/" + shortURL))
		}
	})

}

func RemoveShortURL(db *sql.DB, cache sync.Map, logger *log.Logger) {

	http.HandleFunc("/v1/urls/remove", func(w http.ResponseWriter, r *http.Request) {
		urlArg := r.URL.Query()["url"]
		if len(urlArg) != 1 {
			http.Error(w, "invalid number of query values for parameter <url>, must be 1", http.StatusBadRequest)
			return
		}
		shortUrl := urlArg[0]
		_, err := db.Exec("DELETE FROM urls WHERE short_url = $1", shortUrl)
		if err != nil {
			logger.Println(err)
			// TODO - логгирование асинхронное
			http.Error(w, "internal error", http.StatusInternalServerError)
		} else {
			cache.Delete(shortUrl)
			w.Write([]byte("removed"))
		}
	})
}

func RedirectToFullURL(db *sql.DB, cache sync.Map, logger *log.Logger) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		shortURL := strings.TrimPrefix(r.URL.Path, "/")

		log.Println("get path", shortURL)
		if cacheURLValue, ok := cache.Load(shortURL); ok {

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
			w.Write([]byte("not found"))
		}
	})
}
