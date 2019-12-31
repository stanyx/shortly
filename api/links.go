package api

import (
	"log"
	"encoding/json"
	"net/http"

	"shortly/app/urls"
)

type AddUrlToGroupForm struct {
	GroupID int64 `json:"groupId"`
	UrlID  int64  `json:"urlId"`
}

func AddUrlToGroup(repo *urls.UrlsRepository, logger *log.Logger) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var form AddUrlToGroupForm

		if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
			logError(logger, err)
			apiError(w, "decode form error", http.StatusBadRequest)
			return
		}

		// TODO - check url id

		// TODO - check group by user_id

		if err := repo.AddUrlToGroup(form.GroupID, form.UrlID); err != nil {
			logError(logger, err)
			apiError(w, "decode form error", http.StatusBadRequest)
			return
		}

		ok(w)
	})

}

type DeleteUrlFromGroupForm struct {
	GroupID int64 `json:"groupId"`
	UrlID  int64  `json:"urlId"`
}

func DeleteUrlFromGroup(repo *urls.UrlsRepository, logger *log.Logger) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var form DeleteUrlFromGroupForm

		if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
			logError(logger, err)
			apiError(w, "decode form error", http.StatusBadRequest)
			return
		}

		// TODO - check url id

		// TODO - check group by user_id

		if err := repo.DeleteUrlFromGroup(form.GroupID, form.UrlID); err != nil {
			logError(logger, err)
			apiError(w, "decode form error", http.StatusBadRequest)
			return
		}

		ok(w)
	})

}