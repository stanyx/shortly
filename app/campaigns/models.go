package campaigns

// Campaign ...
type Campaign struct {
	ID          int64
	Name        string
	Description string
	AccountID   int64
	Links       []CampaignLink
}

// UTMSetting ...
type UTMSetting struct {
	Source  string
	Medium  string
	Term    string
	Content string
}

// Channel ...
type Channel struct {
	ID   int64
	Name string
}
