package api

import (
	"log"
	"net/http"
	"shortly/app/billing"
	"shortly/app/data"
	"shortly/app/rbac"
	"time"

	"github.com/go-chi/chi"

	"shortly/app/clicks"
)

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

func GetTotalClicks(repo *clicks.Repository, logger *log.Logger) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		claims := r.Context().Value("user").(*JWTClaims)

		count, err := repo.GetTotalClicks(claims.AccountID)
		if err != nil {
			logError(logger, err)
			apiError(w, "internal error", http.StatusInternalServerError)
			return
		}

		response(w, count, http.StatusOK)
	})
}

type DataSetResponse struct {
	Label string        `json:"label"`
	Fill  bool          `json:"fill"`
	Data  []interface{} `json:"data"`
}

type DataResponse struct {
	Labels   []string          `json:"labels"`
	Datasets []DataSetResponse `json:"datasets"`
}

func GetDayClicksData(repo *clicks.Repository, logger *log.Logger) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		claims := r.Context().Value("user").(*JWTClaims)

		rows, err := repo.GetClicksData(claims.AccountID)
		if err != nil {
			logError(logger, err)
			apiError(w, "internal error", http.StatusInternalServerError)
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

		response(w, resp, http.StatusOK)
	})
}

type LinkStatResponse struct {
	Clicks    DataResponse `json:"clicks"`
	Referrers DataResponse `json:"referrers"`
	Locations DataResponse `json:"locations"`
}

func GetLinkStat(repo *clicks.Repository, historyDB *data.HistoryDB, billingLimiter *billing.BillingLimiter, logger *log.Logger) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		claims := r.Context().Value("user").(*JWTClaims)

		defaultDayLimit := int64(31)

		link := chi.URLParam(r, "link")

		c := time.Now()
		startTime := time.Date(c.Year(), c.Month(), 1, 0, 0, 0, 0, time.UTC)
		endTime := startTime.Add(time.Hour * 24 * time.Duration(defaultDayLimit))

		rows, err := historyDB.GetClicksData(claims.AccountID, link, startTime, endTime, data.Limit(defaultDayLimit))
		if err != nil {
			logError(logger, err)
			apiError(w, "(get link data) - internal error", http.StatusInternalServerError)
			return
		}

		resp := LinkStatResponse{
			Clicks: DataResponse{
				Datasets: []DataSetResponse{{Label: ""}},
			},
			Referrers: DataResponse{
				Labels:   []string{"Location"},
				Datasets: []DataSetResponse{{Label: "", Data: []interface{}{1}}},
			},
			Locations: DataResponse{
				Labels:   []string{"SMS/Email/Direct"},
				Datasets: []DataSetResponse{{Label: "", Data: []interface{}{1}}},
			},
		}

		clickData := make(map[int64]int64)
		for _, r := range rows {
			t := r.Time
			ts := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
			clickData[ts.Unix()] += r.Count
		}

		for i := 0; i < int(defaultDayLimit); i++ {
			ts := startTime.Add(time.Hour * 24 * time.Duration(i))
			resp.Clicks.Datasets[0].Data = append(resp.Clicks.Datasets[0].Data, clickData[ts.Unix()])
			resp.Clicks.Labels = append(resp.Clicks.Labels, ts.Format("01-02"))
		}

		response(w, resp, http.StatusOK)
	})
}
