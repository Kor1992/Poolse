package service

import (
	"context"

	"pulse/internal/domain"
)

type StatsRepository interface {
	GetWebsiteStats(ctx context.Context, websiteID int64) (domain.WebsiteStats, error)
}

type StatsService struct {
	repo StatsRepository
}

func NewStatsService(repo StatsRepository) *StatsService {
	return &StatsService{repo: repo}
}

func (s *StatsService) GetWebsiteStats(
	ctx context.Context,
	websiteID int64,
) (domain.WebsiteStats, error) {
	if websiteID <= 0 {
		return domain.WebsiteStats{}, ErrInvalidID
	}

	stats, err := s.repo.GetWebsiteStats(ctx, websiteID)
	if err != nil {
		return domain.WebsiteStats{}, err
	}

	stats.UptimePercent = CalculateUptime(stats.TotalChecks, stats.SuccessfulChecks)
	return stats, nil
}

func CalculateUptime(total, successful int64) float64 {
	if total == 0 {
		return 0
	}
	res := float64(successful) / float64(total) * 100
	return res
}
