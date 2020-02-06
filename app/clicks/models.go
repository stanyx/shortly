package clicks

import "time"

// ClickData ...
type ClickData struct {
	Time  time.Time
	Count int64
}

type LinkData struct {
	Time     time.Time
	Referer  string
	Location string
	Count    int64
}
