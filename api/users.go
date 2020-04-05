package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"shortly/api/response"
	"shortly/config"

	"shortly/app/accounts"
	"shortly/app/billing"
	"shortly/app/rbac"
	"shortly/app/users"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/go-chi/chi"
	"golang.org/x/crypto/bcrypt"
	validator "gopkg.in/go-playground/validator.v9"
)

func UsersRoutes(r chi.Router,
	auth func(rbac.Permission, http.Handler) http.HandlerFunc,
	permissions map[string]rbac.Permission,
	accountsRepository *accounts.AccountsRepository,
	userRepo *users.UsersRepository,
	billingRepo *billing.BillingRepository,
	rbacRepo rbac.IRbacRepository,
	billingLimiter *billing.BillingLimiter,
	appConfig config.ApplicationConfig,
	logger *log.Logger,
) {
	r.Get("/api/v1/users", auth(
		rbac.NewPermission("/api/v1/users", "read_users", "GET"),
		GetUsers(userRepo, logger),
	))
	r.Post("/api/v1/users/create", auth(
		rbac.NewPermission("/api/v1/users/create", "create_user", "POST"),
		AddUser(userRepo, logger),
	))
	r.Get("/api/v1/users/{id}", auth(
		rbac.NewPermission("/api/v1/users/{id}", "get_user", "GET"),
		GetUser(userRepo, logger),
	))
	r.Put("/api/v1/users/{id}", auth(
		rbac.NewPermission("/api/v1/users/{id}", "update_user", "PUT"),
		UpdateUser(userRepo, logger),
	))
	r.Put("/api/v1/users/{id}/password", auth(
		rbac.NewPermission("/api/v1/users/{id}/password", "update_user_password", "PUT"),
		UpdatePassword(userRepo, logger),
	))
	r.Post("/api/v1/login", Login(userRepo, logger, appConfig.Auth))
	r.Get("/api/v1/user", GetLoggedInUser(userRepo, logger, appConfig.Auth))
	r.Get("/api/v1/profile", auth(
		rbac.NewPermission("/api/v1/profile", "read_profile", "GET"),
		GetProfile(accountsRepository, userRepo, rbacRepo, billingRepo, billingLimiter, logger),
	))
}

// GetUsers request handler thats gets all users for current account
// @Summary Retrieve all users for current account
// @Tags Users
// @ID get-users
// @Produce json
// @Success 200 {object} response.ApiResponse
// @Failure 403 {object} response.ApiResponse
// @Failure 500 {object} response.ApiResponse
// @Router /users [get]
func GetUsers(repo *users.UsersRepository, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		claims := r.Context().Value("user").(*JWTClaims)

		rows, err := repo.GetAccountUsers(claims.AccountID)
		if err != nil {
			logError(logger, err)
			response.InternalError(w, "internal error")
			return
		}

		var list []UserResponse
		for _, r := range rows {
			list = append(list, UserResponse{
				ID:       r.ID,
				Username: r.Username,
				Email:    r.Email,
				Phone:    r.Phone,
			})
		}

		response.Object(w, list, http.StatusOK)
	})
}

