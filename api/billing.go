package api

import (
	"encoding/json"
	"log"
	"net/http"

	"shortly/billing"
)

type BillingOptionResponse struct {
	ID          int64                   `json:"id"`
	Name        string                  `json:"name"`
	Description string                  `json:"description"`
	Value       string                  `json:"value"`
}

type BillingPlanResponse struct {
	ID          int64                   `json:"id"`
	Name        string                  `json:"name"`
	Description string                  `json:"description"`
	Price       string					`json:"price"`
	Options     []BillingOptionResponse `json:"options"`
}

func ListBillingPlans(repo *billing.BillingRepository, logger *log.Logger) {
	http.HandleFunc("/api/v1/billing/plans", func(w http.ResponseWriter, r *http.Request) {

		plans, err := repo.GetAllBillingPlans()
		if err != nil {
			logger.Println(err)
			http.Error(w, "list plan error", http.StatusInternalServerError)
			return
		}

		var list []BillingPlanResponse

		for _, p := range plans {

			var planOptions []BillingOptionResponse
			for _, opt := range p.Options {
				planOptions = append(planOptions, BillingOptionResponse{
					ID:          opt.ID,
					Name:        opt.Name,
					Description: opt.Description,
					Value:       opt.Value,
				})
			}

			list = append(list, BillingPlanResponse{
				ID:          p.ID,
				Name:        p.Name,
				Description: p.Description,
				Price:       p.Price,
				Options:     planOptions,
			})
		}

		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(&list); err != nil {
			http.Error(w, "list plan error", http.StatusInternalServerError)
		}

	})
}


type ApplyBillingPlanForm struct {
	UserID int64
	PlanID int64
}

func ApplyBillingPlan(repo *billing.BillingRepository, logger *log.Logger) {
	http.HandleFunc("/api/v1/billing/apply", func(w http.ResponseWriter, r *http.Request) {

		var form ApplyBillingPlanForm

		err := json.NewDecoder(r.Body).Decode(&form)
		if err != nil {
			http.Error(w, "decode form error", http.StatusBadRequest)
			return
		}

		if err := repo.ApplyBillingPlan(form.UserID, form.PlanID); err != nil {
			http.Error(w, "apply plan error", http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))

	})
}
