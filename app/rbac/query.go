package rbac

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/casbin/casbin/v2"
)

type IRbacRepository interface {
	CreateRole(accountID int64, role Role) (int64, error)
	GetRole(roleID int64) (Role, error)
}

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

func (repo *RbacRepository) GetUserRoles(accountID int64) ([]Role, error) {

	rows, err := repo.DB.Query("select id, name, description from roles where account_id = $1", accountID)
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

func (repo *RbacRepository) CreateRole(accountID int64, r Role) (int64, error) {
	var roleID int64
	err := repo.DB.QueryRow(`
		insert into roles(name, description, account_id) values ($1, $2, $3) returning id`,
		r.Name, r.Description, accountID,
	).Scan(&roleID)

	return roleID, err
}

func (repo *RbacRepository) DeleteRole(r Role) error {
	_, err := repo.DB.Exec("delete from roles where id = $1", r.ID)
	if err == nil {
		_, _ = repo.Enforcer.DeleteRole(fmt.Sprintf("role:%v", r.ID))
	}
	return err
}

func (repo *RbacRepository) ChangeRoleForUser(userID int64, roleID int64) error {
	tx, err := repo.DB.Begin()
	if err != nil {
		return err
	}

	if _, err := tx.Exec("update users set role_id = $1 where id = $2", roleID, userID); err != nil {
		_ = tx.Rollback()
		return err
	}

	_, err = repo.Enforcer.AddRoleForUser(fmt.Sprintf("user:%v", userID), fmt.Sprintf("role:%v", roleID))
	if err != nil {
		return err
	}

	if err := repo.Enforcer.SavePolicy(); err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (repo *RbacRepository) DeleteRoleForUser(userID int64, roleID int64) error {

	tx, err := repo.DB.Begin()
	if err != nil {
		return err
	}

	if _, err := tx.Exec("update users set role_id = $1 where id = $2", 0, userID); err != nil {
		_ = tx.Rollback()
		return err
	}

	_, err = repo.Enforcer.DeleteRoleForUser(fmt.Sprintf("user:%v", userID), fmt.Sprintf("role:%v", roleID))
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := repo.Enforcer.SavePolicy(); err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (repo *RbacRepository) GrantAccessForRole(roleID int64, permission Permission) error {
	_, err := repo.Enforcer.AddNamedPolicy("p", fmt.Sprintf("role:%v", roleID), permission.Url, permission.Method)
	if err != nil {
		return err
	}
	return repo.Enforcer.SavePolicy()
}

func (repo *RbacRepository) RevokeAccessForRole(roleID int64, permission Permission) error {
	_, err := repo.Enforcer.RemovePolicy(fmt.Sprintf("role:%v", roleID), permission.Url, permission.Method)
	if err != nil {
		return err
	}
	return repo.Enforcer.SavePolicy()
}

func (repo *RbacRepository) GetPermissionsForRole(roleID int64) ([]Permission, error) {
	ps, err := repo.Enforcer.GetImplicitPermissionsForUser(fmt.Sprintf("role:%v", roleID))
	if err != nil {
		return nil, err
	}

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
