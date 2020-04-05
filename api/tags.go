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

// AddTagForm ...
type AddTagForm struct {
	LinkID int64  `json:"linkID" binding:"required"`
	Tag    string `json:"tag" binding:"required"`
}

type TagCreatedResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// AddTagToLink handler for adding tag to link
// @Summary Adds tag to link
// @Tags Tags
// @Description create new tag for link
// @ID add-tag-to-link
// @Accept json
// @Produce json
// @Success 201 {object} response.ApiResponse
// @Failure 400 {object} response.ApiResponse
// @Failure 404 {object} response.ApiResponse
// @Failure 500 {object} response.ApiResponse
// @Router /tags/create [post]
func AddTagToLink(repo tags.Repository, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var form AddTagForm

		if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
			response.Bad(w, "decode form error")
			return
		}

		v := validator.New()
		if err := v.Struct(form); err != nil {
			response.Bad(w, err.Error())
			return
		}

		tag := &tags.Tag{
			Name: form.Tag,
		}

		if _, err := repo.AddTagToLink(form.LinkID, tag); err != nil {
			logError(logger, err)
			response.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		response.Created(w, TagCreatedResponse{
			ID:   tag.ID,
			Name: tag.Name,
		})
	})
}

// UpdateTagForm ...
type UpdateTagForm struct {
	LinkID int64  `json:"linkID" binding:"required"`
	Name   string `json:"name" binding:"required"`
}

// UpdateTagName handler for updating tag name
// @Summary Updates tag name
// @Tags Tags
// @Description updates tag name
// @ID update-tag-name
// @Accept json
// @Produce json
// @Success 200 {object} response.ApiResponse
// @Failure 400 {object} response.ApiResponse
// @Failure 404 {object} response.ApiResponse
// @Failure 500 {object} response.ApiResponse
// @Router /tags/{tagName} [put]
func UpdateTagName(repo tags.Repository, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tagNameArg := chi.URLParam(r, "tagName")
		if tagNameArg == "" {
			response.Bad(w, "tag name parameter is required")
			return
		}

		var form UpdateTagForm

		if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
			response.Bad(w, "decode form error")
			return
		}

		if err := repo.UpdateTagName(form.LinkID, tagNameArg, form.Name); err != nil {
			logError(logger, err)
			response.InternalError(w, "internal server error")
			return
		}

		response.Ok(w)
	})
}

// DeleteTagFromLink http handler for deleting tag from a short link
// @Summary Delete tag from link
// @Tags Tags
// @Description delete tag from link
// @ID add-tag-to-link
// @Accept json
// @Produce json
// @Success 204 {object} response.ApiResponse
// @Failure 400 {object} response.ApiResponse
// @Failure 404 {object} response.ApiResponse
// @Failure 500 {object} response.ApiResponse
// @Router /tags/delete [delete]
func DeleteTagFromLink(repo *tags.TagsRepository, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		tagNameArg := chi.URLParam(r, "tagName")
		if tagNameArg == "" {
			response.Bad(w, "tag name parameter is required")
			return
		}

		linkIDArg := chi.URLParam(r, "linkID")
		if linkIDArg == "" {
			response.Bad(w, "url parameter is required")
			return
		}

		linkID, err := strconv.ParseInt(linkIDArg, 0, 64)
		if err != nil {
			response.Bad(w, "linkID is not a number")
			return
		}

		if _, err := repo.DeleteTagFromLink(linkID, tagNameArg); err != nil {
			logError(logger, err)
			response.InternalError(w, "internal server error")
			return
		}

		response.Deleted(w)
	})
}
