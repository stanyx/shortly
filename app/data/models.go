package data

import (
	"time"
)

type Click struct {
	LinkID  int64
	Time    time.Time
	Headers string
}