package accounts

import (
	"database/sql"
	"fmt"

	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	validator "gopkg.in/go-playground/validator.v9"

	"shortly/utils"
)

// UsersRepository ...
type UsersRepository struct {
	utils.AuditQuery
	DB *sql.DB
}

// CreateAccount ...
func (r *UsersRepository) CreateAccount(tx *sql.Tx, u User) (int64, int64, error) {

	accountErrPrefix := "create account error: "

	v := validator.New()
	if err := v.Struct(u); err != nil {
		return 0, 0, errors.Wrap(err, accountErrPrefix)
	}

	accountID, err := r.Create("accounts", tx, `
		insert into accounts (name, created_at, confirmed, verified) 
		values ( $1 , now(), false, false )
		returning id`, u.Company)

	if err != nil {
		return 0, 0, errors.Wrap(err, accountErrPrefix+"(account)")
	}

	var password string

	if u.Password != "" {
		genPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		if err != nil {
			return 0, 0, errors.Wrap(err, accountErrPrefix+"(password)")
		}
		password = string(genPassword)
	}

	userID, err := r.Create("users", tx, `
		insert into "users" (username, password, phone, email, company, is_staff, account_id, role_id) 
		values ( $1, $2, $3, $4, $5, $6, $7, $8 )
		returning id`,
		u.Username,
		password,
		u.Phone,
		u.Email,
		u.Company,
		u.IsStaff,
		accountID,
		u.RoleID,
	)

	if err != nil {
		return 0, 0, errors.Wrap(err, accountErrPrefix+"(user)")
	}

	return userID, accountID, nil
}

// CreateUser ...
func (r *UsersRepository) CreateUser(accountID int64, u User) (int64, error) {

	password, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	return r.CreateTx("users", r.DB, `
		insert into "users" (username, password, phone, email, company, is_staff, account_id, role_id) 
		values ( $1, $2, $3, $4, $5, $6, $7, $8 )
		returning id`,
		u.Username,
		password,
		u.Phone,
		u.Email,
		u.Company,
		u.IsStaff,
		accountID,
		u.RoleID,
	)
}

// GetAccount ...
func (r *UsersRepository) GetAccount(accountID int64) (*Account, error) {

	var account Account
	err := r.DB.QueryRow(
		"select name, created_at, verified from accounts where id = $1",
		accountID,
	).Scan(
		&account.Name,
		&account.CreatedAt,
		&account.Verified,
	)

	if err != nil {
		return nil, err
	}

	return &account, nil
}

// GetAccountUsers ...
func (r *UsersRepository) GetAccountUsers(accountID int64) ([]User, error) {

	rows, err := r.DB.Query(
		"select id, username, email, phone from users where account_id = $1",
		accountID,
	)

	if err != nil {
		return nil, err
	}

	var list []User
	for rows.Next() {
		var u User
		err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.Phone)
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

// GetUserByEmail ...
func (r *UsersRepository) GetUserByEmail(email string) (*User, error) {

	var user User
	var accountID *int64
	var roleID *int64
	var isStaff *bool

	err := r.DB.QueryRow(`
		select id, username, password, phone, email, company, is_staff, account_id, role_id 
		from "users" where email = $1`,
		email,
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

// getUserFiltered ...
func (r *UsersRepository) getUserFiltered(fieldName string, value interface{}) (*User, error) {
	var user User

	err := r.DB.QueryRow(fmt.Sprintf(`
		select id, username, password, phone, email, company, is_staff, account_id, role_id
		from "users" WHERE %s = $1`, fieldName),
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

// GetUserByID ...
func (r *UsersRepository) GetUserByID(userID int64) (*User, error) {
	return r.getUserFiltered("id", userID)
}

// GetUserByAccountID ...
func (r *UsersRepository) GetUserByAccountID(accountID int64) (*User, error) {
	return r.getUserFiltered("account_id", accountID)
}

// TODO - move to another application
// GetAccountGroups ...
func (repo *UsersRepository) GetAccountGroups(accountID int64) ([]Group, error) {
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

// AddGroup ...
func (repo *UsersRepository) AddGroup(g Group) (int64, error) {
	return repo.CreateTx("groups", repo.DB, `
		insert into "groups" (name, description, account_id) 
		values ( $1, $2, $3 )
		returning id`,
		g.Name,
		g.Description,
		g.AccountID,
	)
}

// DeleteGroup ...
func (repo *UsersRepository) DeleteGroup(groupID, accountID int64) error {
	//TODO - cascade delete
	_, err := repo.DeleteTx("groups", repo.DB,
		`delete from groups e where id = $1 and account_id = $2
		returning id`,
		groupID, accountID)
	return err
}

// AddUserToGroup ...
func (repo *UsersRepository) AddUserToGroup(groupID, userID int64) error {
	_, err := repo.CreateTx("users_groups", repo.DB, `
		insert into users_groups (group_id, user_id) values ( $1, $2 )
		returning id`,
		groupID, userID)
	return err
}

// DeleteUserFromGroup ...
func (repo *UsersRepository) DeleteUserFromGroup(groupID int64, userID int64) error {
	_, err := repo.DeleteTx("users_groups", repo.DB, `
		delete from users_groups e where group_id = $1 and user_id = $2
		returning id`,
		groupID, userID)
	return err
}
