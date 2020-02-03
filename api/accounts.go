package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/go-chi/chi"
	"golang.org/x/crypto/bcrypt"
	validator "gopkg.in/go-playground/validator.v9"

	"shortly/api/response"

	"shortly/app/accounts"
	"shortly/app/billing"
	"shortly/app/rbac"
	"shortly/config"
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

// RegisterAccount ...
func RegisterAccount(repo *accounts.UsersRepository, billingRepo *billing.BillingRepository, billingLimiter *billing.BillingLimiter, logger *log.Logger) http.HandlerFunc {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var form AccountRegistrationForm

		if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
			logger.Println(err)
			response.Error(w, "decode form error", http.StatusInternalServerError)
			return
		}

		v := validator.New()
		if err := v.Struct(&form); err != nil {
			response.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := validateEmail(form.Email); err != nil {
			response.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// TODO - password check

		user := accounts.User{
			Username: form.Email,
			Password: form.Password,
			Phone:    form.Phone,
			Email:    form.Email,
			Company:  form.Company,
		}

		tx, err := repo.DB.Begin()
		if err != nil {
			logger.Println(fmt.Errorf("account registration, tx error: %v", err))
			response.Error(w, "error", http.StatusInternalServerError)
		}

		userID, accountID, err := repo.CreateAccount(tx, user)
		if err != nil {
			_ = tx.Rollback()
			logger.Println(err)
			response.Error(w, "create account error", http.StatusInternalServerError)
			return
		}

		billingAccount, err := billingRepo.AttachToDefaultBilling(tx, accountID, 1)
		if err != nil {
			_ = tx.Rollback()
			logger.Println(err)
			response.Error(w, "attach to billing error", http.StatusInternalServerError)
			return
		}

		if err := billingLimiter.UpdateAccount(accountID, *billingAccount); err != nil {
			_ = tx.Rollback()
			logger.Println(err)
			response.Error(w, "update billing account error", http.StatusInternalServerError)
			return
		}

		if err := billingRepo.CreateStripeCustomer(tx, accountID, form.Email); err != nil {
			_ = tx.Rollback()
			logger.Println(err)
			response.Error(w, "create stripe customer error", http.StatusInternalServerError)
			return
		}

		if err := tx.Commit(); err != nil {
			_ = tx.Rollback()
			logger.Println(fmt.Errorf("account registration, commit error: %v", err))
			response.Error(w, "error", http.StatusInternalServerError)
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

// UserRegistrationForm ...
type UserRegistrationForm struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	RoleID   int64  `json:"roleId"`
}

// AddUser ...
func AddUser(repo *accounts.UsersRepository, logger *log.Logger) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		claims := r.Context().Value("user").(*JWTClaims)

		var form UserRegistrationForm

		if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
			logger.Println(err)
			response.Error(w, "decode form error", http.StatusInternalServerError)
			return
		}

		user := accounts.User{
			IsStaff:  true,
			Username: form.Username,
			Password: form.Password,
			Phone:    form.Phone,
			Email:    form.Email,
			RoleID:   form.RoleID,
		}

		userID, err := repo.CreateUser(claims.AccountID, user)
		if err != nil {
			logger.Println(err)
			response.Error(w, "save user error", http.StatusInternalServerError)
			return
		}

		response.Object(w, &UserResponse{
			ID:      userID,
			Email:   user.Email,
			Company: user.Company,
		}, http.StatusOK)
	})
}

// LoginForm ...
type LoginForm struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Login ...
func Login(repo *accounts.UsersRepository, logger *log.Logger, authConfig config.JWTConfig) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Method != "POST" {
			response.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var form LoginForm

		if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
			logger.Println(err)
			response.Error(w, "decode form error", http.StatusInternalServerError)
			return
		}

		user, err := repo.GetUserByEmail(form.Email)
		if err != nil {
			logger.Println(err)
			response.Error(w, "get user error", http.StatusInternalServerError)
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(form.Password))
		if err != nil && err == bcrypt.ErrMismatchedHashAndPassword {
			logger.Println(err)
			response.Error(w, "incorrect password", http.StatusBadRequest)
			return
		}

		claims := &JWTClaims{
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: int64(time.Duration(time.Millisecond * 1000 * 3600)),
				Issuer:    fmt.Sprintf("%v", user.ID),
			},
			UserID:    user.ID,
			Name:      user.Username,
			Email:     user.Email,
			Phone:     user.Phone,
			IsStaff:   user.IsStaff,
			AccountID: user.AccountID,
			RoleID:    user.RoleID,
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

		tokenSigned, err := token.SignedString([]byte(authConfig.Secret))
		if err != nil {
			logger.Println(err)
			response.Error(w, "signing token error", http.StatusInternalServerError)
			return
		}

		response.Object(w, &LoginResponse{
			User: UserResponse{
				ID:      user.ID,
				Email:   user.Email,
				Company: user.Company,
			},
			Token: tokenSigned,
		}, http.StatusOK)

	})
}

