package api

import (
	"database/sql"
	"log"
	"net/http"
	"net/url"
	"strings"

	"shortly/cache"
	"shortly/utils"

	"shortly/app/billing"
	"shortly/app/urls"
)

// Public API

// TODO - auto expired urls

type UrlResponse struct {
	Short string `json:"short"`
	Long  string `json:"long"`
}

func GetURLList(repo urls.IUrlsRepository, logger *log.Logger) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		rows, err := repo.GetAllUrls()
		if err != nil {
			logError(logger, err)
			apiError(w, "internal error", http.StatusInternalServerError)
			return
		}

		var urls []UrlResponse
		for _, r := range rows {
			urls = append(urls, UrlResponse{
				Short: r.Short,
				Long:  r.Long,
			})
		}
		
		response(w, urls, http.StatusOK)
	})

}

func GetUserURLList(repo urls.IUrlsRepository, logger *log.Logger) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		userID := r.Context().Value("user").(*JWTClaims).UserID

		rows, err := repo.GetUserUrls(userID)
		if err != nil {
			logError(logger, err)
			apiError(w, "internal error", http.StatusInternalServerError)
			return
		}

		var urls []UrlResponse
		for _, r := range rows {
			urls = append(urls, UrlResponse{
				Short: r.Short,
				Long:  r.Long,
			})
		}
		
		response(w, urls, http.StatusOK)
	})

}

func CreateShortURL(repo urls.IUrlsRepository, urlCache cache.UrlCache, logger *log.Logger) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Method != "POST" {
			apiError(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		urlArg := r.URL.Query()["url"]
		if len(urlArg) != 1 {
			apiError(w, "invalid number of query values for parameter <url>, must be 1", http.StatusBadRequest)
			return
		}

		fullURL := urlArg[0]

		validFullURL, err := url.Parse(fullURL)
		if err != nil {
			apiError(w, "url has incorrect format", http.StatusBadRequest)
			return
		}

		shortURL := utils.RandomString(5)
		err = repo.CreateUrl(shortURL, validFullURL.String())
		if err != nil {
			logError(logger, err)
			apiError(w, "internal error", http.StatusInternalServerError)
		} else {
			urlCache.Store(shortURL, validFullURL.String())
			response(w, &UrlResponse{Short: r.Host + "/" + shortURL, Long: fullURL}, http.StatusOK)
		}
	})

}

func Redirect(urlCache cache.UrlCache, logger *log.Logger) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

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

			http.Redirect(w, r, validURL.String(), http.StatusSeeOther)
		} else {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("not found"))
		}
	})
}

func CreateUserShortURL(db *sql.DB, urlCache cache.UrlCache, billingLimiter *billing.BillingLimiter, logger *log.Logger) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		userID := r.Context().Value("user").(*JWTClaims).UserID

		if r.Method != "POST" {
			apiError(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		urlArg := r.URL.Query()["url"]
		if len(urlArg) != 1 {
			apiError(w, "invalid number of query values for parameter <url>, must be 1", http.StatusBadRequest)
			return
		}

		fullURL := urlArg[0]

		validFullURL, err := url.Parse(fullURL)
		if err != nil {
			apiError(w, "url has incorrect format", http.StatusBadRequest)
			return
		}

		shortURL := utils.RandomString(5)
		_, err = db.Exec("INSERT INTO urls (short_url, full_url, user_id) VALUES ($1, $2, $3)", 
			shortURL, validFullURL.String(), userID)
		if err != nil {
			logError(logger, err)
			apiError(w, "(create url) - internal error", http.StatusInternalServerError)
		} else {
			urlCache.Store(shortURL, validFullURL.String())

			if err := billingLimiter.Reduce("url_limit", userID); err != nil {
				logError(logger, err)
				apiError(w, "(create url) - internal error", http.StatusInternalServerError)
				return
			}

			response(w, &UrlResponse{Short: r.Host + "/" + shortURL, Long: fullURL}, http.StatusOK)
		}
	})

}

func RemoveUserShortURL(db *sql.DB, urlCache cache.UrlCache, logger *log.Logger) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		userID := r.Context().Value("user").(*JWTClaims).UserID

		if r.Method != "DELETE" {
			apiError(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		urlArg := r.URL.Query()["url"]
		if len(urlArg) != 1 {
			apiError(w, "invalid number of query values for parameter <url>, must be 1", http.StatusBadRequest)
			return
		}

		shortURL := urlArg[0]
		_, err := db.Exec("DELETE FROM urls WHERE short_url = $1 AND user_id = $2", shortURL, userID)
		if err != nil {
			logError(logger, err)
			apiError(w, "internal error", http.StatusInternalServerError)
		} else {
			urlCache.Delete(shortURL)
			ok(w)
		}
	})
}