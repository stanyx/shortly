package campaigns

import (
	"shortly/app/links"
)

type Campaign struct {
	ID          int64
	Name        string
	Description string
	AccountID   int64
	Links       []links.Link
}

type UTMSetting struct {
	Source  string
	Medium  string
	Term    string
	Content string
}
