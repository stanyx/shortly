package webhooks

type Webhook struct {
	ID          int64
	AccountID   int64
	Name        string
	Description string
	Events      []string
	URL         string
	Active      bool
}
