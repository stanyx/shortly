package api

import (
	"log"
	"net/http"
	"shortly/app/billing"
	"shortly/app/data"
	"shortly/app/rbac"
	"time"

	"github.com/go-chi/chi"

	"shortly/api/response"

	"shortly/app/clicks"
)

// ClicksRoutes ...
func ClicksRoutes(r chi.Router, auth func(rbac.Permission, http.Handler) http.HandlerFunc, repository *clicks.Repository, historyDB *data.HistoryDB, billingLimiter *billing.BillingLimiter, logger *log.Logger) {

	r.Get("/api/v1/users/links/clicks/total", auth(
		rbac.NewPermission("/api/v1/users/links/clicks/total", "read_clicks_total", "GET"),
		GetTotalClicks(repository, logger),
	))

	r.Get("/api/v1/users/links/clicks/data", auth(
		rbac.NewPermission("/api/v1/users/links/clicks/data", "read_clicks_data", "GET"),
		GetDayClicksData(repository, logger),
	))

	r.Get("/api/v1/users/links/{link}/stat", auth(
		rbac.NewPermission("/api/v1/users/links/{link}/stat", "read_link_stat", "GET"),
		GetLinkStat(repository, historyDB, billingLimiter, logger),
	))

}

// GetTotalClicks ...
func GetTotalClicks(repo *clicks.Repository, logger *log.Logger) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		claims := r.Context().Value("user").(*JWTClaims)

		count, err := repo.GetTotalClicks(claims.AccountID)
		if err != nil {
			logError(logger, err)
			response.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		response.Object(w, count, http.StatusOK)
	})
}

// DataSetResponse ...
type DataSetResponse struct {
	Label string        `json:"label"`
	Fill  bool          `json:"fill"`
	Data  []interface{} `json:"data"`
}

// DataResponse ...
type DataResponse struct {
	Labels   []string          `json:"labels"`
	Datasets []DataSetResponse `json:"datasets"`
}

// GetDayClicksData ...
func GetDayClicksData(repo *clicks.Repository, logger *log.Logger) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		claims := r.Context().Value("user").(*JWTClaims)

		rows, err := repo.GetClicksData(claims.AccountID)
		if err != nil {
			logError(logger, err)
			response.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		resp := DataResponse{
			Datasets: []DataSetResponse{{Label: ""}},
		}

		c := time.Now()
		t := time.Date(c.Year(), c.Month(), c.Day(), 0, 0, 0, 0, time.UTC)

		for i := 0; i < 24; i++ {
			resp.Labels = append(resp.Labels, t.Add(time.Hour*time.Duration(i)).Format("15:04"))
		}

		for _, r := range rows {
			resp.Datasets[0].Data = append(resp.Datasets[0].Data, r.Count)
		}

		response.Object(w, resp, http.StatusOK)
	})
}

// LinkStatResponse ...
type LinkStatResponse struct {
	Clicks    DataResponse `json:"clicks"`
	Referrers DataResponse `json:"referrers"`
	Locations DataResponse `json:"locations"`
}

// GetLinkStat ...
func GetLinkStat(repo *clicks.Repository, historyDB *data.HistoryDB, billingLimiter *billing.BillingLimiter, logger *log.Logger) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		claims := r.Context().Value("user").(*JWTClaims)

		defaultDayLimit := int64(31)

		link := chi.URLParam(r, "link")

		c := time.Now()
		startTime := time.Date(c.Year(), c.Month(), 1, 0, 0, 0, 0, time.UTC)
		endTime := startTime.Add(time.Hour * 24 * time.Duration(defaultDayLimit))

		data, err := historyDB.GetClicksData(claims.AccountID, link, startTime, endTime, data.Limit(defaultDayLimit))
		if err != nil {
			logError(logger, err)
			response.Error(w, "(get link data) - internal error", http.StatusInternalServerError)
			return
		}

		resp := LinkStatResponse{
			Clicks: DataResponse{
				Datasets: []DataSetResponse{{Label: ""}},
			},
			Referrers: DataResponse{
				Labels:   []string{},
				Datasets: []DataSetResponse{{Label: "", Data: []interface{}{}}},
			},
			Locations: DataResponse{
				Labels:   []string{},
				Datasets: []DataSetResponse{{Label: "", Data: []interface{}{}}},
			},
		}

		clickData := make(map[int64]int64)
		for _, r := range data.Clicks {
			t := r.Time
			ts := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
			clickData[ts.Unix()] += r.Count
		}

		referers := make(map[string]int)
		location := make(map[string]int)

		for _, i := range data.Infos {
			for k, v := range i.Info.Referrers {
				referers[k] += v
			}
			for k, v := range i.Info.Locations {
				location[k] += v
			}
		}

		for k, v := range referers {
			resp.Referrers.Labels = append(resp.Referrers.Labels, k)
			resp.Referrers.Datasets[0].Data = append(resp.Referrers.Datasets[0].Data, v)
		}

		for k, v := range location {
			resp.Locations.Labels = append(resp.Locations.Labels, k)
			resp.Locations.Datasets[0].Data = append(resp.Locations.Datasets[0].Data, v)
		}

		for i := 0; i < int(defaultDayLimit); i++ {
			ts := startTime.Add(time.Hour * 24 * time.Duration(i))
			resp.Clicks.Datasets[0].Data = append(resp.Clicks.Datasets[0].Data, clickData[ts.Unix()])
			resp.Clicks.Labels = append(resp.Clicks.Labels, ts.Format("01-02"))
		}

		response.Object(w, resp, http.StatusOK)
	})
}
