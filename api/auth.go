package api

import (
	"fmt"
	"strings"
	"net/http"
	"context"

	jwt "github.com/dgrijalva/jwt-go"

	"shortly/config"
)

func AuthMiddleware(authConfig config.JWTConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			var header = r.Header.Get("x-access-token")

			if header == "" {
				header = r.FormValue("x-access-token");
			}

			header = strings.TrimSpace(header)

			if header == "" {
				apiError(w, "access token is missing", http.StatusForbidden)
				return
			}

			claims := &JWTClaims{}

			_, err := jwt.ParseWithClaims(header, claims, func(token *jwt.Token) (interface{}, error) {
				return []byte(authConfig.Secret), nil
			})

			if err != nil {
				apiError(w, err.Error(), http.StatusBadRequest)
				return
			}

			ctx := context.WithValue(r.Context(), "user", claims)

			fmt.Println("logged request", claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}