package api

import (
	"fmt"
	"log"
	"net/http"
	"encoding/json"
	"time"

	"golang.org/x/crypto/bcrypt"
	jwt "github.com/dgrijalva/jwt-go"

	"shortly/config"
	"shortly/app/users"
)

type JWTClaims struct {
	UserID    int64   `json:"userId"`
	Name      string  `json:"name"`
	Phone     string  `json:"phone"`
	Email     string  `json:"email"`
	IsStaff   bool    `json:"isStaff"`
	AdminID   int64   `json:"adminId"`
	RoleID    int64	  `json:"roleId"`
	jwt.StandardClaims
}

type RegistrationForm struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Company  string	`json:"company"`
	AdminID  int64  `json:"adminId"`
	IsStaff  bool   `json:"isStaff"`
	RoleID   int64  `json:"roleId"`
}

type UserResponse struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Company  string	`json:"company"`
}

type LoginResponse struct {
	User  UserResponse `json:"user"`
	Token string       `json:"token"`
}

func RegisterUser(repo *users.UsersRepository, logger *log.Logger) {

	http.HandleFunc("/api/v1/users/registration", func(w http.ResponseWriter, r *http.Request) {

		if r.Method != "POST" {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var form RegistrationForm

		if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
			logger.Println(err)
			http.Error(w, "decode form error", http.StatusInternalServerError)
			return
		}

		user := users.User{
			Username: form.Username,
			Password: form.Password,
			Phone:    form.Phone,
			Email:    form.Email,
			Company:  form.Company,
			AdminID:  form.AdminID,
			IsStaff:  form.IsStaff,
			RoleID:   form.RoleID,
		}

		userID, err := repo.CreateUser(user)
		if err != nil {
			logger.Println(err)
			http.Error(w, "save user error", http.StatusInternalServerError)
			return			
		}

		response(w, &UserResponse{
			ID:       userID,
			Username: user.Username,
			Phone:    user.Phone,
			Email:    user.Email,
			Company:  user.Company,
		}, http.StatusOK)
	}) 
}

type LoginForm struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func LoginUser(repo *users.UsersRepository, logger *log.Logger, authConfig config.JWTConfig) {
	http.HandleFunc("/api/v1/login", func(w http.ResponseWriter, r *http.Request) {

		if r.Method != "POST" {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var form LoginForm

		if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
			logger.Println(err)
			http.Error(w, "decode form error", http.StatusInternalServerError)
			return
		}

		user, err := repo.GetUser(form.Username)
		if err != nil {
			logger.Println(err)
			http.Error(w, "get user error", http.StatusInternalServerError)
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(form.Password))
		if err != nil && err == bcrypt.ErrMismatchedHashAndPassword {
			logger.Println(err)
			http.Error(w, "incorrect password", http.StatusBadRequest)
			return
		}

		claims := &JWTClaims{
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: int64(time.Duration(time.Millisecond * 1000 * 3600)),
				Issuer:    fmt.Sprintf("%v", user.ID),
			},
			Name:      user.Username,
			Email:     user.Email,
			Phone:     user.Phone,
			UserID:    user.ID,
			IsStaff:   user.IsStaff,
			AdminID:   user.AdminID,
			RoleID:    user.RoleID,
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

		tokenSigned, err := token.SignedString([]byte(authConfig.Secret))
		if err != nil {
			logger.Println(err)
			http.Error(w, "signing token error", http.StatusInternalServerError)
			return
		}

		response(w, &LoginResponse{
			User: UserResponse{
				Username: user.Username,
				Phone:    user.Phone,
				Email:    user.Email,
				Company:  user.Company,
			},
			Token: tokenSigned,
		}, http.StatusOK)

	})
}

type CreateGroupForm struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type GroupResponse struct {
	ID       int64     `json:"id"`
	Name 	 string    `json:"name"`
	Description string `json:"description"`
}

func AddGroup(repo *users.UsersRepository, logger *log.Logger) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		claims := r.Context().Value("user").(*JWTClaims)
		userID := claims.AdminID
		if userID == 0 {
			userID = claims.UserID
		}

		var form CreateGroupForm

		if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
			logError(logger, err)
			apiError(w, "decode form error", http.StatusBadRequest)
			return
		}

		groupID, err := repo.AddGroup(users.Group{
			UserID:      userID,
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

func DeleteGroup(repo *users.UsersRepository, logger *log.Logger) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		claims := r.Context().Value("user").(*JWTClaims)
		userID := claims.AdminID
		if userID == 0 {
			userID = claims.UserID
		}

		var form DeleteGroupForm

		if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
			logError(logger, err)
			apiError(w, "decode form error", http.StatusBadRequest)
			return
		}

		err := repo.DeleteGroup(form.GroupID, userID)
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

func AddUserToGroup(repo *users.UsersRepository, logger *log.Logger) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var form AddUserToGroupForm

		if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
			logError(logger, err)
			apiError(w, "decode form error", http.StatusBadRequest)
			return
		}

		// TODO - check user by admin_id

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

func DeleteUserFromGroup(repo *users.UsersRepository, logger *log.Logger) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var form DeleteUserFromGroupForm

		if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
			logError(logger, err)
			apiError(w, "decode form error", http.StatusBadRequest)
			return
		}

		// TODO - check user by admin_id

		if err := repo.DeleteUserFromGroup(form.GroupID, form.UserID); err != nil {
			logError(logger, err)
			apiError(w, "delete user from group error", http.StatusInternalServerError)
			return
		}

		ok(w)
	})

}