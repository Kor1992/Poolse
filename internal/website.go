package service

import (
	"context"
	"errors"
	"pulse/internal/domain"
	"pulse/internal/repository"
	"strings"
)

type WebsiteService struct {
	repo repository.WebsiteRepository
}

var ErrInvalidURL = errors.New("invalid website URL")

func NewWebsiteService(
	repo repository.WebsiteRepository,
) *WebsiteService {
	return &WebsiteService{
		repo: repo,
	}
}

func (s *WebsiteService) CreateWebsite(
	ctx context.Context,
	url string,
) (int64, error) {
	url = strings.TrimSpace(url)
	if url == "" {
		return 0, ErrInvalidURL
	}

	return s.repo.AddWebsite(ctx, url)
}

func (s *WebsiteService) GetActiveWebsites(ctx context.Context) ([]domain.Website, error) {

	return s.repo.GetActiveWebsites(ctx)
}

var (
	ErrInvalidID = errors.New("id must be positive")
)

func (s *WebsiteService) GetWebsiteChecks(
	ctx context.Context,
	websiteID int64,
	limit int,
) ([]domain.CheckResult, error) {
	if websiteID <= 0 {
		return nil, ErrInvalidID
	}

	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	return s.repo.GetWebsiteChecks(ctx, websiteID, limit)

}

func (s *WebsiteService) GetWebsiteByID(
	ctx context.Context,
	id int64,
) (domain.Website, error) {
	if id <= 0 {
		return domain.Website{}, ErrInvalidID
	}

	return s.repo.GetWebsiteByID(ctx, id)
}

func (s *WebsiteService) UpdateWebsiteActive(
	ctx context.Context,
	id int64,
	isActive bool,
) error {
	if id <= 0 {
		return ErrInvalidID
	}

	return s.repo.UpdateWebsiteActive(ctx, id, isActive)
}

func (s *WebsiteService) DeleteWebsite(
	ctx context.Context,
	id int64,
) error {
	if id <= 0 {
		return ErrInvalidID
	}

	return s.repo.DeleteWebsite(ctx, id)
}
