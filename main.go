package main

import (
	"flag"
	"net/http"
	"strings"
	"database/sql"
	_ "github.com/lib/pq"
	"log"
	"math/rand"
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
		cache[shortUrl] = fullUrl
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

var cache = make(map[string]string)

func main() {
	flag.Parse()
	
	var err error
	connString := "host=localhost port=5432 user=shortly_user password=1 dbname=shortly sslmode=disable"

	db, err = sql.Open("postgres", connString)
	if err != nil {
		log.Fatal(err)
	}

	LoadCacheFromDatabase()

	http.HandleFunc("/v1/urls/create", func(w http.ResponseWriter, r *http.Request) {
		fullUrl := r.URL.Query()["url"][0]
		shortUrl := RandomString(5)
		cache[shortUrl] = fullUrl
		_, err := db.Exec("INSERT INTO urls (short_url, full_url) VALUES ($1, $2)", shortUrl, fullUrl)
		if err != nil {
			log.Println(err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		} else {
			w.Write([]byte("http://localhost:5000/" + shortUrl))
		}
	})

	http.HandleFunc("/v1/urls/remove", func(w http.ResponseWriter, r *http.Request) {
		_, err := db.Exec("DELETE FROM urls WHERE short_url = $1", r.URL.Query()["url"][0])
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
		} else {
			w.Write([]byte("removed"))
		}
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println(cache)
		shortUrl := strings.TrimPrefix(r.URL.Path, "/")

		log.Println("get path", shortUrl)
		if fullUrl, ok := cache[shortUrl]; ok {

			if !(strings.HasPrefix(fullUrl, "http") || strings.HasPrefix(fullUrl, "https")) {
				fullUrl = "https://" + fullUrl
			}

			http.Redirect(w, r, fullUrl, http.StatusSeeOther)
		} else {
			w.Write([]byte("not found"))
		}
	})

	http.ListenAndServe(":5000", nil)
}