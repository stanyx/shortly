package api

import (
	"fmt"
	"strings"
	"net/http"
	"context"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/casbin/casbin/v2"

	"shortly/config"
	"shortly/app/rbac"
)

func AuthMiddleware(enforcer *casbin.Enforcer, authConfig config.JWTConfig, permissions map[string]rbac.Permission) func(rbac.Permission, http.Handler) http.Handler {
	return func(permission rbac.Permission, next http.Handler) http.Handler {

		permissions[permission.Url] = permission

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

			// authorization
			// admin users by default has all previlegies
			if claims.IsStaff {
				ok, err := enforcer.Enforce(fmt.Sprintf("role:%v", claims.RoleID), permission.Url, permission.Method)
				if err != nil {
					apiError(w, err.Error(), http.StatusInternalServerError)
					return
				}
				if !ok {
					apiError(w, "access denied", http.StatusForbidden)
					return
				}
			}
			
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}