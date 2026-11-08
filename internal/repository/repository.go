package repository

import (
	"context"
	"errors"

	"pulse/internal/domain"
)

var ErrNotFound = errors.New("resource not found")

type WebsiteRepository interface {
	AddWebsite(ctx context.Context, url string) (int64, error)
	GetActiveWebsites(ctx context.Context) ([]domain.Website, error)
	SaveCheck(ctx context.Context, websiteID int64, result domain.CheckResult) error
	GetWebsiteChecks(ctx context.Context, websiteID int64, limit int) ([]domain.CheckResult, error)
	GetWebsiteByID(ctx context.Context, id int64) (domain.Website, error)
	UpdateWebsiteActive(ctx context.Context, id int64, isActive bool) error
	DeleteWebsite(ctx context.Context, id int64) error
	GetWebsiteStats(ctx context.Context, websiteID int64) (domain.WebsiteStats, error)
}
