package api

import (
	"log"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"shortly/cache"
	"shortly/utils"
	"shortly/app/data"
)

type LinkRedirect struct {
	ShortUrl string
	LongUrl  string
	Headers  http.Header
}


func Redirect(redirectLogger utils.DbLogger, historyDB *data.HistoryDB, urlCache cache.UrlCache, logger *log.Logger) http.HandlerFunc {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		shortURL := strings.TrimPrefix(r.URL.Path, "/")

		if shortURL == "/" || shortURL == "" {
			http.ServeFile(w, r, "./static/index.html")
			return
		}

		cacheURLValue, ok := urlCache.Load(shortURL)

		if !ok {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("not found"))
			return
		}

		longURL, ok := cacheURLValue.(string)
		if !ok {
			apiError(w, "url is not a string", http.StatusBadRequest)
			return
		}

		if !(strings.HasPrefix(longURL, "http") || strings.HasPrefix(longURL, "https")) {
			longURL = "https://" + longURL
		}

		validURL, err := url.Parse(longURL)
		if err != nil {
			apiError(w, "url has incorrect format", http.StatusBadRequest)
			return
		}

		if err := historyDB.Insert(shortURL, r); err != nil {
			logError(logger, err)
			apiError(w, "internal server error", http.StatusInternalServerError)
			return
		}

		body, err := json.Marshal(&LinkRedirect{
			ShortUrl: shortURL,
			LongUrl:  validURL.String(),
			Headers:  r.Header,
		})
		if err != nil {
			logError(logger, err)
			apiError(w, "internal server error", http.StatusInternalServerError)
			return
		}

		if err := redirectLogger.Push(body); err != nil {
			logError(logger, err)
			apiError(w, "internal server error", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, validURL.String(), http.StatusSeeOther)

	})
}

