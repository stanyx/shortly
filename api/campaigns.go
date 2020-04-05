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

// CampaignRoutes http handlers for campaigns application
func CampaignRoutes(r chi.Router, auth func(rbac.Permission, http.Handler) http.HandlerFunc, repo *campaigns.Repository, logger *log.Logger) {
	r.Get("/api/v1/campaigns", http.HandlerFunc(auth(
		rbac.NewPermission("/api/v1/campaigns", "read_campaigns", "GET"),
		GetUserCampaigns(repo, logger),
	)))
	r.Post("/api/v1/campaigns", http.HandlerFunc(auth(
		rbac.NewPermission("/api/v1/campaigns", "create_campaign", "POST"),
		CreateCampaign(repo, logger),
	)))
	r.Put("/api/v1/campaigns/{id}", http.HandlerFunc(auth(
		rbac.NewPermission("/api/v1/campaigns/{id}", "update_campaign", "PUT"),
		UpdateCampaign(repo, logger),
	)))
	r.Post("/api/v1/campaigns/start", http.HandlerFunc(auth(
		rbac.NewPermission("/api/v1/campaigns/start", "start_campaign", "POST"),
		StartCampaign(repo, logger),
	)))
	r.Post("/api/v1/campaigns/stop", http.HandlerFunc(auth(
		rbac.NewPermission("/api/v1/campaigns/stop", "stop_campaign", "POST"),
		StopCampaign(repo, logger),
	)))
	r.Delete("/api/v1/campaigns/{id}", http.HandlerFunc(auth(
		rbac.NewPermission("/api/v1/campaigns/{id}", "delete_campaign", "DELETE"),
		DeleteCampaign(repo, logger),
	)))
	r.Get("/api/v1/campaigns/{id}/channels/{channelId}/links", http.HandlerFunc(auth(
		rbac.NewPermission("/api/v1/campaigns/{id}/channels/{channelId}/links", "get_links_for_channel", "GET"),
		GetChannelLinks(repo, logger),
	)))
	r.Post("/api/v1/campaigns/{id}/channels/{channelId}/links", http.HandlerFunc(auth(
		rbac.NewPermission("/api/v1/campaigns/{id}/channels/{channelId}/links", "add_link_to_channel", "POST"),
		AddLinkToCampaignChannel(repo, logger),
	)))
	r.Delete("/api/v1/campaigns/{id}/channels/{channelId}/links/{linkId}", http.HandlerFunc(auth(
		rbac.NewPermission("/api/v1/campaigns/{id}/channels/{channelId}/links/{linkId}", "delete_link_from_channel", "DELETE"),
		DeleteLinkFromCampaignChannel(repo, logger),
	)))
	r.Get("/api/v1/campaigns/data", http.HandlerFunc(auth(
		rbac.NewPermission("/api/v1/campaigns/data", "get_link_data_for_campaign", "GET"),
		GetLinkDataForCampaign(repo, logger),
	)))
	r.Get("/api/v1/campaigns/{id}/freechannels", http.HandlerFunc(auth(
		rbac.NewPermission("/api/v1/campaigns/{id}/freechannels", "get_channels", "GET"),
		GetChannels(repo, logger),
	)))
	r.Post("/api/v1/channels", http.HandlerFunc(auth(
		rbac.NewPermission("/api/v1/channels", "create_channel", "POST"),
		CreateChannel(repo, logger),
	)))
	r.Get("/api/v1/freelinks/{channelId}", http.HandlerFunc(auth(
		rbac.NewPermission("/api/v1/freelinks/{channelId}", "get_freelinks", "GET"),
		GetCampaignFreeLinks(repo, logger),
	)))
	r.Get("/api/v1/campaigns/{id}/channels", http.HandlerFunc(auth(
		rbac.NewPermission("/api/v1/campaigns/{id}/channels", "get_channels_for_campaign", "GET"),
		GetCampaignChannels(repo, logger),
	)))
	r.Post("/api/v1/campaigns/{id}/channels", http.HandlerFunc(auth(
		rbac.NewPermission("/api/v1/campaigns/{id}/channels", "add_channel_to_campaign", "POST"),
		AddChannelToCampaign(repo, logger),
	)))
	r.Delete("/api/v1/campaigns/{id}/channels/{channelId}", http.HandlerFunc(auth(
		rbac.NewPermission("/api/v1/campaigns/{id}/channels/{channelId}", "delete_channel_from_campaign", "DELETE"),
		DeleteChannelFromCampaign(repo, logger),
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
	ChannelID   int64  `json:"channelId"`
	ChannelName string `json:"channelName"`
}

// GetUserCampaigns request handler returns a list consists of campaigns created for current authorized account
// @Tags Campaigns
// @Description read campaigns list for current authorized account
// @ID get-all-campaigns
// @Produce  json
// @Success 200 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Failure 500 {object} response.ApiResponse
// @Router /campaigns [post]
func GetUserCampaigns(repo campaigns.CampaignRepository, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		claims := r.Context().Value("user").(*JWTClaims)
		accountID := claims.AccountID

		cmps, err := repo.GetUserCampaigns(accountID)
		if err != nil {
			logError(logger, err)
			response.Error(w, "get campaigns error", http.StatusInternalServerError)
			return
		}

		resp := make([]CampaignResponse, 0)
		for _, cmp := range cmps {

			var links []CampaignLinkResponse

			for _, l := range cmp.Links {
				links = append(links, CampaignLinkResponse{
					ID:          l.ID,
					ShortUrl:    l.ShortUrl,
					LongUrl:     l.LongUrl,
					Description: l.Description,
					ChannelID:   l.ChannelID,
					ChannelName: l.ChannelName,
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
			response.Bad(w, "decode form error")
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
			response.InternalError(w, "create campaign error")
			return
		}

		response.Object(w, CampaignCreateResponse{
			ID:          rowID,
			Name:        row.Name,
			Description: row.Description,
		}, http.StatusOK)
	})
}

// UpdateCampaignForm ...
type UpdateCampaignForm struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// UpdateCampaign ...
func UpdateCampaign(repo *campaigns.Repository, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		claims := r.Context().Value("user").(*JWTClaims)
		accountID := claims.AccountID

		idArg := chi.URLParam(r, "id")
		if idArg == "" {
			response.Bad(w, "id parameter is required")
			return
		}

		id, err := strconv.ParseInt(idArg, 0, 64)
		if err != nil {
			response.Bad(w, "id is not a number")
			return
		}

		if id == 0 {
			response.Bad(w, "id must be greater than zero")
			return
		}

		var form UpdateCampaignForm

		if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
			logError(logger, err)
			response.Bad(w, "decode form error")
			return
		}

		row := campaigns.Campaign{
			Name:        form.Name,
			Description: form.Description,
			AccountID:   accountID,
		}

		if err := repo.UpdateCampaign(id, row); err != nil {
			logError(logger, err)
			response.InternalError(w, "update campaign error")
			return
		}

		response.Ok(w)
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

// DeleteCampaign ...
func DeleteCampaign(repo *campaigns.Repository, logger *log.Logger) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		idArg := chi.URLParam(r, "id")

		if idArg == "" {
			response.Error(w, "url parameter is required", http.StatusBadRequest)
			return
		}

		id, err := strconv.ParseInt(idArg, 0, 64)
		if err != nil {
			response.Error(w, "id is not a number", http.StatusBadRequest)
			return
		}

		if id == 0 {
			response.Error(w, "campaignId is required", http.StatusBadRequest)
			return
		}

		if err := repo.DeleteCampaign(id); err != nil {
			logError(logger, err)
			response.Error(w, "create form error", http.StatusBadRequest)
			return
		}

		response.Ok(w)
	})
}

func GetChannelLinks(repo *campaigns.Repository, logger *log.Logger) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		accountID := r.Context().Value("user").(*JWTClaims).AccountID

		campaignIDArg := chi.URLParam(r, "id")

		if campaignIDArg == "" {
			response.Error(w, "id parameter is required", http.StatusBadRequest)
			return
		}

		campaignID, err := strconv.ParseInt(campaignIDArg, 0, 64)
		if err != nil {
			response.Error(w, "id is not a number", http.StatusBadRequest)
			return
		}

		if campaignID <= 0 {
			response.Error(w, "id parameter value must be greater than zero", http.StatusBadRequest)
			return
		}

		channelIDArg := chi.URLParam(r, "channelId")

		if channelIDArg == "" {
			response.Error(w, "channelID parameter is required", http.StatusBadRequest)
			return
		}

		channelID, err := strconv.ParseInt(channelIDArg, 0, 64)
		if err != nil {
			response.Error(w, "channelID is not a number", http.StatusBadRequest)
			return
		}

		if channelID == 0 {
			response.Error(w, "channelID parameter value must be greater than zero", http.StatusBadRequest)
			return
		}

		rows, err := repo.GetChannelLinks(accountID, campaignID, channelID)
		if err != nil {
			logError(logger, err)
			response.Error(w, "(get channel links) - internal error", http.StatusInternalServerError)
			return
		}

		var list []LinkResponse
		for _, r := range rows {
			list = append(list, LinkResponse{
				ID:    r.ID,
				Short: r.Short,
				Long:  r.Long,
			})
		}

		response.Object(w, &list, http.StatusOK)
	})
}

