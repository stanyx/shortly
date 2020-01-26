package api

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strings"

	"shortly/app/data"
	"shortly/cache"
	"shortly/utils"

	"shortly/app/links"
)

type LinkRedirect struct {
	ShortUrl string
	LongUrl  string
	Headers  http.Header
}

func Redirect(repo links.ILinksRepository, redirectLogger utils.DbLogger, historyDB *data.HistoryDB, urlCache cache.UrlCache, logger *log.Logger) http.HandlerFunc {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		shortURL := strings.TrimPrefix(r.URL.Path, "/")

		scheme := r.URL.Scheme
		if scheme == "" {
			scheme = "http://"
		}

		if shortURL == "/" || shortURL == "" {
			http.Redirect(w, r, scheme+r.Host+"/static/index.html", http.StatusPermanentRedirect)
			return
		}

		if shortURL == "admin" {
			http.Redirect(w, r, scheme+r.Host+"/static/admin/index.html", http.StatusPermanentRedirect)
			return
		}

		cacheURLValue, ok := urlCache.Load(shortURL)

		var longURL string

		if !ok {
			longURL, _ = repo.UnshortenURL(shortURL)
			if longURL == "" {
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte("not found"))
				return
			}
		}

		longURL, ok = cacheURLValue.(string)
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
