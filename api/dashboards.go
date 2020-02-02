package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"shortly/app/dashboards"
	"shortly/app/rbac"

	"github.com/go-chi/chi"
	validator "gopkg.in/go-playground/validator.v9"

	"shortly/api/response"
)

// DashboardsRoutes ...
func DashboardsRoutes(r chi.Router, auth func(rbac.Permission, http.Handler) http.HandlerFunc, repo *dashboards.Repository, logger *log.Logger) {

	r.Get("/api/v1/dashboards", auth(
		rbac.NewPermission("/api/v1/dashboards", "read_dashboard", "GET"),
		GetDashboards(repo, logger),
	))

	r.Post("/api/v1/dashboards/create", auth(
		rbac.NewPermission("/api/v1/dashboards/create", "create_dashboard", "POST"),
		CreateDashboard(repo, logger),
	))

	r.Get("/api/v1/dashboards/{id}/widgets", auth(
		rbac.NewPermission("/api/v1/dashboards/{id}/widgets", "read_dashboard_widgets", "GET"),
		GetDashboardWidgets(repo, logger),
	))

	r.Post("/api/v1/dashboards/{id}/widgets", auth(
		rbac.NewPermission("/api/v1/dashboards/{id}/widgets", "add_dashboard_widget", "POST"),
		AddDashboardWidget(repo, logger),
	))

	r.Delete("/api/v1/dashboards/{id}", auth(
		rbac.NewPermission("/api/v1/dashboards/{id}", "delete_dashboard", "DELETE"),
		DeleteDashboard(repo, logger),
	))

	r.Delete("/api/v1/dashboards/{id}/widgets/{widgetID}", auth(
		rbac.NewPermission("/api/v1/dashboards/{id}/widgets/{widgetID}", "delete_dashboard_widget", "DELETE"),
		DeleteDashboardWidget(repo, logger),
	))

}

// DashboardResponse ...
type DashboardResponse struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
}

// GetDashboards ...
func GetDashboards(repo *dashboards.Repository, logger *log.Logger) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		claims := r.Context().Value("user").(*JWTClaims)

		rows, err := repo.GetDashboards(claims.AccountID)
		if err != nil {
			logError(logger, err)
			response.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		var list []DashboardResponse
		for _, r := range rows {
			list = append(list, DashboardResponse{
				ID:          r.ID,
				Name:        r.Name,
				Description: r.Description,
				Width:       r.Width,
				Height:      r.Height,
			})
		}

		response.Object(w, list, http.StatusOK)
	})
}

// DashboardWidgetResponse ...
type DashboardWidgetResponse struct {
	ID      int64  `json:"id"`
	Title   string `json:"title"`
	Type    string `json:"type"`
	PosX    int    `json:"posX"`
	PosY    int    `json:"posY"`
	Span    int    `json:"span"`
	DataURL string `json:"dataUrl"`
}

// GetDashboardWidgets ...
func GetDashboardWidgets(repo *dashboards.Repository, logger *log.Logger) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		claims := r.Context().Value("user").(*JWTClaims)

		dashboardIDArg := chi.URLParam(r, "id")
		if dashboardIDArg == "" {
			response.Error(w, "id parameter is required", http.StatusBadRequest)
			return
		}

		dashboardID, _ := strconv.ParseInt(dashboardIDArg, 0, 64)

		rows, err := repo.GetDashboardWidgets(claims.AccountID, dashboardID)
		if err != nil {
			logError(logger, err)
			response.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		var list []DashboardWidgetResponse
		for _, r := range rows {
			list = append(list, DashboardWidgetResponse{
				ID:      r.ID,
				Title:   r.Title,
				Type:    r.Type,
				PosX:    r.PosX,
				PosY:    r.PosY,
				Span:    r.Span,
				DataURL: r.DataURL,
			})
		}

		response.Object(w, list, http.StatusOK)
	})
}

// CreateDashboardForm ...
type CreateDashboardForm struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
}

// CreateDashboard ...
func CreateDashboard(repo *dashboards.Repository, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		claims := r.Context().Value("user").(*JWTClaims)

		var form CreateDashboardForm

		if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
			logError(logger, err)
			response.Error(w, "decode form error", http.StatusBadRequest)
			return
		}

		v := validator.New()
		if err := v.Struct(form); err != nil {
			response.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		dashboard := dashboards.Dashboard{
			Name:        form.Name,
			Description: form.Description,
			Width:       form.Width,
			Height:      form.Height,
		}

		if _, err := repo.CreateDashboard(claims.AccountID, dashboard); err != nil {
			logError(logger, err)
			response.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		response.Ok(w)
	})
}

// CreateDashboardWidgetForm ...
type CreateDashboardWidgetForm struct {
	Title   string `json:"title"`
	Type    string `json:"type"`
	PosX    int    `json:"posX"`
	PosY    int    `json:"posY"`
	Span    int    `json:"span"`
	DataURL string `json:"dataUrl"`
}

// AddDashboardWidget ...
func AddDashboardWidget(repo *dashboards.Repository, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		claims := r.Context().Value("user").(*JWTClaims)

		var form CreateDashboardWidgetForm

		if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
			logError(logger, err)
			response.Error(w, "decode form error", http.StatusBadRequest)
			return
		}

		v := validator.New()
		if err := v.Struct(form); err != nil {
			response.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		dashboardIDArg := chi.URLParam(r, "id")
		if dashboardIDArg == "" {
			response.Error(w, "id parameter is required", http.StatusBadRequest)
			return
		}

		dashboardID, _ := strconv.ParseInt(dashboardIDArg, 0, 64)

		widget := dashboards.DashboardWidget{
			Title:   form.Title,
			Type:    form.Type,
			PosX:    form.PosX,
			PosY:    form.PosY,
			Span:    form.Span,
			DataURL: form.DataURL,
		}

		if err := repo.AddWidget(claims.AccountID, dashboardID, widget); err != nil {
			logError(logger, err)
			response.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		response.Ok(w)
	})
}

// DeleteDashboard ...
func DeleteDashboard(repo *dashboards.Repository, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		claims := r.Context().Value("user").(*JWTClaims)

		dashboardIDArg := chi.URLParam(r, "id")
		if dashboardIDArg == "" {
			response.Error(w, "id parameter is required", http.StatusBadRequest)
			return
		}

		dashboardID, _ := strconv.ParseInt(dashboardIDArg, 0, 64)

		if err := repo.DeleteDashboard(claims.AccountID, dashboardID); err != nil {
			logError(logger, err)
			response.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		response.Ok(w)
	})
}

// DeleteDashboardWidget ...
func DeleteDashboardWidget(repo *dashboards.Repository, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		claims := r.Context().Value("user").(*JWTClaims)

		dashboardIDArg := chi.URLParam(r, "id")
		if dashboardIDArg == "" {
			response.Error(w, "id parameter is required", http.StatusBadRequest)
			return
		}

		dashboardID, _ := strconv.ParseInt(dashboardIDArg, 0, 64)

		dashboardWidgetIDArg := chi.URLParam(r, "widgetID")
		if dashboardWidgetIDArg == "" {
			response.Error(w, "widgetID parameter is required", http.StatusBadRequest)
			return
		}

		dashboardWidgetID, _ := strconv.ParseInt(dashboardWidgetIDArg, 0, 64)

		if err := repo.DeleteDashboardWidget(claims.AccountID, dashboardID, dashboardWidgetID); err != nil {
			logError(logger, err)
			response.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		response.Ok(w)
	})
}
