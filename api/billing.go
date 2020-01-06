package api

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/charge"

	"shortly/config"
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

func ListBillingPlans(repo *billing.BillingRepository, logger *log.Logger) http.HandlerFunc {
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
	PlanID int64       `json:"plan_id"`
	StripeToken string `json:"paymentToken"`
}

func ApplyBillingPlan(repo *billing.BillingRepository, billingLimiter *billing.BillingLimiter, paymentConfig config.PaymentConfig, logger *log.Logger) http.HandlerFunc {

	createPaymentCharge := func(planCost string, r *http.Request) error {
		stripe.Key = paymentConfig.Key

		token := r.FormValue("stripeToken")

		price, _ := strconv.ParseInt(planCost, 0, 64)

		params := &stripe.ChargeParams{
			Amount: stripe.Int64(price),
			Currency: stripe.String(string(stripe.CurrencyUSD)),
			Description: stripe.String("billing plan charge"),
		}

		_ = params.SetSource(token)

		_, err := charge.New(params)
		return err
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		form := &ApplyBillingPlanForm{}

		planID, err := strconv.ParseInt(r.FormValue("planId"), 0, 64)
		if err != nil {
			logError(logger, err)
			apiError(w, "plan id not specified", http.StatusBadRequest)
			return
		}

		form.PlanID = planID

		claims := r.Context().Value("user").(*JWTClaims)

		planCost, err := repo.GetBillingPlanCost(form.PlanID)
		if err != nil {
			logError(logger, err)
			apiError(w, "get billing plan error", http.StatusInternalServerError)
			return
		}

		if err := createPaymentCharge(planCost, r); err != nil {
			logError(logger, err)
			apiError(w, "payment error", http.StatusInternalServerError)
			return
		}

		if err := repo.ApplyBillingPlan(claims.AccountID, form.PlanID); err != nil {
			logError(logger, err)
			apiError(w, "apply plan error", http.StatusInternalServerError)
			return
		}

		options, err := repo.GetBillingPlanOptions(claims.AccountID, form.PlanID)
		if err != nil {
			logError(logger, err)
			apiError(w, "get plan error", http.StatusInternalServerError)
			return
		}

		if err := billingLimiter.SetPlanOptions(claims.AccountID, options); err != nil {
			logError(logger, err)
			apiError(w, "set plan error", http.StatusInternalServerError)
			return
		}

		ok(w)

	})
}

func BillingLimitMiddleware(optionName string, billingLimiter *billing.BillingLimiter, logger *log.Logger) func(http.Handler) http.HandlerFunc {
	return func(next http.Handler) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			claims := r.Context().Value("user").(*JWTClaims)

			if err := billingLimiter.CheckLimits(optionName, claims.AccountID); err == billing.LimitExceededError {
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