// GetLoggedInUser ...
func GetLoggedInUser(repo *accounts.UsersRepository, logger *log.Logger, authConfig config.JWTConfig) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		claims, err := ParseToken(w, r, authConfig)
		if err != nil {
			response.Object(w, &UserResponse{}, http.StatusOK)
			return
		}

		user, err := repo.GetUserByID(claims.UserID)
		if err != nil {
			response.Error(w, "get user error", http.StatusInternalServerError)
			return
		}

		response.Object(w, &UserResponse{
			ID:      user.ID,
			Email:   user.Email,
			Company: user.Company,
		}, http.StatusOK)

	})
}

// GetGroups ...
func GetGroups(repo *accounts.UsersRepository, logger *log.Logger) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		claims := r.Context().Value("user").(*JWTClaims)

		rows, err := repo.GetAccountGroups(claims.AccountID)
		if err != nil {
			logError(logger, err)
			response.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		var list []GroupResponse
		for _, r := range rows {
			list = append(list, GroupResponse{
				ID:          r.ID,
				Name:        r.Name,
				Description: r.Description,
			})
		}

		response.Object(w, list, http.StatusOK)
	})
}

// CreateGroupForm ...
type CreateGroupForm struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// GroupResponse ...
type GroupResponse struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// AddGroup ...
func AddGroup(repo *accounts.UsersRepository, logger *log.Logger) http.HandlerFunc {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		claims := r.Context().Value("user").(*JWTClaims)
		accountID := claims.AccountID

		var form CreateGroupForm

		if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
			logError(logger, err)
			response.Error(w, "decode form error", http.StatusBadRequest)
			return
		}

		groupID, err := repo.AddGroup(accounts.Group{
			AccountID:   accountID,
			Name:        form.Name,
			Description: form.Description,
		})

		if err != nil {
			logError(logger, err)
			response.Error(w, "add group error", http.StatusInternalServerError)
			return
		}

		response.Object(w, &GroupResponse{
			ID:          groupID,
			Name:        form.Name,
			Description: form.Description,
		}, http.StatusOK)
	})

}

// DeleteGroup ...
func DeleteGroup(repo *accounts.UsersRepository, logger *log.Logger) http.HandlerFunc {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		groupIDArg := chi.URLParam(r, "groupID")
		if groupIDArg == "" {
			response.Error(w, "groupID is required argument", http.StatusBadRequest)
			return
		}

		groupID, err := strconv.ParseInt(groupIDArg, 0, 64)
		if err != nil {
			response.Error(w, "groupID is not a number", http.StatusBadRequest)
			return
		}

		claims := r.Context().Value("user").(*JWTClaims)
		accountID := claims.AccountID

		err = repo.DeleteGroup(groupID, accountID)
		if err != nil {
			logError(logger, err)
			response.Error(w, "delete group error", http.StatusInternalServerError)
			return
		}

		response.Ok(w)
	})

}

// AddUserToGroupForm ...
type AddUserToGroupForm struct {
	GroupID int64 `json:"groupId"`
	UserID  int64 `json:"userId"`
}

// AddUserToGroup ...
func AddUserToGroup(repo *accounts.UsersRepository, logger *log.Logger) http.HandlerFunc {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		accountID := r.Context().Value("user").(*JWTClaims).AccountID

		var form AddUserToGroupForm

		if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
			logError(logger, err)
			response.Error(w, "decode form error", http.StatusBadRequest)
			return
		}

		if _, err := repo.GetUserByAccountID(accountID); err != nil {
			logError(logger, err)
			response.Error(w, "internal server error", http.StatusBadRequest)
			return
		}

		if err := repo.AddUserToGroup(form.GroupID, form.UserID); err != nil {
			logError(logger, err)
			response.Error(w, "add user to group error", http.StatusInternalServerError)
			return
		}

		response.Ok(w)
	})

}

// DeleteUserFromGroupForm ...
type DeleteUserFromGroupForm struct {
	GroupID int64 `json:"groupId"`
	UserID  int64 `json:"userId"`
}

// DeleteUserFromGroup ...
func DeleteUserFromGroup(repo *accounts.UsersRepository, logger *log.Logger) http.HandlerFunc {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		accountID := r.Context().Value("user").(*JWTClaims).AccountID

		var form DeleteUserFromGroupForm

		if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
			logError(logger, err)
			response.Error(w, "decode form error", http.StatusBadRequest)
			return
		}

		if _, err := repo.GetUserByAccountID(accountID); err != nil {
			logError(logger, err)
			response.Error(w, "internal server error", http.StatusBadRequest)
			return
		}

		if err := repo.DeleteUserFromGroup(form.GroupID, form.UserID); err != nil {
			logError(logger, err)
			response.Error(w, "delete user from group error", http.StatusInternalServerError)
			return
		}

		response.Ok(w)
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
func GetProfile(repo *accounts.UsersRepository, rbacRepo rbac.IRbacRepository, billingRepo *billing.BillingRepository, billingLimiter *billing.BillingLimiter, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		claims := r.Context().Value("user").(*JWTClaims)

		account, err := repo.GetAccount(claims.AccountID)
		if err != nil {
			logError(logger, err)
			response.Error(w, "get account error", http.StatusInternalServerError)
			return
		}

		user, err := repo.GetUserByID(claims.UserID)
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
