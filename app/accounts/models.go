package accounts

type User struct {
	ID        int64
	Username  string `binding:"required"`
	Password  string `binding:"required"`
	Phone     string
	Email     string `binding:"required"`
	Company   string
	IsStaff   bool
	AccountID int64 `binding:"required"`
	RoleID    int64
}

type Group struct {
	ID          int64
	AccountID   int64  `binding:"required"`
	Name        string `binding:"required"`
	Description string
}
