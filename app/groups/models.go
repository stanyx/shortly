package groups

// Group ...
type Group struct {
	ID          int64
	AccountID   int64  `binding:"required"`
	Name        string `binding:"required"`
	Description string
}
