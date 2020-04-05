package accounts

import (
	"time"
)

// Account ...
type Account struct {
	ID        int64
	Name      string
	CreatedAt time.Time
	Verified  bool
}
