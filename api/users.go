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
	UserID    int64   `json:"user_id"`
	Name      string  `json:"name"`
	Phone     string  `json:"phone"`
	Email     string  `json:"email"`
	jwt.StandardClaims
}

type RegistrationForm struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Company  string	`json:"company"`
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