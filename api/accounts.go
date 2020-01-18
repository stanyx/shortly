package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
	validator "gopkg.in/go-playground/validator.v9"

	"shortly/app/accounts"
	"shortly/app/billing"
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
	Email     string `json:"email"`
	Company   string `json:"company"`
}

type LoginResponse struct {
	User  UserResponse `json:"user"`
	Token string       `json:"token"`
}

func RegisterAccount(repo *accounts.UsersRepository, billingRepo *billing.BillingRepository, billingLimiter *billing.BillingLimiter, logger *log.Logger) http.HandlerFunc {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Method != "POST" {
			apiError(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var form AccountRegistrationForm

		if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
			logger.Println(err)
			apiError(w, "decode form error", http.StatusInternalServerError)
			return
		}

		v := validator.New()
		if err := v.Struct(&form); err != nil {
			apiError(w, err.Error(), http.StatusBadRequest)
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

		userID, accountID, err := repo.CreateAccount(user)
		if err != nil {
			logger.Println(err)
			apiError(w, "create account error", http.StatusInternalServerError)
			return
		}

		billingAccount, err := billingRepo.AttachToDefaultBilling(accountID, 1)
		if err != nil {
			logger.Println(err)
			apiError(w, "attach to billing error", http.StatusInternalServerError)
			return
		}

		if err := billingLimiter.UpdateAccount(accountID, *billingAccount); err != nil {
			logger.Println(err)
			apiError(w, "update billing account error", http.StatusInternalServerError)
			return
		}

		response(w, &UserResponse{
			ID:        userID,
			AccountID: accountID,
			Email:     user.Email,
			Company:   user.Company,
		}, http.StatusOK)
	})
}

type UserRegistrationForm struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	Phone     string `json:"phone"`
	Email     string `json:"email"`
	AccountID int64  `json:"accountId"`
	RoleID    int64  `json:"roleId"`
}

func AddUser(repo *accounts.UsersRepository, logger *log.Logger) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Method != "POST" {
			apiError(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var form UserRegistrationForm

		if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
			logger.Println(err)
			apiError(w, "decode form error", http.StatusInternalServerError)
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

		userID, err := repo.CreateUser(form.AccountID, user)
		if err != nil {
			logger.Println(err)
			apiError(w, "save user error", http.StatusInternalServerError)
			return
		}

		response(w, &UserResponse{
			ID:      userID,
			Email:   user.Email,
			Company: user.Company,
		}, http.StatusOK)
	})
}

type LoginForm struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func Login(repo *accounts.UsersRepository, logger *log.Logger, authConfig config.JWTConfig) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Method != "POST" {
			apiError(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var form LoginForm

		if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
			logger.Println(err)
			apiError(w, "decode form error", http.StatusInternalServerError)
			return
		}

		user, err := repo.GetUserByEmail(form.Email)
		if err != nil {
			logger.Println(err)
			apiError(w, "get user error", http.StatusInternalServerError)
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(form.Password))
		if err != nil && err == bcrypt.ErrMismatchedHashAndPassword {
			logger.Println(err)
			apiError(w, "incorrect password", http.StatusBadRequest)
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
			apiError(w, "signing token error", http.StatusInternalServerError)
			return
		}

		response(w, &LoginResponse{
			User: UserResponse{
				ID:      user.ID,
				Email:   user.Email,
				Company: user.Company,
			},
			Token: tokenSigned,
		}, http.StatusOK)

	})
}

func GetLoggedInUser(repo *accounts.UsersRepository, logger *log.Logger, authConfig config.JWTConfig) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		claims, err := ParseToken(w, r, authConfig)
		if err != nil {
			response(w, &UserResponse{}, http.StatusOK)
			return
		}

		user, err := repo.GetUserByID(claims.UserID)
		if err != nil {
			apiError(w, "get user error", http.StatusInternalServerError)
			return
		}

		response(w, &UserResponse{
			ID:      user.ID,
			Email:   user.Email,
			Company: user.Company,
		}, http.StatusOK)

	})
}

type CreateGroupForm struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type GroupResponse struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func AddGroup(repo *accounts.UsersRepository, logger *log.Logger) http.HandlerFunc {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		claims := r.Context().Value("user").(*JWTClaims)
		accountID := claims.AccountID

		var form CreateGroupForm

		if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
			logError(logger, err)
			apiError(w, "decode form error", http.StatusBadRequest)
			return
		}

		groupID, err := repo.AddGroup(accounts.Group{
			AccountID:   accountID,
			Name:        form.Name,
			Description: form.Description,
		})

		if err != nil {
			logError(logger, err)
			apiError(w, "add group error", http.StatusInternalServerError)
			return
		}

		response(w, &GroupResponse{
			ID:          groupID,
			Name:        form.Name,
			Description: form.Description,
		}, http.StatusOK)
	})

}

type DeleteGroupForm struct {
	GroupID int64 `json:"groupId"`
}

func DeleteGroup(repo *accounts.UsersRepository, logger *log.Logger) http.HandlerFunc {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		claims := r.Context().Value("user").(*JWTClaims)
		accountID := claims.AccountID

		var form DeleteGroupForm

		if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
			logError(logger, err)
			apiError(w, "decode form error", http.StatusBadRequest)
			return
		}

		err := repo.DeleteGroup(form.GroupID, accountID)
		if err != nil {
			logError(logger, err)
			apiError(w, "delete group error", http.StatusInternalServerError)
			return
		}

		ok(w)
	})

}

type AddUserToGroupForm struct {
	GroupID int64 `json:"groupId"`
	UserID  int64 `json:"userId"`
}

func AddUserToGroup(repo *accounts.UsersRepository, logger *log.Logger) http.HandlerFunc {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		accountID := r.Context().Value("user").(*JWTClaims).AccountID

		var form AddUserToGroupForm

		if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
			logError(logger, err)
			apiError(w, "decode form error", http.StatusBadRequest)
			return
		}

		if _, err := repo.GetUserByAccountID(accountID); err != nil {
			logError(logger, err)
			apiError(w, "internal server error", http.StatusBadRequest)
			return
		}

		if err := repo.AddUserToGroup(form.GroupID, form.UserID); err != nil {
			logError(logger, err)
			apiError(w, "add user to group error", http.StatusInternalServerError)
			return
		}

		ok(w)
	})

}

type DeleteUserFromGroupForm struct {
	GroupID int64 `json:"groupId"`
	UserID  int64 `json:"userId"`
}

func DeleteUserFromGroup(repo *accounts.UsersRepository, logger *log.Logger) http.HandlerFunc {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		accountID := r.Context().Value("user").(*JWTClaims).AccountID

		var form DeleteUserFromGroupForm

		if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
			logError(logger, err)
			apiError(w, "decode form error", http.StatusBadRequest)
			return
		}

		if _, err := repo.GetUserByAccountID(accountID); err != nil {
			logError(logger, err)
			apiError(w, "internal server error", http.StatusBadRequest)
			return
		}

		if err := repo.DeleteUserFromGroup(form.GroupID, form.UserID); err != nil {
			logError(logger, err)
			apiError(w, "delete user from group error", http.StatusInternalServerError)
			return
		}

		ok(w)
	})

}
