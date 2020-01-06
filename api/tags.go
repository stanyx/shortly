package api

import (
	"log"
	"net/http"
	"encoding/json"

	"shortly/app/tags"
)

type AddTagForm struct {
	LinkID int64  `json:"linkId"`
	Tag    string `json:"tag"`
}

// AddTagToLink handler for adding tag to link
// @Summary Adds tag to link
// @Description create new tag for link
// @ID add-tag-to-link
// @Accept  json
// @Produce  json
// @Param id path int true "Account ID"
// @Success 200 {object} OkResponse
// @Failure 400 {object} apiError
// @Failure 404 {object} apiError
// @Failure 500 {object} apiError
// @Router /api/v1/tags/create [post]
func AddTagToLink(repo *tags.TagsRepository, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var form AddTagForm

		if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
			logError(logger, err)
			apiError(w, "decode form error", http.StatusBadRequest)
			return
		}

		if _, err := repo.AddTagToLink(form.LinkID, form.Tag); err != nil {
			logError(logger, err)
			apiError(w, "internal server error", http.StatusInternalServerError)
			return
		}

		ok(w)
	})
}

type DeleteTagForm struct {
	TagID int64 `json:"tagId"`
}

func DeleteTagFromLink(repo *tags.TagsRepository, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var form DeleteTagForm

		if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
			logError(logger, err)
			apiError(w, "decode form error", http.StatusBadRequest)
			return
		}

		if _, err := repo.DeleteTagFromLink(form.TagID); err != nil {
			logError(logger, err)
			apiError(w, "internal server error", http.StatusBadRequest)
			return
		}

		ok(w)
	})
}