// AddLinkToCampaignForm ...
type AddLinkToCampaignForm struct {
	LinkID     int64  `json:"linkId"`
	UtmSource  string `json:"utmSource"`
	UtmMedium  string `json:"utmMedium"`
	UtmTerm    string `json:"utmTerm"`
	UtmContent string `json:"utmContent"`
}

// AddLinkToCampaignChannel ...
func AddLinkToCampaignChannel(repo *campaigns.Repository, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var form AddLinkToCampaignForm

		if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
			logError(logger, err)
			response.Error(w, "decode form error", http.StatusBadRequest)
			return
		}

		campaignIDArg := chi.URLParam(r, "id")

		if campaignIDArg == "" {
			response.Error(w, "id parameter is required", http.StatusBadRequest)
			return
		}

		campaignID, err := strconv.ParseInt(campaignIDArg, 0, 64)
		if err != nil {
			response.Error(w, "id is not a number", http.StatusBadRequest)
			return
		}

		if campaignID <= 0 {
			response.Error(w, "id parameter value must be greater than zero", http.StatusBadRequest)
			return
		}

		channelIDArg := chi.URLParam(r, "channelId")

		if channelIDArg == "" {
			response.Error(w, "channelID parameter is required", http.StatusBadRequest)
			return
		}

		channelID, err := strconv.ParseInt(channelIDArg, 0, 64)
		if err != nil {
			response.Error(w, "channelID is not a number", http.StatusBadRequest)
			return
		}

		if channelID == 0 {
			response.Error(w, "channelID parameter value must be greater than zero", http.StatusBadRequest)
			return
		}

		if form.LinkID <= 0 {
			response.Error(w, "linkId is required", http.StatusBadRequest)
			return
		}

		utm := campaigns.UTMSetting{
			Source:  form.UtmSource,
			Medium:  form.UtmMedium,
			Term:    form.UtmTerm,
			Content: form.UtmContent,
		}
		_, err = repo.AddLinkToCampaignChannel(campaignID, channelID, form.LinkID, utm)
		if err != nil {
			logError(logger, err)
			response.Error(w, "add link to channel error", http.StatusInternalServerError)
			return
		}

		response.Ok(w)
	})
}

