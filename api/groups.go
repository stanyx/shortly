package api

import (
	"encoding/json"
	"log"
	"net/http"

	"shortly/api/response"
	"shortly/app/groups"
	"shortly/app/users"
)

// GroupResponse ...
type GroupResponse struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// GetGroups returns http handler that returns all groups in the current account
func GetGroups(repo *groups.GroupsRepository, logger *log.Logger) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		rows, err := repo.GetAccountGroups(getAccountID(r))
		if err != nil {
			logError(logger, err)
			response.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		var list []GroupResponse
		for _, r := range rows {
			list = append(list, GroupResponse{
				ID:          r.ID,
				Name:        r.Name,
				Description: r.Description,
			})
		}

		response.Object(w, list, http.StatusOK)
	})
}

// CreateGroupForm ...
type CreateGroupForm struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// AddGroup returns http handler that adds new group to the current user account
func AddGroup(repo *groups.GroupsRepository, logger *log.Logger) http.HandlerFunc {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var form CreateGroupForm

		if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
			logError(logger, err)
			response.Error(w, "decode form error", http.StatusBadRequest)
			return
		}

		groupID, err := repo.AddGroup(groups.Group{
			AccountID:   getAccountID(r),
			Name:        form.Name,
			Description: form.Description,
		})

		if err != nil {
			logError(logger, err)
			response.Error(w, "add group error", http.StatusInternalServerError)
			return
		}

		response.Object(w, &GroupResponse{
			ID:          groupID,
			Name:        form.Name,
			Description: form.Description,
		}, http.StatusOK)
	})

}

// UpdateGroupForm ...
type UpdateGroupForm struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// UpdateGroup ...
func UpdateGroup(repo *groups.GroupsRepository, logger *log.Logger) http.HandlerFunc {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var form UpdateGroupForm

		if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
			logError(logger, err)
			response.Error(w, "decode form error", http.StatusBadRequest)
			return
		}

		groupID, err := getIntParam(r, "id")
		if err != nil {
			response.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		_, err = repo.UpdateGroup(groupID, getAccountID(r), groups.Group{
			Name:        form.Name,
			Description: form.Description,
		})

		if err != nil {
			logError(logger, err)
			response.Error(w, "update group error", http.StatusInternalServerError)
			return
		}

		response.Object(w, &GroupResponse{
			ID:          groupID,
			Name:        form.Name,
			Description: form.Description,
		}, http.StatusOK)
	})

}

// DeleteGroup ...
func DeleteGroup(repo *groups.GroupsRepository, logger *log.Logger) http.HandlerFunc {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		groupID, err := getIntParam(r, "id")
		if err != nil {
			response.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = repo.DeleteGroup(groupID, getAccountID(r))
		if err != nil {
			logError(logger, err)
			response.Error(w, "delete group error", http.StatusInternalServerError)
			return
		}

		response.Ok(w)
	})

}

// AddUserToGroupForm ...
type AddUserToGroupForm struct {
	GroupID int64 `json:"groupId"`
	UserID  int64 `json:"userId"`
}

// AddUserToGroup ...
func AddUserToGroup(userRepo *users.UsersRepository, groupRepo *groups.GroupsRepository, logger *log.Logger) http.HandlerFunc {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var form AddUserToGroupForm

		if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
			logError(logger, err)
			response.Error(w, "decode form error", http.StatusBadRequest)
			return
		}

		if _, err := userRepo.GetUserByAccountID(getAccountID(r)); err != nil {
			logError(logger, err)
			response.Error(w, "internal server error", http.StatusBadRequest)
			return
		}

		if err := groupRepo.AddUserToGroup(form.GroupID, form.UserID); err != nil {
			logError(logger, err)
			response.Error(w, "add user to group error", http.StatusInternalServerError)
			return
		}

		response.Ok(w)
	})

}

// DeleteUserFromGroupForm ...
type DeleteUserFromGroupForm struct {
	GroupID int64 `json:"groupId"`
	UserID  int64 `json:"userId"`
}

// DeleteUserFromGroup ...
func DeleteUserFromGroup(userRepo *users.UsersRepository, groupRepo *groups.GroupsRepository, logger *log.Logger) http.HandlerFunc {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		accountID := r.Context().Value("user").(*JWTClaims).AccountID

		var form DeleteUserFromGroupForm

		if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
			logError(logger, err)
			response.Error(w, "decode form error", http.StatusBadRequest)
			return
		}

		if _, err := userRepo.GetUserByAccountID(accountID); err != nil {
			logError(logger, err)
			response.Error(w, "internal server error", http.StatusBadRequest)
			return
		}

		if err := groupRepo.DeleteUserFromGroup(form.GroupID, form.UserID); err != nil {
			logError(logger, err)
			response.Error(w, "delete user from group error", http.StatusInternalServerError)
			return
		}

		response.Ok(w)
	})

}
