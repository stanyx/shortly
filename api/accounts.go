package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	validator "gopkg.in/go-playground/validator.v9"

	"shortly/api/response"

	"shortly/app/accounts"
	"shortly/app/billing"
	"shortly/app/rbac"
	"shortly/app/users"
)

type JWTClaims struct {
	UserID    int64  `json:"userId"`
	Name      string `json:"name"`
	Phone     string `json:"phone"`
	Email     string `json:"email"`
	IsStaff   bool   `json:"isStaff"`
	AccountID int64  `json:"accountId"`
	RoleID    int64  `json:"roleId"`
	jwt.StandardClaims
}

type AccountRegistrationForm struct {
	Password string `json:"password" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Company  string `json:"company" binding:"required"`
	Phone    string `json:"phone"`
}

type UserResponse struct {
	ID        int64  `json:"id,omitempty"`
	AccountID int64  `json:"account_id"`
	Username  string `json:"username,omitempty"`
	Email     string `json:"email"`
	Phone     string `json:"phone,omitempty"`
	Company   string `json:"company"`
}

type LoginResponse struct {
	User  UserResponse `json:"user"`
	Token string       `json:"token"`
}

func validateEmail(email string) error {
	var rxEmail = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	if rxEmail.MatchString(email) {
		return nil
	}
	return errors.New("invalid email address")
}

// RegisterAccount returns http handler that creates a new account binded to admin user
func RegisterAccount(accountRepo *accounts.AccountsRepository, userRepo *users.UsersRepository, billingRepo *billing.BillingRepository, billingLimiter *billing.BillingLimiter, logger *log.Logger) http.HandlerFunc {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var form AccountRegistrationForm

		if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
			response.Bad(w, "decode form error")
			return
		}

		v := validator.New()
		if err := v.Struct(&form); err != nil {
			response.Bad(w, err.Error())
			return
		}

		if err := validateEmail(form.Email); err != nil {
			response.Bad(w, err.Error())
			return
		}

		// TODO - password check

		user := users.User{
			Username: form.Email,
			Password: form.Password,
			Phone:    form.Phone,
			Email:    form.Email,
			Company:  form.Company,
		}

		tx, err := accountRepo.DB.Begin()
		if err != nil {
			logger.Println(fmt.Errorf("account registration, tx error: %v", err))
			response.InternalError(w, "transaction begin error")
		}

		userID, accountID, err := accountRepo.CreateAccount(tx, user)
		if err != nil {
			_ = tx.Rollback()
			logger.Println(err)
			response.InternalError(w, "create account error")
			return
		}

		billingAccount, err := billingRepo.AttachToDefaultBilling(tx, accountID, 1)
		if err != nil {
			_ = tx.Rollback()
			logger.Println(err)
			response.InternalError(w, "attach to billing error")
			return
		}

		if err := billingLimiter.UpdateAccount(accountID, *billingAccount); err != nil {
			_ = tx.Rollback()
			logger.Println(err)
			response.InternalError(w, "update billing account error")
			return
		}

		if err := billingRepo.CreateStripeCustomer(tx, accountID, form.Email); err != nil {
			_ = tx.Rollback()
			logger.Println(err)
			response.InternalError(w, "create stripe customer error")
			return
		}

		if err := tx.Commit(); err != nil {
			_ = tx.Rollback()
			logger.Println(fmt.Errorf("account registration, commit error: %v", err))
			response.InternalError(w, "transaction commit error")
			return
		}

		response.Object(w, &UserResponse{
			ID:        userID,
			AccountID: accountID,
			Email:     user.Email,
			Company:   user.Company,
		}, http.StatusOK)
	})
}

// BillingPlanOptionResponse ...
type BillingPlanOptionResponse struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Price string `json:"fee"`
}

// BillingOptionCounterResponse ...
type BillingOptionCounterResponse struct {
	Name  string `json:"name"`
	Value string `json:"value"`
	Limit string `json:"limit"`
}

// ProfileResponse ...
type ProfileResponse struct {
	Username             string                         `json:"username"`
	Company              string                         `json:"company"`
	RoleName             string                         `json:"roleName"`
	BillingPlan          string                         `json:"billingPlan"`
	BillingPlanFee       string                         `json:"billingPlanFee"`
	BillingPlanExpiredAt string                         `json:"billingPlanExpiredAt"`
	PlansAvailable       []BillingPlanOptionResponse    `json:"plansAvailable"`
	BillingUsage         []BillingOptionCounterResponse `json:"billingUsage"`
}

// GetProfile ...
func GetProfile(accountRepo *accounts.AccountsRepository, userRepo *users.UsersRepository, rbacRepo rbac.IRbacRepository, billingRepo *billing.BillingRepository, billingLimiter *billing.BillingLimiter, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		claims := r.Context().Value("user").(*JWTClaims)

		account, err := accountRepo.GetAccount(claims.AccountID)
		if err != nil {
			logError(logger, err)
			response.Error(w, "get account error", http.StatusInternalServerError)
			return
		}

		user, err := userRepo.GetUserByID(claims.UserID)
		if err != nil {
			logError(logger, err)
			response.Error(w, "get user error", http.StatusInternalServerError)
			return
		}

		var role rbac.Role

		if claims.IsStaff {
			if user.RoleID > 0 {
				userRole, err := rbacRepo.GetRole(user.RoleID)
				if err != nil {
					logError(logger, err)
					response.Error(w, "get role error", http.StatusInternalServerError)
					return
				}
				role = userRole
			} else {
				role.Name = "not assigned"
			}
		} else {
			role.Name = "account administator"
		}

		billingPlan, err := billingRepo.GetAccountBillingPlan(claims.AccountID)
		if err != nil {
			logError(logger, err)
			response.Error(w, "get billing plan error", http.StatusInternalServerError)
			return
		}

		billingPlanToUpgrade, err := billingRepo.GetBillingPlansToUpgrade(billingPlan.ID)
		if err != nil {
			logError(logger, err)
			response.Error(w, "get billing plans error", http.StatusInternalServerError)
			return
		}

		var plansOptionsResponse []BillingPlanOptionResponse
		for _, bp := range billingPlanToUpgrade {
			plansOptionsResponse = append(plansOptionsResponse, BillingPlanOptionResponse{
				ID:    bp.ID,
				Name:  bp.Name,
				Price: bp.Price,
			})
		}

		billingStat, err := billingLimiter.GetBillingStatistics(claims.AccountID, billingPlan.Start, billingPlan.End)
		if err != nil {
			logError(logger, err)
			response.Error(w, "get billing statistics error", http.StatusInternalServerError)
			return
		}

		var billingPlanUsageResponse []BillingOptionCounterResponse
		for _, bs := range billingStat {
			billingPlanUsageResponse = append(billingPlanUsageResponse, BillingOptionCounterResponse{
				Name:  bs.Name,
				Value: bs.CurrentValue,
				Limit: bs.Value,
			})
		}

		resp := ProfileResponse{
			Username:             claims.Name,
			Company:              account.Name,
			RoleName:             role.Name,
			BillingPlan:          billingPlan.Name,
			BillingPlanFee:       billingPlan.Price,
			BillingPlanExpiredAt: billingPlan.End.Format(time.RFC3339),
			PlansAvailable:       plansOptionsResponse,
			BillingUsage:         billingPlanUsageResponse,
		}

		response.Object(w, resp, http.StatusOK)
	})
}
