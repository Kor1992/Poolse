package domain

import (
	"fmt"
	"time"
)

type CheckResult struct {
	WebsiteID  int64
	URL        string
	StatusCode int
	Duration   time.Duration
	IsUp       bool
	Error      string
	CheckedAt  time.Time
}

func (r *CheckResult) String() string {
	return fmt.Sprintf(
		"WebsiteID: %d | URL: %s | UP: %t | Status: %d | Duration: %s | Error: %s",
		r.WebsiteID, r.URL, r.IsUp, r.StatusCode, r.Duration, r.Error,
	)
}
