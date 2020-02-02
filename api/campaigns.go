package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi"

	"shortly/api/response"

	"shortly/app/campaigns"
	"shortly/app/rbac"
)

// CampaignRoutes ...
func CampaignRoutes(r chi.Router, auth func(rbac.Permission, http.Handler) http.HandlerFunc, repo *campaigns.Repository, logger *log.Logger) {
	r.Get("/api/v1/campaigns", http.HandlerFunc(auth(
		rbac.NewPermission("/api/v1/campaings", "read_campaigns", "GET"),
		GetUserCampaigns(repo, logger),
	)))
	r.Post("/api/v1/campaigns/create", http.HandlerFunc(auth(
		rbac.NewPermission("/api/v1/campaings", "create_campaign", "POST"),
		CreateCampaign(repo, logger),
	)))
	r.Post("/api/v1/campaigns/start", http.HandlerFunc(auth(
		rbac.NewPermission("/api/v1/campaings/start", "start_campaign", "POST"),
		StartCampaign(repo, logger),
	)))
	r.Post("/api/v1/campaigns/stop", http.HandlerFunc(auth(
		rbac.NewPermission("/api/v1/campaings/stop", "stop_campaign", "POST"),
		StopCampaign(repo, logger),
	)))
	r.Delete("/api/v1/campaigns/delete", http.HandlerFunc(auth(
		rbac.NewPermission("/api/v1/campaings/delete", "delete_campaign", "DELETE"),
		DeleteCampaign(repo, logger),
	)))
	r.Post("/api/v1/campaigns/add_link", http.HandlerFunc(auth(
		rbac.NewPermission("/api/v1/campaings/add_link", "add_link_to_campaign", "POST"),
		AddLinkToCampaign(repo, logger),
	)))
	r.Delete("/api/v1/campaigns/delete_link", http.HandlerFunc(auth(
		rbac.NewPermission("/api/v1/campaings/delete_link", "delete_link_from_campaign", "DELETE"),
		DeleteLinkFromCampaign(repo, logger),
	)))
	r.Get("/api/v1/campaigns/data", http.HandlerFunc(auth(
		rbac.NewPermission("/api/v1/campaings/data", "get_link_data_for_campaign", "GET"),
		GetLinkDataForCampaign(repo, logger),
	)))
}