// DeleteLinkFromCampaignChannel ...
func DeleteLinkFromCampaignChannel(repo *campaigns.Repository, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		campaignIDArg := chi.URLParam(r, "id")

		if campaignIDArg == "" {
			response.Error(w, "id parameter is required", http.StatusBadRequest)
			return
		}

		campaignID, err := strconv.ParseInt(campaignIDArg, 0, 64)
		if err != nil {
			response.Error(w, "id is not a number", http.StatusBadRequest)
			return
		}

		if campaignID <= 0 {
			response.Error(w, "id parameter value must be greater than zero", http.StatusBadRequest)
			return
		}

		channelIDArg := chi.URLParam(r, "channelId")

		if channelIDArg == "" {
			response.Error(w, "channelID parameter is required", http.StatusBadRequest)
			return
		}

		channelID, err := strconv.ParseInt(channelIDArg, 0, 64)
		if err != nil {
			response.Error(w, "channelID is not a number", http.StatusBadRequest)
			return
		}

		if channelID == 0 {
			response.Error(w, "channelID parameter value must be greater than zero", http.StatusBadRequest)
			return
		}

		linkIDArg := chi.URLParam(r, "linkId")

		if linkIDArg == "" {
			response.Error(w, "linkId parameter is required", http.StatusBadRequest)
			return
		}

		linkID, err := strconv.ParseInt(linkIDArg, 0, 64)
		if err != nil {
			response.Error(w, "linkId is not a number", http.StatusBadRequest)
			return
		}

		if linkID <= 0 {
			response.Error(w, "linkId parameter value must be greater than zero", http.StatusBadRequest)
			return
		}

		err = repo.DeleteLinkFromCampaignChannel(campaignID, channelID, linkID)
		if err != nil {
			logError(logger, err)
			response.Error(w, "delete link from channel error", http.StatusBadRequest)
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

		campaignData, err := repo.GetCampaignClicksData(accountID, campaignID, startTime, endTime)
		if err != nil {
			logError(logger, err)
			response.Error(w, "(get link data) - internal error", http.StatusInternalServerError)
			return
		}

		var list []CampaignClickDataResponse
		for _, l := range campaignData.LinkData {

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

type ChannelResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// GetChannels http handler returns channels that not yet assigned to provided campaign
func GetChannels(repo *campaigns.Repository, logger *log.Logger) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		accountID := r.Context().Value("user").(*JWTClaims).AccountID

		campaignIDArg := chi.URLParam(r, "id")

		if campaignIDArg == "" {
			response.Error(w, "id parameter is required", http.StatusBadRequest)
			return
		}

		campaignID, err := strconv.ParseInt(campaignIDArg, 0, 64)
		if err != nil {
			response.Error(w, "id is not a number", http.StatusBadRequest)
			return
		}

		if campaignID <= 0 {
			response.Error(w, "id parameter value must be greater than zero", http.StatusBadRequest)
			return
		}

		rows, err := repo.GetChannels(accountID, campaignID)
		if err != nil {
			logError(logger, err)
			response.Error(w, "(get channels) - internal error", http.StatusInternalServerError)
			return
		}

		var list []ChannelResponse
		for _, r := range rows {
			list = append(list, ChannelResponse{
				ID:   r.ID,
				Name: r.Name,
			})
		}

		response.Object(w, &list, http.StatusOK)
	})
}

