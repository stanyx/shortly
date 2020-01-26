package links

type Link struct {
	ID          int64
	AccountID   int64
	Short       string
	Long        string
	Description string
	Tags        []string
	Hidden      bool
}
