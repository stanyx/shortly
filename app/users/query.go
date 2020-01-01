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
		INSERT INTO users(username, password, phone, email, company, is_staff, admin_id, role_id) 
		VALUES( $1, $2, $3, $4, $5, $6, $7, $8 )
		RETURNING id`,
		u.Username,
		password,
		u.Phone,
		u.Email,
		u.Company,
		u.IsStaff,
		u.AdminID,
		u.RoleID,
	).Scan(&userID)

	if err != nil {
		return 0, err
	}

	return userID, nil
}

func (r *UsersRepository) GetUser(username string) (*User, error) {

	var user User

	var adminID *int64
	var roleID *int64
	var isStaff *bool

	err := r.DB.QueryRow(`
		SELECT id, username, password, phone, email, company, is_staff, admin_id, role_id 
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
		&adminID,
		&roleID,
	) 

	if adminID != nil {
		user.AdminID = *adminID
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

func (r *UsersRepository) GetUserByID(userID int64) (*User, error) {

	var user User

	err := r.DB.QueryRow(`
		SELECT id, username, password, phone, email, company, is_staff, admin_id, role_id
		FROM users WHERE id = $1`,
		userID,
	).Scan(
		&user.ID, 
		&user.Username, 
		&user.Password, 
		&user.Phone, 
		&user.Email, 
		&user.Company,
		&user.IsStaff,
		&user.AdminID,
		&user.RoleID,
	) 

	if err != nil {
		return nil, err
	}

	return &user, nil

}

func (repo *UsersRepository) AddGroup(g Group) (int64, error) {
	
	var groupID int64
	err := repo.DB.QueryRow(`
		INSERT INTO groups (name, description, user_id) 
		VALUES( $1, $2, $3)
		RETURNING id`,
		g.Name,
		g.Description,
		g.UserID,
	).Scan(&groupID)

	if err != nil {
		return 0, err
	}

	return groupID, nil
}

func (repo *UsersRepository) DeleteGroup(groupID, userID int64) error {
	_, err := repo.DB.Exec(`delete from groups where id = $1 and user_id = $2`, groupID, userID)
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