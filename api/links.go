package api

import (
	"log"
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"shortly/cache"
	"shortly/utils"

	"shortly/app/links"
	"shortly/app/billing"
	"shortly/app/data"
)


// Public API

// TODO - auto expired links

type LinkResponse struct {
	ID          int64    `json:"id,omitempty"`
	Short       string   `json:"short"`
	Long        string   `json:"long"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
}

func GetURLList(repo links.ILinksRepository, logger *log.Logger) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// TODO - add search and filters functions
		// TODO - add pagination

		rows, err := repo.GetAllLinks()
		if err != nil {
			logError(logger, err)
			apiError(w, "internal error", http.StatusInternalServerError)
			return
		}

		var list []LinkResponse
		for _, r := range rows {
			list = append(list, LinkResponse{
				Short:       r.Short,
				Long:        r.Long,
				Description: r.Description,
			})
		}
		
		response(w, list, http.StatusOK)
	})

}

func GetUserURLList(repo links.ILinksRepository, logger *log.Logger) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		claims := r.Context().Value("user").(*JWTClaims)

		tagsFilter := r.URL.Query()["tags"]
		shortUrlFilter := r.URL.Query()["shortUrl"]
		longUrlFilter := r.URL.Query()["longUrl"]
		fullTextFilter := r.URL.Query()["fullText"]

		var filters []links.LinkFilter
		if len(tagsFilter) > 0 {
			filters = append(filters, links.LinkFilter{Tags: tagsFilter})
		}

		if len(shortUrlFilter) > 0 {
			filters = append(filters, links.LinkFilter{ShortUrl: shortUrlFilter})
		}

		if len(longUrlFilter) > 0 {
			filters = append(filters, links.LinkFilter{LongUrl: longUrlFilter})
		}

		if len(fullTextFilter) > 0 {
			filters = append(filters, links.LinkFilter{FullText: fullTextFilter[0]})
		}

		rows, err := repo.GetUserLinks(claims.AccountID, claims.UserID, filters...)
		if err != nil {
			logError(logger, err)
			apiError(w, "internal error", http.StatusInternalServerError)
			return
		}

		var list []LinkResponse
		for _, r := range rows {
			list = append(list, LinkResponse{
				Short:       r.Short,
				Long:        r.Long,
				Description: r.Description,
				Tags:        r.Tags,
			})
		}
		
		response(w, list, http.StatusOK)
	})

}

type CreateLinkForm struct {
	Url         string `json:"url"`
	Description string `json:"description"`
}

func CreateLink(repo links.ILinksRepository, urlCache cache.UrlCache, logger *log.Logger) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Method != "POST" {
			apiError(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var form CreateLinkForm

		if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
			apiError(w, "decode form error", http.StatusBadRequest)
			return
		}

		if form.Url == "" {
			apiError(w, "url parameter is required", http.StatusBadRequest)
			return
		}

		longURL := form.Url

		validLongURL, err := url.Parse(longURL)
		if err != nil {
			apiError(w, "url has incorrect format", http.StatusBadRequest)
			return
		}

		link := &links.Link{
			Short:       utils.RandomString(5),
			Long:        validLongURL.String(),
			Description: form.Description,
		}

		err = repo.CreateLink(link)
		if err != nil {
			logError(logger, err)
			apiError(w, "internal error", http.StatusInternalServerError)
			return
		}

		urlCache.Store(link.Short, validLongURL.String())

		response(w, &LinkResponse{
			Short:       r.Host + "/" + link.Short, 
			Long:        link.Long,
			Description: link.Description,
		}, http.StatusOK)

	})

}

func CreateUserLink(repo *links.LinksRepository, historyDB *data.HistoryDB, urlCache cache.UrlCache, billingLimiter *billing.BillingLimiter, logger *log.Logger) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		claims := r.Context().Value("user").(*JWTClaims)
		accountID := claims.AccountID

		var form CreateLinkForm

		if r.Method != "POST" {
			apiError(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
			apiError(w, "decode form error", http.StatusBadRequest)
			return
		}

		if form.Url == "" {
			apiError(w, "url parameter is required", http.StatusBadRequest)
			return
		}

		validLongURL, err := url.Parse(form.Url)
		if err != nil {
			apiError(w, "long url has incorrect format", http.StatusBadRequest)
			return
		}

		link := &links.Link{
			Short:       utils.RandomString(5),
			Long:        validLongURL.String(),
			Description: form.Description,
		}

		urlCache.Store(link.Short, link.Long)

		linkID, err := repo.CreateUserLink(accountID, link)
		if err != nil {
			logError(logger, err)
			apiError(w, "(create link) - internal error", http.StatusInternalServerError)
			return
		}

		if err := billingLimiter.Reduce("url_limit", accountID); err != nil {
			logError(logger, err)
			apiError(w, "(create link) - internal error", http.StatusInternalServerError)
			return
		}

		if err := historyDB.InsertDetail(link.Short, accountID); err != nil {
			logError(logger, err)
			apiError(w, "(create link) - internal error", http.StatusInternalServerError)
			return
		}

		response(w, &LinkResponse{
			ID:          linkID,
			Short:       r.Host + "/" + link.Short, 
			Long:        link.Long,
			Description: link.Description,
		}, http.StatusOK)
	})

}

type DeleteLinkForm struct {
	Url string `json:"url"`
}

func DeleteUserLink(repo *links.LinksRepository, urlCache cache.UrlCache, logger *log.Logger) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		claims := r.Context().Value("user").(*JWTClaims)
		accountID := claims.AccountID

		if r.Method != "DELETE" {
			apiError(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var form DeleteLinkForm
		if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
			apiError(w, "decode form error", http.StatusBadRequest)
			return
		}

		if form.Url == "" {
			apiError(w, "url parameter is required", http.StatusBadRequest)
			return
		}

		_, err := repo.DeleteUserLink(accountID, form.Url)
		if err != nil {
			logError(logger, err)
			apiError(w, "internal error", http.StatusInternalServerError)
			return
		}
			
		urlCache.Delete(form.Url)
		
		ok(w)
	})
}

type AddUrlToGroupForm struct {
	GroupID int64 `json:"groupId"`
	UrlID  int64  `json:"urlId"`
}

func AddUrlToGroup(repo *links.LinksRepository, logger *log.Logger) http.Handler {

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

func DeleteUrlFromGroup(repo *links.LinksRepository, logger *log.Logger) http.Handler {

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