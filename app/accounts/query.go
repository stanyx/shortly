package accounts

import (
	"database/sql"

	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	validator "gopkg.in/go-playground/validator.v9"

	"shortly/app/users"
	"shortly/utils"
)

// AccountsRepository ...
type AccountsRepository struct {
	utils.AuditQuery
	DB *sql.DB
}

// CreateAccount ...
func (r *AccountsRepository) CreateAccount(tx *sql.Tx, u users.User) (int64, int64, error) {

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

// GetAccount ...
func (r *AccountsRepository) GetAccount(accountID int64) (*Account, error) {

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