package rbac

// Permission ...
type Permission struct {
	Url    string
	Name   string
	Method string
}

func NewPermission(url, name, method string) Permission {
	return Permission{url, name, method}
}

// Role ...
type Role struct {
	ID          int64
	AccountID   int64
	Name        string
	Description string
}
