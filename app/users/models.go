package users

type User struct {
	ID       int64
	Username string
	Password string
	Phone    string
	Email    string
	Company  string
	IsStaff  bool
	AdminID  int64
	RoleID   int64
}

type Group struct {
	ID          int64
	UserID      int64
	Name        string
	Description string
}