// GetCampaignChannels http handler returns channels that has been assigned to provided campaign
func GetCampaignChannels(repo *campaigns.Repository, logger *log.Logger) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		accountID := r.Context().Value("user").(*JWTClaims).AccountID

		campaignIDArg := chi.URLParam(r, "id")

		if campaignIDArg == "" {
			response.Error(w, "id parameter is required", http.StatusBadRequest)
			return
		}

		campaignID, err := strconv.ParseInt(campaignIDArg, 0, 64)
		if err != nil {
			response.Error(w, "id is not a number", http.StatusBadRequest)
			return
		}

		if campaignID <= 0 {
			response.Error(w, "id parameter value must be greater than zero", http.StatusBadRequest)
			return
		}

		rows, err := repo.GetCampaignChannels(accountID, campaignID)
		if err != nil {
			logError(logger, err)
			response.Error(w, "(get channels) - internal error", http.StatusInternalServerError)
			return
		}

		var list []ChannelResponse
		for _, r := range rows {
			list = append(list, ChannelResponse{
				ID:   r.ID,
				Name: r.Name,
			})
		}

		response.Object(w, &list, http.StatusOK)
	})
}

// GetCampaignFreeLinks http handler returns links that has not been assigned to provided campaign
func GetCampaignFreeLinks(repo *campaigns.Repository, logger *log.Logger) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		accountID := r.Context().Value("user").(*JWTClaims).AccountID

		channelIDArg := chi.URLParam(r, "channelId")

		if channelIDArg == "" {
			response.Error(w, "channelId parameter is required", http.StatusBadRequest)
			return
		}

		channelID, err := strconv.ParseInt(channelIDArg, 0, 64)
		if err != nil {
			response.Error(w, "channelId is not a number", http.StatusBadRequest)
			return
		}

		if channelID == 0 {
			response.Error(w, "channelId is required", http.StatusBadRequest)
			return
		}

		rows, err := repo.GetCampaignFreeLinks(accountID, channelID)
		if err != nil {
			logError(logger, err)
			response.Error(w, "(get freelinks) - internal error", http.StatusInternalServerError)
			return
		}

		var list []LinkResponse
		for _, r := range rows {
			list = append(list, LinkResponse{
				ID:    r.ID,
				Short: r.Short,
				Long:  r.Long,
			})
		}

		response.Object(w, &list, http.StatusOK)
	})
}

