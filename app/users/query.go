package users

import (
	"database/sql"
)

type UsersRepository struct {
	db *sql.DB
}