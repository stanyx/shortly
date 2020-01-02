package api

import (
	"database/sql"
	"log"
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"shortly/cache"
	"shortly/utils"

	"shortly/app/urls"
	"shortly/app/billing"
	"shortly/app/data"
)


// Public API

// TODO - auto expired urls

type UrlResponse struct {
	Short string `json:"short"`
	Long  string `json:"long"`
}

func GetURLList(repo urls.IUrlsRepository, logger *log.Logger) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// TODO - add search and filters functions
		// TODO - add pagination

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

		claims := r.Context().Value("user").(*JWTClaims)

		rows, err := repo.GetUserUrls(claims.AccountID, claims.UserID)
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

func CreateUserShortURL(historyDB *data.HistoryDB, db *sql.DB, urlCache cache.UrlCache, billingLimiter *billing.BillingLimiter, logger *log.Logger) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		claims := r.Context().Value("user").(*JWTClaims)
		accountID := claims.AccountID

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
		_, err = db.Exec("INSERT INTO urls (short_url, full_url, account_id) VALUES ($1, $2, $3)", 
			shortURL, validFullURL.String(), accountID)
		if err != nil {
			logError(logger, err)
			apiError(w, "(create url) - internal error", http.StatusInternalServerError)
			return
		}

		urlCache.Store(shortURL, validFullURL.String())

		if err := billingLimiter.Reduce("url_limit", accountID); err != nil {
			logError(logger, err)
			apiError(w, "(create url) - internal error", http.StatusInternalServerError)
			return
		}

		if err := historyDB.InsertDetail(shortURL, accountID); err != nil {
			logError(logger, err)
			apiError(w, "(create url) - internal error", http.StatusInternalServerError)
			return
		}

		response(w, &UrlResponse{Short: r.Host + "/" + shortURL, Long: fullURL}, http.StatusOK)
	})

}

func RemoveUserShortURL(db *sql.DB, urlCache cache.UrlCache, logger *log.Logger) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		claims := r.Context().Value("user").(*JWTClaims)
		accountID := claims.AccountID

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
		_, err := db.Exec("DELETE FROM urls WHERE short_url = $1 AND account_id = $2", shortURL, accountID)
		if err != nil {
			logError(logger, err)
			apiError(w, "internal error", http.StatusInternalServerError)
		} else {
			urlCache.Delete(shortURL)
			ok(w)
		}
	})
}

type AddUrlToGroupForm struct {
	GroupID int64 `json:"groupId"`
	UrlID  int64  `json:"urlId"`
}

func AddUrlToGroup(repo *urls.UrlsRepository, logger *log.Logger) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var form AddUrlToGroupForm

		if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
			logError(logger, err)
			apiError(w, "decode form error", http.StatusBadRequest)
			return
		}

		// TODO - check url id

		// TODO - check group by account_id

		if err := repo.AddUrlToGroup(form.GroupID, form.UrlID); err != nil {
			logError(logger, err)
			apiError(w, "decode form error", http.StatusBadRequest)
			return
		}

		ok(w)
	})

}

type DeleteUrlFromGroupForm struct {
	GroupID int64 `json:"groupId"`
	UrlID  int64  `json:"urlId"`
}

func DeleteUrlFromGroup(repo *urls.UrlsRepository, logger *log.Logger) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var form DeleteUrlFromGroupForm

		if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
			logError(logger, err)
			apiError(w, "decode form error", http.StatusBadRequest)
			return
		}

		// TODO - check url id

		// TODO - check group by account_id

		if err := repo.DeleteUrlFromGroup(form.GroupID, form.UrlID); err != nil {
			logError(logger, err)
			apiError(w, "decode form error", http.StatusBadRequest)
			return
		}

		ok(w)
	})

}

type ClickDataResponse struct {
	Time  time.Time
	Count int64
}

func GetClicksData(historyDB *data.HistoryDB, logger *log.Logger) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		claims := r.Context().Value("user").(*JWTClaims)
		accountID := claims.AccountID

		if r.Method != "GET" {
			apiError(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		urlArg := r.URL.Query()["url"]
		if len(urlArg) != 1 {
			apiError(w, "invalid number of query values for parameter <url>, must be 1", http.StatusBadRequest)
			return
		}

		startArg := r.URL.Query()["start"]
		if len(startArg) != 1 {
			apiError(w, "invalid number of query values for parameter <start>, must be 1", http.StatusBadRequest)
			return
		}

		endArg := r.URL.Query()["end"]
		if len(endArg) != 1 {
			apiError(w, "invalid number of query values for parameter <end>, must be 1", http.StatusBadRequest)
			return
		}

		startTime, err := time.Parse(time.RFC3339, startArg[0])
		if err != nil {
			apiError(w, "start parameter must be a valid RFC3339 datetime string", http.StatusBadRequest)
			return
		}
		endTime, err := time.Parse(time.RFC3339, endArg[0])
		if err != nil {
			apiError(w, "end parameter must be a valid RFC3339 datetime string", http.StatusBadRequest)
			return
		}

		rows, err := historyDB.GetClicksData(accountID, urlArg[0], startTime, endTime)
		if err != nil {
			logError(logger, err)
			apiError(w, "(get link data) - internal error", http.StatusInternalServerError)
			return
		}

		var list []ClickDataResponse
		for _, r := range rows {
			list = append(list, ClickDataResponse{Time: r.Time, Count: r.Count})
		}

		response(w, &list, http.StatusOK)
	})

}