// CreateChannelForm ...
type CreateChannelForm struct {
	Name string `json:"name"`
}

// CreateChannel ...
func CreateChannel(repo *campaigns.Repository, logger *log.Logger) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		accountID := r.Context().Value("user").(*JWTClaims).AccountID

		var form CreateChannelForm
		if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
			logError(logger, err)
			response.Error(w, "decode form error", http.StatusBadRequest)
			return
		}

		ch := &campaigns.Channel{
			Name: form.Name,
		}

		rowID, err := repo.CreateChannel(accountID, ch)
		if err != nil {
			logError(logger, err)
			response.Error(w, "(create channel) - internal error", http.StatusInternalServerError)
			return
		}

		resp := &ChannelResponse{
			ID:   rowID,
			Name: ch.Name,
		}

		response.Object(w, resp, http.StatusOK)

	})
}

// AddChannelToCampaignForm ...
type AddChannelToCampaignForm struct {
	Channels []int64 `json:"channels"`
}

// AddChannelToCampaign ...
func AddChannelToCampaign(repo *campaigns.Repository, logger *log.Logger) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var form AddChannelToCampaignForm
		if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
			logError(logger, err)
			response.Error(w, "decode form error", http.StatusBadRequest)
			return
		}

		campaignIDArg := chi.URLParam(r, "id")

		if campaignIDArg == "" {
			response.Error(w, "id parameter is required", http.StatusBadRequest)
			return
		}

		campaignID, err := strconv.ParseInt(campaignIDArg, 0, 64)
		if err != nil {
			response.Error(w, "id is not a number", http.StatusBadRequest)
			return
		}

		if campaignID <= 0 {
			response.Error(w, "id parameter value must be greater than zero", http.StatusBadRequest)
			return
		}

		err = repo.AddChannelsToCampaign(campaignID, form.Channels)
		if err != nil {
			logError(logger, err)
			response.Error(w, "(create channel) - internal error", http.StatusInternalServerError)
			return
		}

		response.Ok(w)

	})
}

// DeleteChannelFromCampaign ...
func DeleteChannelFromCampaign(repo *campaigns.Repository, logger *log.Logger) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		campaignIDArg := chi.URLParam(r, "id")

		if campaignIDArg == "" {
			response.Error(w, "id parameter is required", http.StatusBadRequest)
			return
		}

		campaignID, err := strconv.ParseInt(campaignIDArg, 0, 64)
		if err != nil {
			response.Error(w, "id is not a number", http.StatusBadRequest)
			return
		}

		if campaignID <= 0 {
			response.Error(w, "id parameter value must be greater than zero", http.StatusBadRequest)
			return
		}

		channelIDArg := chi.URLParam(r, "channelId")

		if channelIDArg == "" {
			response.Error(w, "channelId parameter is required", http.StatusBadRequest)
			return
		}

		channelID, err := strconv.ParseInt(channelIDArg, 0, 64)
		if err != nil {
			response.Error(w, "channelId is not a number", http.StatusBadRequest)
			return
		}

		if channelID <= 0 {
			response.Error(w, "channelId parameter value must be greater than zero", http.StatusBadRequest)
			return
		}

		_, err = repo.DeleteChannelFromCampaign(campaignID, channelID)
		if err != nil {
			logError(logger, err)
			response.Error(w, "(delete channel) - internal error", http.StatusInternalServerError)
			return
		}

		response.Ok(w)

	})
}
