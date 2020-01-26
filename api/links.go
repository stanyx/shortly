package api

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/go-chi/chi"

	"shortly/cache"
	"shortly/utils"

	"shortly/app/billing"
	"shortly/app/data"
	"shortly/app/links"
	"shortly/app/rbac"
)

func LinksRoutes(r chi.Router, auth func(rbac.Permission, http.Handler) http.HandlerFunc, linksRepository links.ILinksRepository, logger *log.Logger, historyDB *data.HistoryDB) {
	r.Get("/api/v1/users/links", auth(
		rbac.NewPermission("/api/v1/users/links", "read_links", "GET"),
		GetUserURLList(linksRepository, logger),
	))

	r.Get("/api/v1/users/links/clicks", auth(
		rbac.NewPermission("/api/v1/users/links/clicks", "get_links_clicks", "GET"),
		GetClicksData(historyDB, logger),
	))

	r.Post("/api/v1/users/links/add_group", auth(
		rbac.NewPermission("/api/v1/users/links/add_group", "add_link_to_group", "POST"),
		AddUrlToGroup(linksRepository, logger),
	))

	r.Delete("/api/v1/users/links/delete_group", auth(
		rbac.NewPermission("/api/v1/users/links/delete_group", "delete_link_from_group", "DELETE"),
		DeleteUrlFromGroup(linksRepository, logger),
	))

	r.Get("/api/v1/users/links/total", auth(
		rbac.NewPermission("/api/v1/users/links/total", "read_total_links", "GET"),
		GetTotalLinks(linksRepository, logger),
	))

}

func GetAccountID(r *http.Request) int64 {
	claims := r.Context().Value("user").(*JWTClaims)
	return claims.AccountID
}

// Public API

// TODO - auto expired links

