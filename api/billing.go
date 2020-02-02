package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/charge"
	"github.com/stripe/stripe-go/sub"
	"github.com/stripe/stripe-go/webhook"

	"shortly/api/response"

	"shortly/app/billing"
	"shortly/config"
)

type BillingOptionResponse struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Value       string `json:"value"`
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
			response.Error(w, "list plan error", http.StatusInternalServerError)
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

		response.Object(w, &list, http.StatusOK)

	})
}

type ApplyBillingPlanForm struct {
	PlanID      int64  `json:"planID"`
	StripeToken string `json:"paymentToken"`
	FullName    string `json:"fullName"`
	Country     string `json:"country"`
	ZipCode     string `json:"zipCode"`
	IsAnnual    bool   `json:"isAnnual"`
}

func UpgradeBillingPlan(repo *billing.BillingRepository, billingLimiter *billing.BillingLimiter, paymentConfig config.PaymentConfig, logger *log.Logger) http.HandlerFunc {

	stripe.Key = paymentConfig.Key

	createPaymentCharge := func(accountID, planID int64, planCost string, token string, isSubscription bool) error {

		customerID, err := repo.GetStripeCustomer(accountID)
		if err != nil {
			return err
		}

		if isSubscription {

			stripeSubscriptionID, err := repo.GetStripeSubscriptionID(planID)
			if err != nil {
				return err
			}

			params := &stripe.SubscriptionParams{
				Customer: stripe.String(customerID),
				Items: []*stripe.SubscriptionItemsParams{
					{
						Plan: stripe.String(stripeSubscriptionID),
					},
				},
			}
			s, err := sub.New(params)
			if err != nil {
				return err
			}

			return repo.CreateStripeSubscription(accountID, planID, s)
		}

		price, _ := strconv.ParseInt(planCost, 0, 64)

		params := &stripe.ChargeParams{
			Customer:    stripe.String(customerID),
			Amount:      stripe.Int64(price),
			Currency:    stripe.String(string(stripe.CurrencyUSD)),
			Description: stripe.String("billing plan charge"),
		}

		_ = params.SetSource(token)

		chrg, err := charge.New(params)
		if err != nil {
			return err
		}
		return repo.CreateStripeCharge(accountID, planID, chrg)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		claims := r.Context().Value("user").(*JWTClaims)

		var form ApplyBillingPlanForm
		if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
			logError(logger, err)
			response.Error(w, "decode form error", http.StatusBadRequest)
			return
		}

		planCost, err := repo.GetBillingPlanCost(form.PlanID, form.IsAnnual)
		if err != nil {
			logError(logger, err)
			response.Error(w, "get billing plan error", http.StatusInternalServerError)
			return
		}
		l, _ := time.LoadLocation("UTC")
		tNow := time.Now().In(l)
		start := time.Date(tNow.Year(), tNow.Month(), tNow.Day(), 0, 0, 0, 0, time.UTC).In(l)
		end := start.Add(time.Hour * 24 * 30)

		if form.IsAnnual {
			end = start.Add(time.Hour * 24 * 365)
		}

		if err := createPaymentCharge(claims.AccountID, form.PlanID, planCost, form.StripeToken, !form.IsAnnual); err != nil {
			logError(logger, err)
			response.Error(w, "payment error", http.StatusInternalServerError)
			return
		}

		charge, _ := strconv.ParseInt(planCost, 0, 64)

		billingActivation := billing.BillingPlanActivation{
			PlanID:   form.PlanID,
			Start:    start,
			End:      end,
			Charge:   int(charge),
			IsAnnual: form.IsAnnual,
		}

		tx, err := repo.DB.Begin()
		if err != nil {
			logError(logger, fmt.Errorf("upgrade plan error, tx: %v", err))
			response.Error(w, "upgrade plan error", http.StatusInternalServerError)
			return
		}

		if err := repo.UpgradeBillingPlan(tx, claims.AccountID, billingActivation); err != nil {
			_ = tx.Rollback()
			logError(logger, err)
			response.Error(w, "upgrade plan error", http.StatusInternalServerError)
			return
		}

		options, err := repo.GetBillingPlanOptions(tx, claims.AccountID, form.PlanID)
		if err != nil {
			_ = tx.Rollback()
			logError(logger, err)
			response.Error(w, "get plan error", http.StatusInternalServerError)
			return
		}

		billingAccount := billing.BillingAccount{
			Start:   start,
			End:     end,
			Options: options,
		}

		if err := billingLimiter.UpdateAccount(claims.AccountID, billingAccount); err != nil {
			logError(logger, err)
			response.Error(w, "set plan error", http.StatusInternalServerError)
			return
		}

		if err := tx.Commit(); err != nil {
			_ = tx.Rollback()
			logError(logger, err)
			response.Error(w, "upgrade plan error", http.StatusInternalServerError)
		}

		response.Ok(w)

	})
}

type CancelSubscriptionForm struct {
	PlanID int64
}

