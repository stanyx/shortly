package data

import (
	"time"
)

// Click ...
type Click struct {
	LinkID  int64
	Time    time.Time
	Headers string
}
