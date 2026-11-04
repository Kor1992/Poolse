package monitor

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"pulse/internal/repository"
)

func RunMonitor(
	ctx context.Context,
	repo repository.WebsiteRepository,
	interval time.Duration,
	maxConcurrent int,
	client *http.Client,
) {
	if err := ctx.Err(); err != nil {
		return
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		if err := runCycle(ctx, repo, maxConcurrent, client); err != nil {
			if ctx.Err() != nil {
				return
			}
			slog.Error("monitor cycle failed", "error", err)
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func runCycle(
	ctx context.Context,
	repo repository.WebsiteRepository,
	maxConcurrent int,
	client *http.Client,
) error {
	websites, err := repo.GetActiveWebsites(ctx)
	if err != nil {
		return fmt.Errorf("get active websites: %w", err)
	}

	var errs error

	result := checkAll(ctx, websites, maxConcurrent, client)
	for _, v := range result {
		err := repo.SaveCheck(ctx, v.WebsiteID, v)
		if err != nil {
			errs = errors.Join(errs, err)
			continue
		}
	}

	return errs
}
