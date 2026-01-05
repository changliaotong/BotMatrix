package bot

import (
	"time"
)

// NewTicker returns a channel that ticks at the specified interval in seconds
func NewTicker(seconds int) <-chan time.Time {
	return time.NewTicker(time.Duration(seconds) * time.Second).C
}
