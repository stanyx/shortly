package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"shortly/app/accounts"
	"shortly/app/rbac"

	"github.com/go-chi/chi"
	validator "gopkg.in/go-playground/validator.v9"
)

func RbacRoutes(r chi.Router, auth func(rbac.Permission, http.Handler) http.HandlerFunc, permissions map[string]rbac.Permission, userRepo *accounts.UsersRepository, repo *rbac.RbacRepository, logger *log.Logger) {

	r.Get("/api/v1/roles", auth(
		rbac.NewPermission("/api/v1/roles", "read_roles", "GET"),
		GetUserRoles(userRepo, repo, logger),
	))

	r.Post("/api/v1/roles/create", auth(
		rbac.NewPermission("/api/v1/roles/create", "create_role", "POST"),
		CreateUserRole(repo, logger),
	))

	r.Delete("/api/v1/roles/delete", auth(
		rbac.NewPermission("/api/v1/roles/delete", "delete_role", "DELETE"),
		DeleteUserRole(userRepo, repo, logger),
	))

	r.Post("/api/v1/roles/set", auth(
		rbac.NewPermission("/api/v1/roles/set", "set_user_role", "POST"),
		SetUserRole(userRepo, repo, logger),
	))

	r.Post("/api/v1/roles/grant", auth(
		rbac.NewPermission("/api/v1/roles/grant", "grant_permission", "POST"),
		GrantAccessForRole(userRepo, repo, logger),
	))

	r.Post("/api/v1/roles/revoke", auth(
		rbac.NewPermission("/api/v1/roles/revoke", "revoke_permission", "POST"),
		RevokeAccessForRole(userRepo, repo, logger),
	))

	r.Get("/api/v1/permissions", auth(
		rbac.NewPermission("/api/v1/permissions", "read_permissions", "GET"),
		GetAllPermissions(permissions, userRepo, repo, logger),
	))

}

func GetUserRoles(userRepo *accounts.UsersRepository, repo *rbac.RbacRepository, logger *log.Logger) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		accountID := r.Context().Value("user").(*JWTClaims).AccountID

		roles, err := repo.GetUserRoles(accountID)
		if err != nil {
			logError(logger, err)
			apiError(w, "get user roles error", http.StatusInternalServerError)
			return
		}

		var list []RoleResponse

		for _, r := range roles {
			list = append(list, RoleResponse{
				ID:          r.ID,
				Name:        r.Name,
				Description: r.Description,
			})
		}

		response(w, list, http.StatusOK)
	})
}

type CreateRoleForm struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type RoleResponse struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func CreateUserRole(repo rbac.IRbacRepository, logger *log.Logger) http.HandlerFunc {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		accountID := r.Context().Value("user").(*JWTClaims).AccountID

		var form CreateRoleForm

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

		role := rbac.Role{
			Name:        form.Name,
			Description: form.Description,
		}

		roleID, err := repo.CreateRole(accountID, role)
		if err != nil {
			logError(logger, err)
			apiError(w, "create role error", http.StatusInternalServerError)
			return
		}

		response(w, RoleResponse{
			ID:          roleID,
			Name:        role.Name,
			Description: role.Description,
		}, http.StatusOK)
	})
}

type SetRoleForm struct {
	UserID int64 `json:"userId"`
	RoleID int64 `json:"roleId"`
}

func SetUserRole(userRepo *accounts.UsersRepository, repo *rbac.RbacRepository, logger *log.Logger) http.HandlerFunc {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var form SetRoleForm

		if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
			logError(logger, err)
			apiError(w, "decode form error", http.StatusBadRequest)
			return
		}

		if form.RoleID == 0 {
			apiError(w, "roleId is required", http.StatusBadRequest)
			return
		}

		if form.UserID == 0 {
			apiError(w, "userId is required", http.StatusBadRequest)
			return
		}

		role, err := repo.GetRole(form.RoleID)
		if err != nil {
			logError(logger, err)
			apiError(w, "get role error", http.StatusInternalServerError)
			return
		}

		user, err := userRepo.GetUserByID(form.UserID)
		if err != nil {
			logError(logger, err)
			apiError(w, "get user error", http.StatusInternalServerError)
			return
		}

		if err := repo.AddRoleForUser(user.ID, role.ID); err != nil {
			logError(logger, err)
			apiError(w, "add role for user error", http.StatusInternalServerError)
			return
		}

		ok(w)
	})
}

