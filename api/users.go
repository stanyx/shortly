package api

import (
	"log"
	"net/http"

	"shortly/api/response"

	"shortly/app/accounts"
)

// GetUsers ...
// @Summary Retrieve all users for current account
// @Tags Users
// @ID get-users
// @Produce json
// @Success 200 {object} response.ApiResponse
// @Failure 403 {object} response.ApiResponse
// @Failure 500 {object} response.ApiResponse
// @Router /users [get]
func GetUsers(repo *accounts.UsersRepository, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		claims := r.Context().Value("user").(*JWTClaims)

		rows, err := repo.GetAccountUsers(claims.AccountID)
		if err != nil {
			logError(logger, err)
			response.Error(w, "internal error", http.StatusInternalServerError)
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
