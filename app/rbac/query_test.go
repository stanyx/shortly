package rbac

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestCreateAccount(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	accountID := int64(10)
	roleID := int64(1)

	r := Role{
		Name:        "role",
		Description: "new role for account manager",
	}

	mock.ExpectQuery("insert into roles(.+) returning id").WithArgs(r.Name, r.Description, accountID).WillReturnRows(
		sqlmock.NewRows([]string{"id"}).AddRow(roleID))

	repo := &RbacRepository{DB: db}

	if _, err := repo.CreateRole(accountID, r); err != nil {
		t.Errorf("error create account: %s", err)
	}

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