type DeleteRoleForm struct {
	UserID int64 `json:"userId"`
	RoleID int64 `json:"roleId"`
}

func DeleteUserRole(userRepo *accounts.UsersRepository, repo *rbac.RbacRepository, logger *log.Logger) http.HandlerFunc {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var form DeleteRoleForm

		if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
			logError(logger, err)
			apiError(w, "decode form error", http.StatusBadRequest)
			return
		}

		if form.RoleID == 0 {
			apiError(w, "roleId is required", http.StatusBadRequest)
			return
		}

		if form.UserID == 0 {
			apiError(w, "userId is required", http.StatusBadRequest)
			return
		}

		role, err := repo.GetRole(form.RoleID)
		if err != nil {
			logError(logger, err)
			apiError(w, "get role error", http.StatusInternalServerError)
			return
		}

		user, err := userRepo.GetUserByID(form.UserID)
		if err != nil {
			logError(logger, err)
			apiError(w, "get user error", http.StatusInternalServerError)
			return
		}

		if err := repo.DeleteRoleForUser(user.ID, role.ID); err != nil {
			logError(logger, err)
			apiError(w, "delete role for user error", http.StatusInternalServerError)
			return
		}

		ok(w)
	})
}

type GrantRoleForm struct {
	RoleID   int64  `json:"roleId"`
	Resource string `json:"resource"`
	Method   string `json:"method"`
}

func GrantAccessForRole(userRepo *accounts.UsersRepository, repo *rbac.RbacRepository, logger *log.Logger) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var form GrantRoleForm

		if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
			logError(logger, err)
			apiError(w, "decode form error", http.StatusBadRequest)
			return
		}

		role, err := repo.GetRole(form.RoleID)
		if err != nil {
			logError(logger, err)
			apiError(w, "get role error", http.StatusInternalServerError)
			return
		}

		if err := repo.GrantAccessForRole(role.ID, rbac.Permission{Url: form.Resource, Method: form.Method}); err != nil {
			logError(logger, err)
			apiError(w, "grant access for role error", http.StatusInternalServerError)
			return
		}

		ok(w)
	})
}

type RevokeRoleForm struct {
	RoleID   int64  `json:"roleId"`
	Resource string `json:"resource"`
	Method   string `json:"method"`
}

func RevokeAccessForRole(userRepo *accounts.UsersRepository, repo *rbac.RbacRepository, logger *log.Logger) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var form RevokeRoleForm

		if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
			logError(logger, err)
			apiError(w, "decode form error", http.StatusBadRequest)
			return
		}

		role, err := repo.GetRole(form.RoleID)
		if err != nil {
			logError(logger, err)
			apiError(w, "get role error", http.StatusInternalServerError)
			return
		}

		if err := repo.RevokeAccessForRole(role.ID, rbac.Permission{Url: form.Resource, Method: form.Method}); err != nil {
			logError(logger, err)
			apiError(w, "revoke access for role error", http.StatusInternalServerError)
			return
		}

		ok(w)
	})
}

type PermissionResponse struct {
	Url           string `json:"url"`
	Name          string `json:"name"`
	Method        string `json:"method"`
	AccessGranted bool   `json:"accessGranted"`
}

func GetAllPermissions(permissions map[string]rbac.Permission, userRepo *accounts.UsersRepository, repo *rbac.RbacRepository, logger *log.Logger) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		roleIds := r.URL.Query()["roleId"]
		if len(roleIds) != 1 {
			apiError(w, "invalid number of query values for parameter <roleId>, must be 1", http.StatusBadRequest)
			return
		}

		if roleIds[0] == "" {
			apiError(w, "roleId query parameter is required", http.StatusBadRequest)
			return
		}

		roleID, _ := strconv.ParseInt(roleIds[0], 0, 64)

		casbinPerms, err := repo.GetPermissionsForRole(roleID)
		if err != nil {
			apiError(w, "get permission error", http.StatusInternalServerError)
			return
		}

		casbinMap := make(map[[2]string]rbac.Permission)
		for _, cp := range casbinPerms {
			casbinMap[[2]string{cp.Url, cp.Method}] = cp
		}

		var list []PermissionResponse

		for k, v := range permissions {

			_, accessGranted := casbinMap[[2]string{k, v.Method}]

			list = append(list, PermissionResponse{
				Url:           k,
				Name:          v.Name,
				Method:        v.Method,
				AccessGranted: accessGranted,
			})
		}

		response(w, list, http.StatusOK)
	})
}
