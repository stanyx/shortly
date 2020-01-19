package accounts

import (
	"database/sql"
	"fmt"

	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	validator "gopkg.in/go-playground/validator.v9"

	"shortly/utils"
)

type UsersRepository struct {
	utils.AuditQuery
	DB *sql.DB
}

func (r *UsersRepository) CreateAccount(u User) (int64, int64, error) {

	accountErrPrefix := "create account error: "

	v := validator.New()
	if err := v.Struct(u); err != nil {
		return 0, 0, errors.Wrap(err, accountErrPrefix)
	}

	tx, err := r.DB.Begin()
	if err != nil {
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
			_ = tx.Rollback()
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
		_ = tx.Rollback()
		return 0, 0, errors.Wrap(err, accountErrPrefix+"(user)")
	}

	if err := tx.Commit(); err != nil {
		return 0, 0, errors.Wrap(err, accountErrPrefix)
	}

	return userID, accountID, nil
}

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

func (r *UsersRepository) GetUserByID(userID int64) (*User, error) {
	return r.getUserFiltered("id", userID)
}

func (r *UsersRepository) GetUserByAccountID(accountID int64) (*User, error) {
	return r.getUserFiltered("account_id", accountID)
}

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

func (repo *UsersRepository) DeleteGroup(groupID, accountID int64) error {
	//TODO - cascade delete
	_, err := repo.DeleteTx("groups", repo.DB,
		`delete from groups e where id = $1 and account_id = $2
		returning id`,
		groupID, accountID)
	return err
}

func (repo *UsersRepository) AddUserToGroup(groupID, userID int64) error {
	_, err := repo.CreateTx("users_groups", repo.DB, `
		insert into users_groups (group_id, user_id) values ( $1, $2 )
		returning id`,
		groupID, userID)
	return err
}

func (repo *UsersRepository) DeleteUserFromGroup(groupID int64, userID int64) error {
	_, err := repo.DeleteTx("users_groups", repo.DB, `
		delete from users_groups e where group_id = $1 and user_id = $2
		returning id`,
		groupID, userID)
	return err
}
