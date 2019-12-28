package api

import (
	"errors"
	"encoding/json"
	"log"
	"net/http"

	"shortly/app/billing"
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
	Price       string                  `json:"price"`
	Options     []BillingOptionResponse `json:"options"`
}

func ListBillingPlans(repo *billing.BillingRepository, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		plans, err := repo.GetAllBillingPlans()
		if err != nil {
			logError(logger, err)
			apiError(w, "list plan error", http.StatusInternalServerError)
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

		response(w, &list, http.StatusOK)

	})
}


type ApplyBillingPlanForm struct {
	PlanID int64 `json:"plan_id"`
}

func ApplyBillingPlan(repo *billing.BillingRepository, billingLimiter *billing.BillingLimiter, logger *log.Logger) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var form ApplyBillingPlanForm

		err := json.NewDecoder(r.Body).Decode(&form)
		if err != nil {
			logError(logger, err)
			apiError(w, "decode form error", http.StatusBadRequest)
			return
		}

		claims := r.Context().Value("user").(*JWTClaims)

		if err := repo.ApplyBillingPlan(claims.UserID, form.PlanID); err != nil {
			logError(logger, err)
			apiError(w, "apply plan error", http.StatusInternalServerError)
			return
		}

		options, err := repo.GetBillingPlanOptions(claims.UserID, form.PlanID)
		if err != nil {
			logError(logger, err)
			apiError(w, "get plan error", http.StatusInternalServerError)
			return
		}

		if err := billingLimiter.SetPlanOptions(claims.UserID, options); err != nil {
			logError(logger, err)
			apiError(w, "set plan error", http.StatusInternalServerError)
			return
		}

		ok(w)

	})
}

func BillingLimitMiddleware(optionName string, billingLimiter *billing.BillingLimiter, logger *log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			claims := r.Context().Value("user").(*JWTClaims)

			if err := billingLimiter.CheckLimits(optionName, claims.UserID); err == billing.LimitExceededError {
				logError(logger, errors.New("plan limit exceeded"))
				apiError(w, "plan limit exceeded", http.StatusBadRequest)
				return
			} else if err != nil {
				logError(logger, err)
				apiError(w, "internal server error", http.StatusInternalServerError)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}