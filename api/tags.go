package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	validator "gopkg.in/go-playground/validator.v9"

	"shortly/api/response"

	"shortly/app/tags"
)

// AddTagForm string
type AddTagForm struct {
	LinkID int64  `json:"linkID" binding:"required"`
	Tag    string `json:"tag" binding:"required"`
}

// AddTagToLink handler for adding tag to link
// @Summary Adds tag to link
// @Tags Tags
// @Description create new tag for link
// @ID add-tag-to-link
// @Accept  json
// @Produce  json
// @Success 200 {object} response.ApiResponse
// @Failure 400 {object} response.ApiResponse
// @Failure 404 {object} response.ApiResponse
// @Failure 500 {object} response.ApiResponse
// @Router /tags/create [post]
func AddTagToLink(repo *tags.TagsRepository, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var form AddTagForm

		if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
			logError(logger, err)
			response.Error(w, "decode form error", http.StatusBadRequest)
			return
		}

		v := validator.New()
		if err := v.Struct(form); err != nil {
			response.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if _, err := repo.AddTagToLink(form.LinkID, form.Tag); err != nil {
			logError(logger, err)
			response.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		response.Ok(w)
	})
}

// DeleteTagFromLink ...
func DeleteTagFromLink(repo *tags.TagsRepository, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		tagNameArg := chi.URLParam(r, "tagName")
		if tagNameArg == "" {
			response.Error(w, "tag name parameter is required", http.StatusBadRequest)
			return
		}

		linkIDArg := chi.URLParam(r, "linkID")
		if linkIDArg == "" {
			response.Error(w, "url parameter is required", http.StatusBadRequest)
			return
		}

		linkID, err := strconv.ParseInt(linkIDArg, 0, 64)
		if err != nil {
			response.Error(w, "linkID is not a number", http.StatusBadRequest)
			return
		}

		if _, err := repo.DeleteTagFromLink(linkID, tagNameArg); err != nil {
			logError(logger, err)
			response.Error(w, "internal server error", http.StatusBadRequest)
			return
		}

		response.Ok(w)
	})
}
