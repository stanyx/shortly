package handlers

import (
	"log"
	"net/http"
	"encoding/json"

	"shortly/billing"
)

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
		w.Write([]byte("ok"))

	})
}