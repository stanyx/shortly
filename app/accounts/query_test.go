package accounts

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"shortly/app/users"
)

func TestCreateAccount(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	accountID := int64(10)

	u := users.User{
		AccountID: accountID,
		Username:  "test",
		Password:  "", // no password to testing purposes (no encryption)
		Phone:     "555-555-555",
		Email:     "test@example.com",
		Company:   "test_company",
	}

	testUserID := 100

	mock.ExpectBegin()
	mock.ExpectQuery("insert into accounts (.+) returning id").WithArgs("test_company").WillReturnRows(
		sqlmock.NewRows([]string{"id"}).AddRow(accountID))
	mock.ExpectExec("insert into audit").WithArgs("accounts", accountID).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(`insert into "users" (.+) returning id`).WithArgs(
		u.Username,
		"",
		u.Phone,
		u.Email,
		u.Company,
		u.IsStaff,
		accountID,
		u.RoleID,
	).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(testUserID))
	mock.ExpectExec("insert into audit").WithArgs("users", testUserID).WillReturnResult(sqlmock.NewResult(1, 1))

	repo := &AccountsRepository{DB: db}

	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		t.Errorf("tx error: %s", err)
	}

	if _, _, err := repo.CreateAccount(tx, u); err != nil {
		t.Errorf("error create account: %s", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
