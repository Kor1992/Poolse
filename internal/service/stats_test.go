package service

import (
	"context"
	"errors"
	"testing"

	"pulse/internal/domain"
	"pulse/internal/repository"
)

type mockStatsRepository struct {
	stats     domain.WebsiteStats
	err       error
	called    bool
	websiteID int64
}

func (m *mockStatsRepository) GetWebsiteStats(
	ctx context.Context,
	websiteID int64,
) (domain.WebsiteStats, error) {
	m.called = true
	m.websiteID = websiteID
	return m.stats, m.err
}

func TestCalculateUptime(t *testing.T) {
	tests := []struct {
		name       string
		total      int64
		successful int64
		want       float64
	}{
		{name: "all successful", total: 100, successful: 100, want: 100},
		{name: "75 percent successful", total: 100, successful: 75, want: 75},
		{name: "none successful", total: 100, successful: 0, want: 0},
		{name: "no checks", total: 0, successful: 0, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateUptime(tt.total, tt.successful)
			if got != tt.want {
				t.Fatalf("CalculateUptime(%d, %d) = %v, want %v", tt.total, tt.successful, got, tt.want)
			}
		})
	}
}

func TestStatsService_GetWebsiteStats(t *testing.T) {
	expected := domain.WebsiteStats{
		WebsiteID:        1,
		TotalChecks:      100,
		SuccessfulChecks: 85,
		FailedChecks:     15,
		AvgResponseMs:    250,
	}

	service := NewStatsService(&mockStatsRepository{stats: expected})

	got, err := service.GetWebsiteStats(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetWebsiteStats() unexpected error = %v", err)
	}
	if got.WebsiteID != 1 {
		t.Fatalf("GetWebsiteStats() WebsiteID = %d, want %d", got.WebsiteID, 1)
	}
	if got.UptimePercent != 85 {
		t.Fatalf("GetWebsiteStats() UptimePercent = %v, want %v", got.UptimePercent, 85.0)
	}
	if got.TotalChecks != expected.TotalChecks || got.SuccessfulChecks != expected.SuccessfulChecks || got.FailedChecks != expected.FailedChecks {
		t.Fatalf("GetWebsiteStats() stats mismatch: got %+v, want %+v", got, expected)
	}
}

func TestStatsService_GetWebsiteStats_Errors(t *testing.T) {
	t.Run("invalid id", func(t *testing.T) {
		repo := &mockStatsRepository{}
		svc := NewStatsService(repo)

		_, err := svc.GetWebsiteStats(context.Background(), -1)
		if !errors.Is(err, ErrInvalidID) {
			t.Fatalf("GetWebsiteStats() error = %v, want %v", err, ErrInvalidID)
		}
		if repo.called {
			t.Fatalf("GetWebsiteStats() called repository for invalid ID")
		}
	})

	t.Run("repository error is propagated", func(t *testing.T) {
		repo := &mockStatsRepository{err: repository.ErrNotFound}
		svc := NewStatsService(repo)

		_, err := svc.GetWebsiteStats(context.Background(), 1)
		if !errors.Is(err, repository.ErrNotFound) {
			t.Fatalf("GetWebsiteStats() error = %v, want %v", err, repository.ErrNotFound)
		}
	})
}