func GetUser(repo *users.UsersRepository, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		claims := r.Context().Value("user").(*JWTClaims)

		idArg := chi.URLParam(r, "id")
		if idArg == "" {
			response.Bad(w, "id parameter is required")
			return
		}

		id, err := strconv.ParseInt(idArg, 0, 64)
		if err != nil {
			response.Bad(w, "id is not a number")
			return
		}

		if id <= 0 {
			response.Bad(w, "id value must be greater than zero")
			return
		}

		user, err := repo.GetOne(claims.AccountID, id)
		if err != nil {
			logError(logger, err)
			response.InternalError(w, "internal error")
			return
		}

		response.Object(w, UserResponse{
			ID:        user.ID,
			AccountID: claims.AccountID,
			Username:  user.Username,
			Email:     user.Email,
			Phone:     user.Phone,
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

// AddUser returns a http handler that creates new user binded to the current account
func AddUser(repo *users.UsersRepository, logger *log.Logger) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		claims := r.Context().Value("user").(*JWTClaims)

		var form UserRegistrationForm

		if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
			logger.Println(fmt.Sprintf("add user - decode form error: %w", err))
			response.Error(w, "decode form error", http.StatusInternalServerError)
			return
		}

		user := users.User{
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

type UserUpdateForm struct {
	Email   string `json:"email" binding:"required"`
	Company string `json:"company" binding:"required"`
	Phone   string `json:"phone" binding:"required"`
}

// UpdateUser request handler for updating a user information
// @Summary Update user information by user id
// @Tags Users
// @ID update-user
// @Accept json
// @Produce json
// @Success 200 {object} response.ApiResponse
// @Failure 403 {object} response.ApiResponse
// @Failure 500 {object} response.ApiResponse
// @Router /users/{id} [put]
func UpdateUser(repo *users.UsersRepository, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		claims := r.Context().Value("user").(*JWTClaims)

		var form UserUpdateForm

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

		idArg := chi.URLParam(r, "id")
		if idArg == "" {
			response.Bad(w, "id parameter is required")
			return
		}

		id, err := strconv.ParseInt(idArg, 0, 64)
		if err != nil {
			response.Bad(w, "id is not a number")
			return
		}

		if id <= 0 {
			response.Bad(w, "id value must be greater than zero")
			return
		}

		user, err := repo.GetOne(claims.AccountID, id)
		if err != nil {
			response.InternalError(w, "get user error")
			return
		}

		user.Email = form.Email
		user.Phone = form.Phone
		user.Company = form.Company

		if err := repo.UpdateUser(id, user); err != nil {
			response.InternalError(w, "update user error")
			return
		}

		response.Ok(w)
	})
}

type UpdatePasswordForm struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

// UpdatePassword request handler for updating a user password
// @Summary Update user password by user id
// @Tags Users
// @ID update-user
// @Accept json
// @Produce json
// @Success 200 {object} response.ApiResponse
// @Failure 403 {object} response.ApiResponse
// @Failure 500 {object} response.ApiResponse
// @Router /users/{id}/password [put]
func UpdatePassword(repo *users.UsersRepository, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		claims := r.Context().Value("user").(*JWTClaims)

		var form UpdatePasswordForm

		if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
			response.Bad(w, "decode form error")
			return
		}

		idArg := chi.URLParam(r, "id")
		if idArg == "" {
			response.Bad(w, "id parameter is required")
			return
		}

		id, err := strconv.ParseInt(idArg, 0, 64)
		if err != nil {
			response.Bad(w, "id is not a number")
			return
		}

		if id <= 0 {
			response.Bad(w, "id value must be greater than zero")
			return
		}

		if form.OldPassword == "" {
			response.Bad(w, "old password is required")
			return
		}

		if form.NewPassword == "" {
			response.Bad(w, "new password is required")
			return
		}

		user, err := repo.GetOne(claims.AccountID, id)
		if err != nil {
			response.InternalError(w, "get user error")
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(form.OldPassword))
		if err == bcrypt.ErrMismatchedHashAndPassword {
			response.Bad(w, "old password incorrect")
			return
		} else if err != nil {
			response.InternalError(w, "password check error")
			return
		}

		password, err := bcrypt.GenerateFromPassword([]byte(form.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			response.Bad(w, "password not generated")
			return
		}

		if err := repo.UpdatePassword(id, string(password)); err != nil {
			response.InternalError(w, "update password error")
			return
		}

		response.Ok(w)
	})
}

// LoginForm ...
type LoginForm struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Login ...
func Login(repo *users.UsersRepository, logger *log.Logger, authConfig config.JWTConfig) http.HandlerFunc {
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

// GetLoggedInUser returns http handler that
func GetLoggedInUser(repo *users.UsersRepository, logger *log.Logger, authConfig config.JWTConfig) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		claims, err := ParseToken(w, r, authConfig)
		if err != nil {
			response.Object(w, &UserResponse{}, http.StatusOK)
			return
		}

		user, err := repo.GetUserByID(claims.UserID)
		if err != nil {
			logger.Println(fmt.Sprintf("get logged user error: %w", err))
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