type LinkResponse struct {
	ID          int64    `json:"id,omitempty"`
	Short       string   `json:"short"`
	Long        string   `json:"long"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	Active      bool     `json:"is_active"`
}

// TODO refactor to top links
func GetURLList(repo links.ILinksRepository, logger *log.Logger) http.HandlerFunc {

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

func GetUserURLList(repo links.ILinksRepository, logger *log.Logger) http.HandlerFunc {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		claims := r.Context().Value("user").(*JWTClaims)

		query := r.URL.Query()
		tagsFilter := query["tags"]
		shortUrlFilter := query["shortUrl"]
		longUrlFilter := query["longUrl"]
		fullTextFilter := query["fullText"]
		linkIDFilter := query["linkID"]

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

		if len(linkIDFilter) > 0 {
			linkID, _ := strconv.ParseInt(linkIDFilter[0], 0, 64)
			filters = append(filters, links.LinkFilter{LinkID: linkID})
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
				ID:          r.ID,
				Short:       r.Short,
				Long:        r.Long,
				Description: r.Description,
				Tags:        r.Tags,
				Active:      !r.Hidden,
			})
		}

		response(w, list, http.StatusOK)
	})

}

type CreateLinkForm struct {
	Url         string `json:"url"`
	Description string `json:"description"`
}

func CreateLink(repo links.ILinksRepository, urlCache cache.UrlCache, logger *log.Logger) http.HandlerFunc {

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
			Short:       repo.GenerateLink(),
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

		urlScheme := "http"
		if r.URL.Scheme != "" {
			urlScheme = r.URL.Scheme
		}

		response(w, &LinkResponse{
			Short:       urlScheme + "://" + r.Host + "/" + link.Short,
			Long:        link.Long,
			Description: link.Description,
		}, http.StatusOK)

	})

}

type UpdateLinkForm struct {
	LinkID      int64  `json:"linkId"`
	Url         string `json:"url"`
	Description string `json:"description"`
}

func UpdateLink(repo links.ILinksRepository, urlCache cache.UrlCache, logger *log.Logger) http.HandlerFunc {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		claims := r.Context().Value("user").(*JWTClaims)
		accountID := claims.AccountID

		var form UpdateLinkForm

		if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
			apiError(w, "decode form error", http.StatusBadRequest)
			return
		}

		if form.LinkID == 0 {
			apiError(w, "linkId parameter is required", http.StatusBadRequest)
			return
		}

		longURL := form.Url

		validLongURL, err := url.Parse(longURL)
		if err != nil {
			apiError(w, "url has incorrect format", http.StatusBadRequest)
			return
		}

		link, err := repo.GetLinkByID(form.LinkID)
		if err != nil {
			logError(logger, err)
			apiError(w, "get link error", http.StatusInternalServerError)
			return
		}

		link.Long = longURL
		link.Description = form.Description

		tx, err := repo.UpdateUserLink(accountID, form.LinkID, &link)
		if err != nil {
			_ = tx.Rollback()
			logError(logger, err)
			apiError(w, "internal error", http.StatusInternalServerError)
			return
		}

		urlCache.Store(link.Short, validLongURL.String())

		if err := tx.Commit(); err != nil {
			logError(logger, err)
			apiError(w, "internal error", http.StatusInternalServerError)
			return
		}

		urlScheme := "http"
		if r.URL.Scheme != "" {
			urlScheme = r.URL.Scheme
		}

		response(w, &LinkResponse{
			Short:       urlScheme + "://" + r.Host + "/" + link.Short,
			Long:        link.Long,
			Description: link.Description,
		}, http.StatusOK)

	})

}

func CreateUserLink(repo *links.LinksRepository, historyDB *data.HistoryDB, urlCache cache.UrlCache, billingLimiter *billing.BillingLimiter, logger *log.Logger) http.HandlerFunc {

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

		l := billingLimiter.Lock(accountID)
		defer l.Unlock()

		tx, linkID, err := repo.CreateUserLink(accountID, link)
		if err != nil {
			logError(logger, err)
			apiError(w, "(create link) - internal error", http.StatusInternalServerError)
			return
		}

		if err := billingLimiter.Reduce("url_limit", accountID); err != nil {
			_ = tx.Rollback()
			logError(logger, err)
			apiError(w, "(create link) - internal error", http.StatusInternalServerError)
			return
		}

		if err := historyDB.InsertDetail(link.Short, accountID); err != nil {
			_ = tx.Rollback()
			logError(logger, err)
			apiError(w, "(create link) - internal error", http.StatusInternalServerError)
			return
		}

		if err := tx.Commit(); err != nil {
			_ = billingLimiter.Reset("url_limit", accountID)
			logError(logger, err)
			apiError(w, "internal error", http.StatusInternalServerError)
			return
		}

		urlCache.Store(link.Short, link.Long)

		urlScheme := "http"
		if r.URL.Scheme != "" {
			urlScheme = r.URL.Scheme
		}

		response(w, &LinkResponse{
			ID:          linkID,
			Short:       urlScheme + "://" + r.Host + "/" + link.Short,
			Long:        link.Long,
			Description: link.Description,
		}, http.StatusOK)
	})

}

func DeleteUserLink(repo *links.LinksRepository, urlCache cache.UrlCache, billingLimiter *billing.BillingLimiter, logger *log.Logger) http.HandlerFunc {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		claims := r.Context().Value("user").(*JWTClaims)
		accountID := claims.AccountID

		linkIDArg := chi.URLParam(r, "linkID")

		if linkIDArg == "" {
			apiError(w, "url parameter is required", http.StatusBadRequest)
			return
		}

		linkID, err := strconv.ParseInt(linkIDArg, 0, 64)
		if err != nil {
			apiError(w, "linkID is not a number", http.StatusBadRequest)
			return
		}

		links, err := repo.GetUserLinks(accountID, claims.UserID, links.LinkFilter{LinkID: linkID})
		if err != nil {
			logError(logger, err)
			apiError(w, "internal error", http.StatusInternalServerError)
			return
		}

		link := links[0]

		lock := billingLimiter.Lock(accountID)
		defer lock.Unlock()

		tx, _, err := repo.DeleteUserLink(accountID, linkID)
		if err != nil {
			_ = tx.Rollback()
			logError(logger, err)
			apiError(w, "internal error", http.StatusInternalServerError)
			return
		}

		if err := billingLimiter.Increase("url_limit", accountID); err != nil {
			_ = tx.Rollback()
			logError(logger, err)
			apiError(w, "internal error", http.StatusInternalServerError)
			return
		}

		urlCache.Delete(link.Short)

		if err := tx.Commit(); err != nil {
			_ = billingLimiter.Reset("url_limit", accountID)
			logError(logger, err)
			apiError(w, "internal error", http.StatusInternalServerError)
			return
		}

		ok(w)
	})
}

type AddUrlToGroupForm struct {
	GroupID int64 `json:"groupId"`
	UrlID   int64 `json:"urlId"`
}

func AddUrlToGroup(repo links.ILinksRepository, logger *log.Logger) http.HandlerFunc {

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
	UrlID   int64 `json:"urlId"`
}

func DeleteUrlFromGroup(repo links.ILinksRepository, logger *log.Logger) http.HandlerFunc {

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

func GetClicksData(historyDB *data.HistoryDB, logger *log.Logger) http.HandlerFunc {

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

func UploadLinksInBulk(limiter *billing.BillingLimiter, repo *links.LinksRepository, historyDB *data.HistoryDB, urlCache cache.UrlCache, logger *log.Logger) http.HandlerFunc {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		claims := r.Context().Value("user").(*JWTClaims)
		accountID := claims.AccountID

		if err := r.ParseMultipartForm(10 << 20); err != nil {
			logError(logger, err)
			fmt.Fprintf(w, "parse form error")
			return
		}

		file, _, err := r.FormFile("file")
		if err != nil {
			logError(logger, err)
			fmt.Fprintf(w, "internal server error")
			return
		}
		defer file.Close()

		reader := csv.NewReader(bufio.NewReader(file))

		var links []string

		for {
			line, err := reader.Read()
			if err == io.EOF {
				break
			} else if err != nil {
				logError(logger, err)
				fmt.Fprintf(w, "error")
				return
			}
			links = append(links, line[0])
		}

		option, err := limiter.GetOptionValue("url_limit", accountID)

		if err != nil {
			logError(logger, err)
			fmt.Fprintf(w, "error")
			return
		}

		maxLinks, _ := strconv.ParseInt(option.Value, 0, 64)

		if len(links) >= int(maxLinks) {
			fmt.Fprintf(w, "plan limit exceeded")
			return
		}

		linksCreated, err := repo.BulkCreateLinks(accountID, links)
		if err != nil {
			logError(logger, err)
			fmt.Fprintf(w, "error")
			return
		}

		for _, l := range linksCreated {

			// TODO - make one method for creating links

			urlCache.Store(l.Short, l.Long)

			if err := limiter.Reduce("url_limit", accountID); err != nil {
				logError(logger, err)
				fmt.Fprintf(w, "error")
				return
			}

			if err := historyDB.InsertDetail(l.Short, accountID); err != nil {
				logError(logger, err)
				fmt.Fprintf(w, "error")
				return
			}

		}

		fmt.Fprintf(w, "ok")
	})
}

type HideLinkForm struct {
	LinkID int64 `json:"linkId"`
}

func HideUserLink(repo *links.LinksRepository, urlCache cache.UrlCache, logger *log.Logger) http.HandlerFunc {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		claims := r.Context().Value("user").(*JWTClaims)
		accountID := claims.AccountID

		var form HideLinkForm

		if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
			apiError(w, "decode form error", http.StatusBadRequest)
			return
		}

		if form.LinkID == 0 {
			apiError(w, "linkId parameter is required", http.StatusBadRequest)
			return
		}

		link, err := repo.GetLinkByID(form.LinkID)
		if err != nil {
			apiError(w, "get link error", http.StatusBadRequest)
			return
		}

		tx, err := repo.HideUserLink(accountID, form.LinkID)
		if err != nil {
			_ = tx.Rollback()
			logError(logger, err)
			apiError(w, "internal error", http.StatusInternalServerError)
			return
		}

		urlCache.Delete(link.Short)

		if err := tx.Commit(); err != nil {
			logError(logger, err)
			apiError(w, "internal error", http.StatusInternalServerError)
			return
		}

		ok(w)

	})

}

func GetTotalLinks(repo links.ILinksRepository, logger *log.Logger) http.HandlerFunc {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		claims := r.Context().Value("user").(*JWTClaims)
		accountID := claims.AccountID

		count, err := repo.GetUserLinksCount(accountID, time.Time{}, time.Now())
		if err != nil {
			logError(logger, err)
			apiError(w, "internal error", http.StatusInternalServerError)
			return
		}

		response(w, count, http.StatusOK)

	})
}
