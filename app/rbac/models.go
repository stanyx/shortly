package rbac

type Permission struct {
	Url    string
	Name   string
	Method string
}

func NewPermission(url, name, method string) Permission {
	return Permission{url, name, method}
}

type Role struct {
	ID          int64
	Name        string
	Description string
}