package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
)

func getAccountID(r *http.Request) int64 {
	return r.Context().Value("user").(*JWTClaims).AccountID
}

func getIntParam(r *http.Request, param string) (int64, error) {
	argStr := chi.URLParam(r, param)
	if argStr == "" {
		return 0, fmt.Errorf("%s is a required argument", param)
	}
	argInt, err := strconv.ParseInt(argStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%s is not a number", param)
	}
	return argInt, nil
}