// CampaignResponse ...
type CampaignResponse struct {
	ID          int64                  `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Links       []CampaignLinkResponse `json:"links"`
}

// CampaignLinkResponse ...
type CampaignLinkResponse struct {
	ID          int64  `json:"id"`
	ShortUrl    string `json:"shortUrl"`
	LongUrl     string `json:"longUrl"`
	Description string `json:"description"`
}

// GetUserCampaigns ...
func GetUserCampaigns(repo *campaigns.Repository, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		claims := r.Context().Value("user").(*JWTClaims)
		accountID := claims.AccountID

		cmps, err := repo.GetUserCampaigns(accountID)
		if err != nil {
			logError(logger, err)
			response.Error(w, "get campaigns error", http.StatusBadRequest)
			return
		}

		var resp []CampaignResponse
		for _, cmp := range cmps {

			var links []CampaignLinkResponse

			for _, l := range cmp.Links {
				links = append(links, CampaignLinkResponse{
					ID:          l.ID,
					ShortUrl:    l.Short,
					LongUrl:     l.Long,
					Description: l.Description,
				})
			}

			resp = append(resp, CampaignResponse{
				ID:          cmp.ID,
				Name:        cmp.Name,
				Description: cmp.Description,
				Links:       links,
			})
		}

		response.Object(w, resp, http.StatusOK)
	})
}

// CreateCampaignForm ...
type CreateCampaignForm struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// CampaignCreateResponse ...
type CampaignCreateResponse struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// CreateCampaign ...
func CreateCampaign(repo *campaigns.Repository, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		claims := r.Context().Value("user").(*JWTClaims)
		accountID := claims.AccountID

		var form CreateCampaignForm

		if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
			logError(logger, err)
			response.Error(w, "decode form error", http.StatusBadRequest)
			return
		}

		row := campaigns.Campaign{
			Name:        form.Name,
			Description: form.Description,
			AccountID:   accountID,
		}

		rowID, err := repo.CreateCampaign(row)
		if err != nil {
			logError(logger, err)
			response.Error(w, "create campaign error", http.StatusBadRequest)
			return
		}

		response.Object(w, CampaignCreateResponse{
			ID:          rowID,
			Name:        row.Name,
			Description: row.Description,
		}, http.StatusOK)
	})
}

// StartCampaignForm ...
type StartCampaignForm struct {
	CampaignID int64 `json:"campaignId"`
}

// StartCampaign ...
func StartCampaign(repo *campaigns.Repository, logger *log.Logger) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var form StartCampaignForm

		if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
			logError(logger, err)
			response.Error(w, "decode form error", http.StatusBadRequest)
			return
		}

		if form.CampaignID == 0 {
			response.Error(w, "campaignId is required", http.StatusBadRequest)
			return
		}

		err := repo.StartCampaign(form.CampaignID)
		if err != nil {
			logError(logger, err)
			response.Error(w, "create form error", http.StatusBadRequest)
			return
		}

		response.Ok(w)
	})
}

// StopCampaignForm ...
type StopCampaignForm struct {
	CampaignID int64 `json:"campaignId"`
}

// StopCampaign ...
func StopCampaign(repo *campaigns.Repository, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var form StopCampaignForm

		if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
			logError(logger, err)
			response.Error(w, "decode form error", http.StatusBadRequest)
			return
		}

		if form.CampaignID == 0 {
			response.Error(w, "campaignId is required", http.StatusBadRequest)
			return
		}

		err := repo.StopCampaign(form.CampaignID)
		if err != nil {
			logError(logger, err)
			response.Error(w, "create form error", http.StatusBadRequest)
			return
		}

		response.Ok(w)
	})
}

// DeleteCampaignForm ...
type DeleteCampaignForm struct {
	CampaignID int64 `json:"campaignId"`
}

// DeleteCampaign ...
func DeleteCampaign(repo *campaigns.Repository, logger *log.Logger) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var form DeleteCampaignForm

		if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
			logError(logger, err)
			response.Error(w, "decode form error", http.StatusBadRequest)
			return
		}

		if form.CampaignID == 0 {
			response.Error(w, "campaignId is required", http.StatusBadRequest)
			return
		}

		err := repo.DeleteCampaign(form.CampaignID)
		if err != nil {
			logError(logger, err)
			response.Error(w, "create form error", http.StatusBadRequest)
			return
		}

		response.Ok(w)
	})
}

// AddLinkToCampaignForm ...
type AddLinkToCampaignForm struct {
	CampaignID int64  `json:"campaignId"`
	LinkID     int64  `json:"linkId"`
	UtmSource  string `json:"utmSource"`
	UtmMedium  string `json:"utmMedium"`
	UtmTerm    string `json:"utmTerm"`
	UtmContent string `json:"utmContent"`
}

// AddLinkToCampaign ...
func AddLinkToCampaign(repo *campaigns.Repository, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var form AddLinkToCampaignForm

		if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
			logError(logger, err)
			response.Error(w, "decode form error", http.StatusBadRequest)
			return
		}

		if form.CampaignID == 0 {
			response.Error(w, "campaignId is required", http.StatusBadRequest)
			return
		}

		if form.LinkID == 0 {
			response.Error(w, "linkId is required", http.StatusBadRequest)
			return
		}

		utm := campaigns.UTMSetting{
			Source:  form.UtmSource,
			Medium:  form.UtmMedium,
			Term:    form.UtmTerm,
			Content: form.UtmContent,
		}
		_, err := repo.AddLinkToCampaign(form.CampaignID, form.LinkID, utm)
		if err != nil {
			logError(logger, err)
			response.Error(w, "create form error", http.StatusBadRequest)
			return
		}

		response.Ok(w)
	})
}

// DeleteLinkFromCampaignForm ...
type DeleteLinkFromCampaignForm struct {
	CampaignID int64 `json:"campaignId"`
	LinkID     int64 `json:"linkId"`
}

// DeleteLinkFromCampaign ...
func DeleteLinkFromCampaign(repo *campaigns.Repository, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var form DeleteLinkFromCampaignForm

		if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
			logError(logger, err)
			response.Error(w, "decode form error", http.StatusBadRequest)
			return
		}

		if form.CampaignID == 0 {
			response.Error(w, "campaignId is required", http.StatusBadRequest)
			return
		}

		if form.LinkID == 0 {
			response.Error(w, "linkId is required", http.StatusBadRequest)
			return
		}

		err := repo.DeleteLinkFromCampaign(form.CampaignID, form.LinkID)
		if err != nil {
			logError(logger, err)
			response.Error(w, "create form error", http.StatusBadRequest)
			return
		}

		response.Ok(w)
	})
}

// CampaignClickDataResponse ...
type CampaignClickDataResponse struct {
	ShortURL string              `json:"shortUrl"`
	Data     []ClickDataResponse `json:"data"`
}

// GetLinkDataForCampaign ...
func GetLinkDataForCampaign(repo *campaigns.Repository, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := r.Context().Value("user").(*JWTClaims)
		accountID := claims.AccountID

		if r.Method != "GET" {
			response.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		campaignIDArg := r.URL.Query()["campaignId"]
		if len(campaignIDArg) != 1 {
			response.Error(w, "invalid number of query values for parameter <campaignID>, must be 1", http.StatusBadRequest)
			return
		}

		campaignID, err := strconv.ParseInt(campaignIDArg[0], 0, 64)
		if err != nil {
			response.Error(w, "campaignID is not number", http.StatusBadRequest)
			return
		}

		startArg := r.URL.Query()["start"]
		if len(startArg) != 1 {
			response.Error(w, "invalid number of query values for parameter <start>, must be 1", http.StatusBadRequest)
			return
		}

		endArg := r.URL.Query()["end"]
		if len(endArg) != 1 {
			response.Error(w, "invalid number of query values for parameter <end>, must be 1", http.StatusBadRequest)
			return
		}

		startTime, err := time.Parse(time.RFC3339, startArg[0])
		if err != nil {
			response.Error(w, "start parameter must be a valid RFC3339 datetime string", http.StatusBadRequest)
			return
		}
		endTime, err := time.Parse(time.RFC3339, endArg[0])
		if err != nil {
			response.Error(w, "end parameter must be a valid RFC3339 datetime string", http.StatusBadRequest)
			return
		}

		rows, err := repo.GetCampaignClicksData(accountID, campaignID, startTime, endTime)
		if err != nil {
			logError(logger, err)
			response.Error(w, "(get link data) - internal error", http.StatusInternalServerError)
			return
		}

		var list []CampaignClickDataResponse
		for _, l := range rows {

			var data []ClickDataResponse
			for _, linkData := range l.Data {
				data = append(data, ClickDataResponse{
					Time:  linkData.Time,
					Count: linkData.Count,
				})
			}

			list = append(list, CampaignClickDataResponse{
				ShortURL: l.ShortURL,
				Data:     data,
			})
		}

		response.Object(w, &list, http.StatusOK)
	})
}
