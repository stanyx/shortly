package api

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/casbin/casbin/v2"
	jwt "github.com/dgrijalva/jwt-go"

	"shortly/api/response"
	"shortly/app/rbac"
	"shortly/config"
)

// ParseToken ...
func ParseToken(w http.ResponseWriter, r *http.Request, authConfig config.JWTConfig) (*JWTClaims, error) {
	var header = r.Header.Get("x-access-token")

	if header == "" {
		header = r.FormValue("x-access-token")
	}

	header = strings.TrimSpace(header)

	claims := &JWTClaims{}
	if header == "" {
		return claims, nil
	}

	_, err := jwt.ParseWithClaims(header, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(authConfig.Secret), nil
	})

	return claims, err
}

// AuthMiddleware ...
func AuthMiddleware(enforcer *casbin.Enforcer, authConfig config.JWTConfig, permissions map[string]rbac.Permission) func(rbac.Permission, http.Handler) http.HandlerFunc {
	return func(permission rbac.Permission, next http.Handler) http.HandlerFunc {

		permissions[permission.Url] = permission

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			var header = r.Header.Get("x-access-token")

			if header == "" {
				header = r.FormValue("x-access-token")
			}

			header = strings.TrimSpace(header)

			if header == "" {
				response.Error(w, "access token is missing", http.StatusForbidden)
				return
			}

			claims := &JWTClaims{}

			_, err := jwt.ParseWithClaims(header, claims, func(token *jwt.Token) (interface{}, error) {
				return []byte(authConfig.Secret), nil
			})

			if err != nil {
				response.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			ctx := context.WithValue(r.Context(), "user", claims)

			// authorization
			// admin users by default has all previlegies
			if claims.IsStaff {
				ok, err := enforcer.Enforce(fmt.Sprintf("role:%v", claims.RoleID), permission.Url, permission.Method)
				if err != nil {
					response.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				if !ok {
					response.Error(w, "access denied", http.StatusForbidden)
					return
				}
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
