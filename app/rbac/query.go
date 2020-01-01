package rbac

import (
	"fmt"
	"database/sql"
	"log"

	"github.com/casbin/casbin/v2"
)

type RbacRepository struct {
	DB       *sql.DB
	Logger   *log.Logger
	Enforcer *casbin.Enforcer
}

func (repo *RbacRepository) GetRole(roleID int64) (Role, error) {
	var row Role
	err := repo.DB.QueryRow(`
		select id, name, description from roles where id = $1
	`, roleID).Scan(&row.ID, &row.Name, &row.Description)
	return row, err
}

func (repo *RbacRepository) GetUserRoles(ownerID int64) ([]Role, error) {

	rows, err := repo.DB.Query("select id, name, description from roles where owner_id = $1", ownerID)
	if err != nil {
		return nil, err
	}

	var roles []Role

	for rows.Next() {
		var r Role
		err := rows.Scan(&r.ID, &r.Name, &r.Description)
		if err != nil {
			return nil, err
		}
		roles = append(roles, r)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	defer rows.Close()

	return roles, nil

}

func (repo *RbacRepository) CreateRole(ownerID int64, r Role) (int64, error) {
	var roleID int64
	err := repo.DB.QueryRow(`
		insert into roles(name, description, owner_id) values ($1, $2, $3) returning id`, 
		r.Name, r.Description, ownerID,
	).Scan(&roleID)

	return roleID, err
}

func (repo *RbacRepository) DeleteRole(r Role) error {
	_, err := repo.DB.Exec("delete from roles where id = $1", r.ID)
	if err == nil {
		repo.Enforcer.DeleteRole(fmt.Sprintf("role:%v", r.ID))
	}
	return err
}

func (repo *RbacRepository) AddRoleForUser(userID int64, roleID int64) error {
	repo.Enforcer.AddRoleForUser(fmt.Sprintf("user:%v", userID), fmt.Sprintf("role:%v", roleID))
	return repo.Enforcer.SavePolicy()
}

func (repo *RbacRepository) DeleteRoleForUser(userID int64, roleID int64) error {
	repo.Enforcer.DeleteRoleForUser(fmt.Sprintf("user:%v", userID), fmt.Sprintf("role:%v", roleID))
	return repo.Enforcer.SavePolicy()
}

func (repo *RbacRepository) GrantAccessForUser(userID int64, method string) error {
	repo.Enforcer.AddPermissionForUser(fmt.Sprintf("user:%v", userID), method)
	return repo.Enforcer.SavePolicy()
}

func (repo *RbacRepository) RevokeAccessForUser(userID int64, method string) error {
	repo.Enforcer.DeletePermissionForUser(fmt.Sprintf("user:%v", userID), method)
	return repo.Enforcer.SavePolicy()
}

func (repo *RbacRepository) GrantAccessForRole(roleID int64, permission Permission) error {
	repo.Enforcer.AddNamedPolicy("p", fmt.Sprintf("role:%v", roleID), permission.Url, permission.Method)
	return repo.Enforcer.SavePolicy()
}

func (repo *RbacRepository) RevokeAccessForRole(roleID int64, permission Permission) error {
	repo.Enforcer.RemovePolicy(fmt.Sprintf("role:%v", roleID), permission.Url, permission.Method)
	return repo.Enforcer.SavePolicy()
}

func (repo *RbacRepository) GetPermissionsForRole(roleID int64) ([]Permission, error) {
	ps, _ := repo.Enforcer.GetImplicitPermissionsForUser(fmt.Sprintf("role:%v", roleID))

	var perms []Permission
	for _, p := range ps {
		_, resource, method := p[0], p[1], p[2]
		perms = append(perms, Permission{
			Url:    resource,
			Method: method,
		})
	}

	return perms, nil
}

func (repo *RbacRepository) GetPermissionsForUser(userID int64) ([]Permission, error) {
	ps, _ := repo.Enforcer.GetImplicitPermissionsForUser(fmt.Sprintf("user:%v", userID))

	var perms []Permission
	for _, p := range ps {
		_, resource, method := p[0], p[1], p[2]
		perms = append(perms, Permission{
			Url:    resource,
			Method: method,
		})
	}

	return perms, nil
}