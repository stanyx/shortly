package api

import (
	"log"
	"net/http"

	"shortly/app/accounts"
)

func GetUsers(repo *accounts.UsersRepository, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		claims := r.Context().Value("user").(*JWTClaims)

		rows, err := repo.GetAccountUsers(claims.AccountID)
		if err != nil {
			logError(logger, err)
			apiError(w, "internal error", http.StatusInternalServerError)
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

		response(w, list, http.StatusOK)
	})
}