func CancelSubscription(repo *billing.BillingRepository, billingLimiter *billing.BillingLimiter, logger *log.Logger) http.HandlerFunc {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var form CancelSubscriptionForm

		claims := r.Context().Value("user").(*JWTClaims)

		if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
			logError(logger, err)
			response.Error(w, "decode form error", http.StatusBadRequest)
			return
		}

		account, err := repo.GetAccountBillingPlan(claims.AccountID)
		if err != nil {
			logError(logger, err)
			response.Error(w, "get account billing plan error", http.StatusInternalServerError)
			return
		}

		if account.IsAnnual {
			response.Error(w, "annual plan cancelation is prohibited", http.StatusBadRequest)
			return
		}

		if err := repo.CancelSubscription(claims.AccountID); err != nil {
			response.Error(w, "annual plan cancelation is prohibited", http.StatusBadRequest)
			return
		}

		if err := billingLimiter.DowngradeToDefaultPlan(claims.AccountID); err != nil {
			logError(logger, err)
			response.Error(w, "set plan error", http.StatusInternalServerError)
			return
		}

		response.Ok(w)

	})
}

func StripeWebhook(repo *billing.BillingRepository, billingLimiter *billing.BillingLimiter, logger *log.Logger, webhookKey string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		const MaxBodyBytes = int64(65536)
		r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
		payload, err := ioutil.ReadAll(r.Body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading request body: %v\\n", err)
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		endpointSecret := webhookKey

		event, err := webhook.ConstructEvent(payload, r.Header.Get("Stripe-Signature"), endpointSecret)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error verifying webhook signature: %v\n", err)
			w.WriteHeader(http.StatusBadRequest) // Return a 400 error on a bad signature
			return
		}

		if err := json.Unmarshal(payload, &event); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to parse webhook body json: %v\\n", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var loggedEvent *stripe.Event
		var eventHandleError error

		switch event.Type {
		case "charge.expired":
			var chr stripe.Charge
			err := json.Unmarshal(event.Data.Raw, &chr)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error parsing webhook JSON: %v\\n", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			accountID, err := repo.GetAccountIDByStripeCharge(chr.ID)
			if err == nil {
				if err := billingLimiter.DowngradeToDefaultPlan(accountID); err != nil {
					logger.Println("charge.expired event error handling", err)
					eventHandleError = err
					// TODO - send error to queue for retries
				}
			} else {
				logger.Println("charge not found", err)
				eventHandleError = err
			}
			loggedEvent = &event
		case "charge.failed":
			var chr stripe.Charge
			err := json.Unmarshal(event.Data.Raw, &chr)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error parsing webhook JSON: %v\\n", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			accountID, err := repo.GetAccountIDByStripeCharge(chr.ID)
			if err == nil {
				if err := billingLimiter.DowngradeToDefaultPlan(accountID); err != nil {
					logger.Println("charge.expired event error handling", err)
					// TODO - send error to queue for retries
					eventHandleError = err
				}
			} else {
				logger.Println("charge not found", err)
			}
			loggedEvent = &event
		case "subscription.aborted":
			var s stripe.SubscriptionSchedule
			err := json.Unmarshal(event.Data.Raw, &s)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error parsing webhook JSON: %v\\n", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			accountID, err := repo.CancelSubscriptionExternal(s.Subscription.ID, s.CanceledAt)
			if err != nil {
				logger.Println("subscription_schedule.canceled event error handling", err)
				// TODO - send error to queue for retries
				eventHandleError = err
			} else {
				if err := billingLimiter.DowngradeToDefaultPlan(accountID); err != nil {
					logger.Println("subscription_schedule.canceled event error handling", err)
					// TODO - send error to queue for retries
					eventHandleError = err
				}
			}
			loggedEvent = &event
		case "subscription_schedule.canceled":
			var s stripe.SubscriptionSchedule
			err := json.Unmarshal(event.Data.Raw, &s)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error parsing webhook JSON: %v\\n", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			accountID, err := repo.CancelSubscriptionExternal(s.Subscription.ID, s.CanceledAt)
			if err != nil {
				logger.Println("subscription_schedule.canceled event error handling", err)
				// TODO - send error to queue for retries
				eventHandleError = err
			} else {
				if err := billingLimiter.DowngradeToDefaultPlan(accountID); err != nil {
					logger.Println("subscription_schedule.canceled event error handling", err)
					// TODO - send error to queue for retries
					eventHandleError = err
				}
			}
			loggedEvent = &event
		default:
			fmt.Fprintf(os.Stderr, "Unexpected event type: %s\\n", event.Type)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if loggedEvent != nil {
			_ = repo.CreateStripeEvent(loggedEvent.ID, string(event.Data.Raw), loggedEvent.Created, eventHandleError)
		}

		w.WriteHeader(http.StatusOK)
	})
}

func BillingLimitMiddleware(optionName string, billingLimiter *billing.BillingLimiter, logger *log.Logger) func(http.Handler) http.HandlerFunc {
	return func(next http.Handler) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			claims := r.Context().Value("user").(*JWTClaims)

			if err := billingLimiter.CheckLimits(optionName, claims.AccountID); err == billing.LimitExceededError {
				logError(logger, errors.New("plan limit exceeded"))
				response.Error(w, "plan limit exceeded", http.StatusBadRequest)
				return
			} else if err != nil {
				logError(logger, err)
				response.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
