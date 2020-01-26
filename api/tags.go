package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"shortly/app/tags"

	"github.com/go-chi/chi"
	validator "gopkg.in/go-playground/validator.v9"
)

type AddTagForm struct {
	LinkID int64  `json:"linkID" binding:"required"`
	Tag    string `json:"tag" binding:"required"`
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

		v := validator.New()
		if err := v.Struct(form); err != nil {
			apiError(w, err.Error(), http.StatusBadRequest)
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

func DeleteTagFromLink(repo *tags.TagsRepository, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		tagNameArg := chi.URLParam(r, "tagName")
		if tagNameArg == "" {
			apiError(w, "tag name parameter is required", http.StatusBadRequest)
			return
		}

		linkIDArg := chi.URLParam(r, "linkID")
		if linkIDArg == "" {
			apiError(w, "url parameter is required", http.StatusBadRequest)
			return
		}

		linkID, err := strconv.ParseInt(linkIDArg, 0, 64)
		if err != nil {
			apiError(w, "linkID is not a number", http.StatusBadRequest)
			return
		}

		if _, err := repo.DeleteTagFromLink(linkID, tagNameArg); err != nil {
			logError(logger, err)
			apiError(w, "internal server error", http.StatusBadRequest)
			return
		}

		ok(w)
	})
}
