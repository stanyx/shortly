package users

import (
	"database/sql"
	"fmt"

	"shortly/utils"

	"golang.org/x/crypto/bcrypt"
)

// UsersRepository ...
type UsersRepository struct {
	utils.AuditQuery
	DB *sql.DB
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

// UpdateUser ...
func (repo *UsersRepository) UpdateUser(id int64, user *User) error {
	_, err := repo.DB.Exec(
		"update users set email=$1, username=$2, phone=$3, company=$4 where id = $5",
		user.Email, user.Username, user.Phone, user.Company, id,
	)
	return err
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

	query := fmt.Sprintf(`
	select id, username, password, phone, email, company, is_staff, account_id, role_id
	from "users" WHERE %s = $1`, fieldName)

	err := r.DB.QueryRow(query, value).Scan(
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

const (
	UserIDField        = "id"
	UserAccountIDField = "account_id"
)

// GetUserByID ...
func (r *UsersRepository) GetUserByID(userID int64) (*User, error) {
	return r.getUserFiltered(UserIDField, userID)
}

// GetUserByAccountID ...
func (r *UsersRepository) GetUserByAccountID(accountID int64) (*User, error) {
	return r.getUserFiltered(UserAccountIDField, accountID)
}

// GetOne gets a single user from database by its id
func (repo *UsersRepository) GetOne(accountID, id int64) (*User, error) {

	var u User

	err := repo.DB.QueryRow(
		"select id, email, password, username, phone, company from users where account_id = $1 and id = $2",
		accountID, id,
	).Scan(
		&u.ID,
		&u.Email,
		&u.Password,
		&u.Username,
		&u.Phone,
		&u.Company,
	)

	if err != nil {
		return nil, err
	}

	return &u, nil
}

// UpdatePassword ...
func (repo *UsersRepository) UpdatePassword(id int64, password string) error {

	_, err := repo.DB.Exec(
		"update users set password=$1 where id = $2", password, id,
	)

	return err
}
