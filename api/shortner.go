package api

import (
	"log"
	"net/http"
	"net/url"
	"strings"

	"shortly/cache"
	"shortly/app/data"
)


func Redirect(historyDB *data.HistoryDB, urlCache cache.UrlCache, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		shortURL := strings.TrimPrefix(r.URL.Path, "/")

		if shortURL == "/" || shortURL == "" {
			http.ServeFile(w, r, "./static/index.html")
			return
		}

		if cacheURLValue, ok := urlCache.Load(shortURL); ok {

			fullURL, ok := cacheURLValue.(string)
			if !ok {
				apiError(w, "url is not a string", http.StatusBadRequest)
				return
			}

			if !(strings.HasPrefix(fullURL, "http") || strings.HasPrefix(fullURL, "https")) {
				fullURL = "https://" + fullURL
			}

			validURL, err := url.Parse(fullURL)
			if err != nil {
				apiError(w, "url has incorrect format", http.StatusBadRequest)
				return
			}

			if err := historyDB.Insert(shortURL, r); err != nil {
				apiError(w, "internal server error", http.StatusInternalServerError)
				return
			}

			// TODO - persistent logging to postgres with rmq

			http.Redirect(w, r, validURL.String(), http.StatusSeeOther)
		} else {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("not found"))
		}
	})
}


