package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"math/rand"
	"strings"
	"sync"
	"time"
	_ "github.com/lib/pq"
	"database/sql"
)

var db *sql.DB

func LoadCacheFromDatabase() {

	rows, err := db.Query("select short_url, full_url from urls")
	if err != nil {
		log.Fatal(err)
	}

	for rows.Next(){
		var shortUrl, fullUrl string
		err := rows.Scan(&shortUrl, &fullUrl)
		if err != nil{
			log.Fatal(err)
		}
		cache.Store(shortUrl, fullUrl)
	}
}

func RandomString(n int) string {
	var letter = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	b := make([]rune, n)
	for i := range b {
		b[i] = letter[rand.Intn(len(letter))]
	}
	return string(b)
}

var cache = sync.Map{}

func main() {

	rand.Seed(time.Now().UnixNano())

	flag.Parse()
	
	// TODO - read from config
	var err error
	connString := "host=localhost port=5432 user=shortly_user password=1 dbname=shortly sslmode=disable"

	db, err = sql.Open("postgres", connString)
	if err != nil {
		log.Fatal(err)
	}

	LoadCacheFromDatabase()

	http.HandleFunc("/v1/urls", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte("<!DOCTYPE html><html><body><ul>Urls:\n"))
		cache.Range(func(shortUrl interface{}, fullUrl interface{}) bool {
			w.Write([]byte(fmt.Sprintf("<li>%s - %s</li>\n", shortUrl.(string), fullUrl.(string))))
			return true
		})
		w.Write([]byte("</ul></body></html>\n"))
	})

	http.HandleFunc("/v1/urls/create", func(w http.ResponseWriter, r *http.Request) {

		urlArg := r.URL.Query()["url"]
		if len(urlArg) != 1 {
			http.Error(w, "invalid number of query values for parameter <url>, must be 1", http.StatusBadRequest)
			return
		}

		fullUrl := urlArg[0]

		validFullUrl, err := url.Parse(fullUrl)
		if err != nil {
			http.Error(w, "url has incorrect format", http.StatusBadRequest)
			return
		}

		shortUrl := RandomString(5)
		_, err = db.Exec("INSERT INTO urls (short_url, full_url) VALUES ($1, $2)", shortUrl, validFullUrl.String())
		if err != nil {
			log.Println(err)
			// TODO - логгирование асинхронное
			http.Error(w, "internal error", http.StatusInternalServerError)
		} else {
			cache.Store(shortUrl, validFullUrl.String())
			// TODO - определять хост
			w.Write([]byte("http://localhost:5000/" + shortUrl))
		}
	})

	http.HandleFunc("/v1/urls/remove", func(w http.ResponseWriter, r *http.Request) {
		urlArg := r.URL.Query()["url"]
		if len(urlArg) != 1 {
			http.Error(w, "invalid number of query values for parameter <url>, must be 1", http.StatusBadRequest)
			return
		}
		shortUrl := urlArg[0]
		_, err := db.Exec("DELETE FROM urls WHERE short_url = $1", shortUrl)
		if err != nil {
			log.Println(err)
			// TODO - логгирование асинхронное
			http.Error(w, "internal error", http.StatusInternalServerError)
		} else {
			cache.Delete(shortUrl)
			w.Write([]byte("removed"))
		}
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println(cache)
		shortUrl := strings.TrimPrefix(r.URL.Path, "/")

		log.Println("get path", shortUrl)
		if cacheUrlValue, ok := cache.Load(shortUrl); ok {

			fullUrl, ok := cacheUrlValue.(string)
			if !ok {
				http.Error(w, "url is not a string", http.StatusBadRequest)
				return
			}

			if !(strings.HasPrefix(fullUrl, "http") || strings.HasPrefix(fullUrl, "https")) {
				fullUrl = "https://" + fullUrl
			}

			validUrl, err := url.Parse(fullUrl)
			if err != nil {
				http.Error(w, "url has incorrect format", http.StatusBadRequest)
				return
			}

			http.Redirect(w, r, validUrl.String(), http.StatusSeeOther)
		} else {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("not found"))
		}
	})

	http.ListenAndServe(":5000", nil)
}