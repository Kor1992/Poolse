package domain

type WebsiteStats struct {
	WebsiteID        int64
	TotalChecks      int64
	SuccessfulChecks int64
	FailedChecks     int64
	AvgResponseMs    float64
	UptimePercent    float64
}
