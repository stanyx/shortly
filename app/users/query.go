package users

import (
	"database/sql"
	"golang.org/x/crypto/bcrypt"
)

type UsersRepository struct {
	DB *sql.DB
}

func (r *UsersRepository) CreateUser(u User) (int64, error) {

	password, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	var userID int64

	err = r.DB.QueryRow(`
		INSERT INTO users(username, password, phone, email, company) VALUES( $1, $2, $3, $4, $5 )
		RETURNING id`,
		u.Username,
		password,
		u.Phone,
		u.Email,
		u.Company,
	).Scan(&userID)

	if err != nil {
		return 0, err
	}

	return userID, nil
}

func (r *UsersRepository) GetUser(username string) (*User, error) {

	var user User

	err := r.DB.QueryRow(`
		SELECT id, username, password, phone, email, company FROM users WHERE username = $1`,
		username,
	).Scan(&user.ID, &user.Username, &user.Password, &user.Phone, &user.Email, &user.Company) 

	if err != nil {
		return nil, err
	}

	return &user, nil

}