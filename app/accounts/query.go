package users

import (
	"fmt"
	"database/sql"
	"golang.org/x/crypto/bcrypt"
)

type UsersRepository struct {
	DB *sql.DB
}

func (r *UsersRepository) CreateAccount(u User) (int64, error) {

	tx, err := r.DB.Begin()
	if err != nil {
		return 0, err
	}

	var accountID int64
	err = tx.QueryRow(`
		insert into accounts (name, created_at, confirmed, verified) 
		values ( $1 , now(), false, false )
		returning id`,
		u.Company,
	).Scan(&accountID)

	if err != nil {
		return 0, err
	}

	password, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}

	var userID int64

	err = tx.QueryRow(`
		INSERT INTO users(username, password, phone, email, company, is_staff, account_id, role_id) 
		VALUES ( $1, $2, $3, $4, $5, $6, $7, $8 )
		RETURNING id`,
		u.Username,
		password,
		u.Phone,
		u.Email,
		u.Company,
		u.IsStaff,
		accountID,
		u.RoleID,
	).Scan(&userID)

	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return userID, nil
}

func (r *UsersRepository) CreateUser(accountID int64, u User) (int64, error) {

	password, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	var userID int64

	err = r.DB.QueryRow(`
		INSERT INTO users (username, password, phone, email, company, is_staff, account_id, role_id) 
		VALUES( $1, $2, $3, $4, $5, $6, $7, $8 )
		RETURNING id`,
		u.Username,
		password,
		u.Phone,
		u.Email,
		u.Company,
		u.IsStaff,
		accountID,
		u.RoleID,
	).Scan(&userID)

	if err != nil {
		return 0, err
	}

	return userID, nil
}

func (r *UsersRepository) GetUser(username string) (*User, error) {

	var user User

	var accountID *int64
	var roleID *int64
	var isStaff *bool

	err := r.DB.QueryRow(`
		SELECT id, username, password, phone, email, company, is_staff, account_id, role_id 
		FROM users WHERE username = $1`,
		username,
	).Scan(
		&user.ID, 
		&user.Username, 
		&user.Password,
		&user.Phone, 
		&user.Email, 
		&user.Company,
		&user.IsStaff,
		&accountID,
		&roleID,
	) 

	if accountID != nil {
		user.AccountID = *accountID
	}

	if roleID != nil {
		user.RoleID = *roleID
	}

	if isStaff != nil {
		user.IsStaff = *isStaff
	}

	if err != nil {
		return nil, err
	}

	return &user, nil

}

func (r *UsersRepository) getUserFiltered(fieldName string, value interface{}) (*User, error) {
	var user User

	err := r.DB.QueryRow(fmt.Sprintf(`
		SELECT id, username, password, phone, email, company, is_staff, account_id, role_id
		FROM users WHERE %s = $1`, fieldName),
		value,
	).Scan(
		&user.ID, 
		&user.Username, 
		&user.Password, 
		&user.Phone, 
		&user.Email, 
		&user.Company,
		&user.IsStaff,
		&user.AccountID,
		&user.RoleID,
	) 

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UsersRepository) GetUserByID(userID int64) (*User, error) {
	return r.getUserFiltered("id", userID)
}

func (r *UsersRepository) GetUserByAccountID(accountID int64) (*User, error) {
	return r.getUserFiltered("account_id", accountID)
}

func (repo *UsersRepository) AddGroup(g Group) (int64, error) {
	
	var groupID int64
	err := repo.DB.QueryRow(`
		INSERT INTO groups (name, description, account_id) 
		VALUES( $1, $2, $3)
		RETURNING id`,
		g.Name,
		g.Description,
		g.AccountID,
	).Scan(&groupID)

	if err != nil {
		return 0, err
	}

	return groupID, nil
}

func (repo *UsersRepository) DeleteGroup(groupID, accountID int64) error {
	//TODO - cascade delete
	_, err := repo.DB.Exec(`delete from groups where id = $1 and account_id = $2`, groupID, accountID)
	return err
}

func (repo *UsersRepository) AddUserToGroup(groupID, userID int64) error {
	_, err := repo.DB.Exec(`
		insert into users_groups (group_id, user_id) values ( $1, $2 )`,
		groupID, userID,
	)
	return err
}

func (repo *UsersRepository) DeleteUserFromGroup(groupID int64, userID int64) error {
	_, err := repo.DB.Exec(`
		delete from users_groups where group_id = $1 and user_id = $2`,
		userID, groupID,
	)
	return err
}