package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"

	"shortly/api/response"

	"shortly/app/rbac"
	"shortly/app/webhooks"
)

func WebhooksRoutes(r chi.Router, auth func(rbac.Permission, http.Handler) http.HandlerFunc, repo webhooks.Repository, logger *log.Logger) {

	r.Get("/api/v1/webhooks", auth(
		rbac.NewPermission("/api/v1/webhooks", "read_webhooks", "GET"),
		GetWebhooks(repo, logger),
	))

	r.Post("/api/v1/webhooks/create", auth(
		rbac.NewPermission("/api/v1/webhooks/create", "create_webhook", "POST"),
		CreateWebhook(repo, logger),
	))

	r.Put("/api/v1/webhooks/update", auth(
		rbac.NewPermission("/api/v1/webhooks/update", "update_webhook", "PUT"),
		UpdateWebhook(repo, logger),
	))

	r.Delete("/api/v1/webhooks/{id}", auth(
		rbac.NewPermission("/api/v1/webhooks/{id}", "delete_webhook", "DELETE"),
		DeleteWebhook(repo, logger),
	))

	r.Post("/api/v1/webhooks/{id}/enable", auth(
		rbac.NewPermission("/api/v1/webhooks/{id}/enable", "enable_webhook", "POST"),
		EnableWebhook(repo, logger),
	))

	r.Post("/api/v1/webhooks/{id}/disable", auth(
		rbac.NewPermission("/api/v1/webhooks/{id}/disable", "disable_webhook", "POST"),
		DisableWebhook(repo, logger),
	))

	r.Post("/webhook/{event}", TestWebhook(repo, logger))
}

type WebhookResponse struct {
	ID          int64    `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Events      []string `json:"events"`
	URL         string   `json:"url"`
	Active      bool     `json:"active"`
}

func GetWebhooks(repo webhooks.Repository, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		claims := r.Context().Value("user").(*JWTClaims)

		rows, err := repo.GetWebhooks(claims.AccountID)
		if err != nil {
			logError(logger, err)
			response.Error(w, "get webhooks error", http.StatusBadRequest)
			return
		}

		var resp []WebhookResponse
		for _, r := range rows {
			resp = append(resp, WebhookResponse{
				ID:          r.ID,
				Name:        r.Name,
				Description: r.Description,
				Events:      r.Events,
				URL:         r.URL,
				Active:      r.Active,
			})
		}

		response.Object(w, resp, http.StatusOK)
	})
}

type CreateWebhookForm struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Events      []string `json:"events"`
	URL         string   `json:"url"`
}

type CreateWebhookResponse struct {
	ID          int64    `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Events      []string `json:"events"`
	URL         string   `json:"url"`
}

func CreateWebhook(repo webhooks.Repository, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		claims := r.Context().Value("user").(*JWTClaims)

		var form CreateWebhookForm
		if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
			logError(logger, err)
			response.Error(w, "decode form error", http.StatusBadRequest)
			return
		}

		m := webhooks.Webhook{
			Name:        form.Name,
			Description: form.Description,
			Events:      form.Events,
			URL:         form.URL,
		}

		rowID, err := repo.CreateWebhook(claims.AccountID, m)
		if err != nil {
			logError(logger, err)
			response.Error(w, "create webhook error", http.StatusBadRequest)
			return
		}

		response.Object(w, WebhookResponse{
			ID:          rowID,
			Name:        m.Name,
			Description: m.Description,
			Events:      m.Events,
			URL:         m.URL,
		}, http.StatusOK)
	})
}

type UpdateWebhookForm struct {
	ID          int64    `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Events      []string `json:"events"`
	URL         string   `json:"url"`
}

type UpdateWebhookResponse struct {
	ID          int64    `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Events      []string `json:"events"`
	URL         string   `json:"url"`
}

func UpdateWebhook(repo webhooks.Repository, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		claims := r.Context().Value("user").(*JWTClaims)

		var form UpdateWebhookForm
		if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
			logError(logger, err)
			response.Error(w, "decode form error", http.StatusBadRequest)
			return
		}

		m := webhooks.Webhook{
			ID:          form.ID,
			Name:        form.Name,
			Description: form.Description,
			Events:      form.Events,
			URL:         form.URL,
		}

		err := repo.UpdateWebhook(claims.AccountID, m)
		if err != nil {
			logError(logger, err)
			response.Error(w, "update webhook error", http.StatusBadRequest)
			return
		}

		response.Object(w, UpdateWebhookResponse{
			Name:        m.Name,
			Description: m.Description,
			Events:      m.Events,
			URL:         m.URL,
		}, http.StatusOK)
	})
}

func EnableWebhook(repo webhooks.Repository, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		claims := r.Context().Value("user").(*JWTClaims)

		idArg := chi.URLParam(r, "id")
		if idArg == "" {
			response.Error(w, "id parameter is required", http.StatusBadRequest)
			return
		}

		rowID, err := strconv.ParseInt(idArg, 0, 64)
		if err != nil {
			response.Error(w, "id is not a number", http.StatusBadRequest)
			return
		}

		err = repo.EnableWebhook(claims.AccountID, rowID)
		if err != nil {
			logError(logger, err)
			response.Error(w, "enable webhook error", http.StatusBadRequest)
			return
		}

		response.Ok(w)
	})
}

func DisableWebhook(repo webhooks.Repository, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		claims := r.Context().Value("user").(*JWTClaims)

		idArg := chi.URLParam(r, "id")
		if idArg == "" {
			response.Error(w, "id parameter is required", http.StatusBadRequest)
			return
		}

		rowID, err := strconv.ParseInt(idArg, 0, 64)
		if err != nil {
			response.Error(w, "id is not a number", http.StatusBadRequest)
			return
		}

		err = repo.DisableWebhook(claims.AccountID, rowID)
		if err != nil {
			logError(logger, err)
			response.Error(w, "disable webhook error", http.StatusBadRequest)
			return
		}

		response.Ok(w)
	})
}

func DeleteWebhook(repo webhooks.Repository, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		claims := r.Context().Value("user").(*JWTClaims)

		idArg := chi.URLParam(r, "id")
		if idArg == "" {
			response.Error(w, "id parameter is required", http.StatusBadRequest)
			return
		}

		rowID, err := strconv.ParseInt(idArg, 0, 64)
		if err != nil {
			response.Error(w, "id is not a number", http.StatusBadRequest)
			return
		}

		err = repo.DeleteWebhook(claims.AccountID, rowID)
		if err != nil {
			logError(logger, err)
			response.Error(w, "delete webhook error", http.StatusBadRequest)
			return
		}

		response.Ok(w)
	})
}

func TestWebhook(repo webhooks.Repository, logger *log.Logger) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		event := chi.URLParam(r, "event")

		fmt.Println("GET WEBHOOK REQUEST: ", event)

		w.WriteHeader(http.StatusOK)

		response.Ok(w)
	})
}
