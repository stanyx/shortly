package api

import (
	"log"
	"net/http"
	"encoding/json"
	"strconv"

	"shortly/app/users"
	"shortly/app/rbac"
)

func RbacRoutes(auth func(rbac.Permission, http.Handler) http.Handler, permissions map[string]rbac.Permission, userRepo *users.UsersRepository, repo *rbac.RbacRepository, logger *log.Logger) {

	http.Handle("/api/v1/users/roles", auth(
		rbac.NewPermission("/api/v1/users/roles", "read_roles", "GET"), 
		GetUserRoles(userRepo, repo, logger),
	))

	http.Handle("/api/v1/users/roles/create", auth(
		rbac.NewPermission("/api/v1/users/roles/create", "create_role", "POST"), 
		CreateUserRole(userRepo, repo, logger),
	))

	http.Handle("/api/v1/users/roles/delete", auth(
		rbac.NewPermission("/api/v1/users/roles/delete", "delete_role", "DELETE"), 
		DeleteUserRole(userRepo, repo, logger),
	))

	http.Handle("/api/v1/users/roles/set", auth(
		rbac.NewPermission("/api/v1/users/roles/set", "set_user_role", "POST"), 
		SetUserRole(userRepo, repo, logger),
	))

	http.Handle("/api/v1/users/roles/grant", auth(
		rbac.NewPermission("/api/v1/users/roles/grant", "grant_permission", "POST"), 
		GrantAccessForRole(userRepo, repo, logger),
	))

	http.Handle("/api/v1/users/roles/revoke", auth(
		rbac.NewPermission("/api/v1/users/roles/revoke", "revoke_permission", "POST"), 
		RevokeAccessForRole(userRepo, repo, logger),
	))

	http.Handle("/api/v1/users/permissions", auth(
		rbac.NewPermission("/api/v1/users/permissions", "read_permissions", "GET"),
		GetAllPermissions(permissions, userRepo, repo, logger),
	))

}

func GetUserRoles(userRepo *users.UsersRepository, repo *rbac.RbacRepository, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		ownerID := r.Context().Value("user").(*JWTClaims).UserID

		roles, err := repo.GetUserRoles(ownerID)
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
	Name        string `json:"name"`
	Description string `json:"description"`
}

type RoleResponse struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func CreateUserRole(userRepo *users.UsersRepository, repo *rbac.RbacRepository, logger *log.Logger) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		ownerID := r.Context().Value("user").(*JWTClaims).UserID

		var form CreateRoleForm

		if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
			logError(logger, err)
			apiError(w, "decode form error", http.StatusBadRequest)
			return
		}

		role := rbac.Role{
			Name:        form.Name,
			Description: form.Description,
		}

		roleID, err := repo.CreateRole(ownerID, role)
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

func SetUserRole(userRepo *users.UsersRepository, repo *rbac.RbacRepository, logger *log.Logger) http.Handler {

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

func DeleteUserRole(userRepo *users.UsersRepository, repo *rbac.RbacRepository, logger *log.Logger) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var form DeleteRoleForm

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

func GrantAccessForRole(userRepo *users.UsersRepository, repo *rbac.RbacRepository, logger *log.Logger) http.Handler {
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

func RevokeAccessForRole(userRepo *users.UsersRepository, repo *rbac.RbacRepository, logger *log.Logger) http.Handler {
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

func GetAllPermissions(permissions map[string]rbac.Permission, userRepo *users.UsersRepository, repo *rbac.RbacRepository, logger *log.Logger) http.Handler {
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