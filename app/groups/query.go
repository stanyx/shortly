package groups

import (
	"database/sql"
	"shortly/utils"
)

// GroupsRepository ...
type GroupsRepository struct {
	utils.AuditQuery
	DB *sql.DB
}

// AddGroup ...
func (repo *GroupsRepository) AddGroup(g Group) (int64, error) {
	return repo.CreateTx("groups", repo.DB, `
		insert into "groups" (name, description, account_id) 
		values ( $1, $2, $3 )
		returning id`,
		g.Name,
		g.Description,
		g.AccountID,
	)
}

// UpdateGroup ...
func (repo *GroupsRepository) UpdateGroup(id, accountID int64, g Group) (int64, error) {
	return repo.UpdateTx(id, "groups", repo.DB, `
		update "groups" set name = $1, description = $2 
		where id = $3 and account_id = $4
		returning id`,
		g.Name,
		g.Description,
		id, accountID,
	)
}

// DeleteGroup ...
func (repo *GroupsRepository) DeleteGroup(groupID, accountID int64) error {
	//TODO - cascade delete
	_, err := repo.DeleteTx("groups", repo.DB,
		`delete from groups e where id = $1 and account_id = $2
		returning id`,
		groupID, accountID)
	return err
}

// AddUserToGroup ...
func (repo *GroupsRepository) AddUserToGroup(groupID, userID int64) error {
	_, err := repo.CreateTx("users_groups", repo.DB, `
		insert into users_groups (group_id, user_id) values ( $1, $2 )
		returning id`,
		groupID, userID)
	return err
}

// DeleteUserFromGroup ...
func (repo *GroupsRepository) DeleteUserFromGroup(groupID int64, userID int64) error {
	_, err := repo.DeleteTx("users_groups", repo.DB, `
		delete from users_groups e where group_id = $1 and user_id = $2
		returning id`,
		groupID, userID)
	return err
}

// TODO - move to another application
// GetAccountGroups ...
func (repo *GroupsRepository) GetAccountGroups(accountID int64) ([]Group, error) {
	rows, err := repo.DB.Query(
		"select id, name, description from groups where account_id = $1",
		accountID,
	)

	if err != nil {
		return nil, err
	}

	var list []Group
	for rows.Next() {
		var u Group
		err := rows.Scan(&u.ID, &u.Name, &u.Description)
		if err != nil {
			return nil, err
		}
		list = append(list, u)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	defer rows.Close()

	return list, nil
}
