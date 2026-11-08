package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"pulse/internal/domain"
	"pulse/internal/repository"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgres(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, err
	}
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("database is not reachable: %w", err)
	}

	return pool, nil
}

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{
		pool: pool,
	}
}

func (r *PostgresRepository) AddWebsite(
	ctx context.Context,
	url string,
) (int64, error) {
	query := `
	INSERT INTO websites (url)
VALUES ($1)
RETURNING id;
	`

	var id int64

	err := r.pool.QueryRow(ctx, query, url).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("add website %q: %w", url, err)
	}
	return id, nil
}

func (r *PostgresRepository) GetActiveWebsites(
	ctx context.Context,
) ([]domain.Website, error) {
	query := `
	SELECT id, url
FROM websites
WHERE is_active = TRUE
ORDER BY id;
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	websites := make([]domain.Website, 0)

	for rows.Next() {
		var website domain.Website
		err := rows.Scan(&website.ID, &website.URL)
		if err != nil {
			return nil, err
		}
		websites = append(websites, website)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return websites, nil
}

func (r *PostgresRepository) SaveCheck(
	ctx context.Context,
	websiteID int64,
	result domain.CheckResult,
) error {
	query := `
	INSERT INTO checks (
    website_id,
    status_code,
    duration_ms,
    is_up,
    error,
    checked_at
)
VALUES ($1, $2, $3, $4, $5, $6);
	`
	_, err := r.pool.Exec(ctx, query, websiteID, result.StatusCode, result.Duration.Milliseconds(), result.IsUp, result.Error, result.CheckedAt)
	if err != nil {
		return fmt.Errorf("save check for website %d: %w", websiteID, err)
	}
	return nil
}

func (r *PostgresRepository) GetWebsiteChecks(
	ctx context.Context,
	websiteID int64,
	limit int,
) ([]domain.CheckResult, error) {
	query := `
		SELECT status_code, duration_ms, is_up, error, checked_at
		FROM checks
		WHERE website_id = $1
		ORDER BY checked_at DESC
		LIMIT $2;
	`

	rows, err := r.pool.Query(ctx, query, websiteID, limit)
	if err != nil {
		return nil, fmt.Errorf("get checks for website %d: %w", websiteID, err)
	}
	defer rows.Close()

	checks := make([]domain.CheckResult, 0)

	for rows.Next() {
		var check domain.CheckResult
		var durationMs int64
		var statusCode sql.NullInt32
		var errorText sql.NullString

		err := rows.Scan(
			&statusCode,
			&durationMs,
			&check.IsUp,
			&errorText,
			&check.CheckedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan website check: %w", err)
		}

		check.WebsiteID = websiteID
		check.Duration = time.Duration(durationMs) * time.Millisecond

		if statusCode.Valid {
			check.StatusCode = int(statusCode.Int32)
		}

		if errorText.Valid {
			check.Error = errorText.String
		}

		checks = append(checks, check)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate website checks: %w", err)
	}

	return checks, nil
}

func (r *PostgresRepository) GetWebsiteByID(
	ctx context.Context,
	id int64,
) (domain.Website, error) {
	query := `
    SELECT id, url
    FROM websites
    WHERE id = $1;
`

	var website domain.Website

	err := r.pool.QueryRow(ctx, query, id).Scan(&website.ID, &website.URL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Website{}, repository.ErrNotFound
		}
		return domain.Website{}, fmt.Errorf(
			"get website %d: %w",
			id,
			err,
		)
	}

	return website, nil

}

func (r *PostgresRepository) UpdateWebsiteActive(
	ctx context.Context,
	id int64,
	isActive bool,
) error {
	query := `
		UPDATE websites
		SET is_active = $2
		WHERE id = $1
		RETURNING id;
	`

	var updatedID int64
	err := r.pool.QueryRow(ctx, query, id, isActive).Scan(&updatedID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return repository.ErrNotFound
		}
		return fmt.Errorf("update website %d active status: %w", id, err)
	}

	return nil
}

func (r *PostgresRepository) DeleteWebsite(
	ctx context.Context,
	id int64,
) error {
	query := `
		DELETE FROM websites
		WHERE id = $1
		RETURNING id;
	`

	var deletedID int64
	err := r.pool.QueryRow(ctx, query, id).Scan(&deletedID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return repository.ErrNotFound
		}
		return fmt.Errorf("delete website %d: %w", id, err)
	}

	return nil
}

func (r *PostgresRepository) GetWebsiteStats(
	ctx context.Context,
	websiteID int64,
) (domain.WebsiteStats, error) {

	var exists bool

	err := r.pool.QueryRow(ctx, `
    SELECT EXISTS(
        SELECT 1 FROM websites WHERE id = $1
    )
`, websiteID).Scan(&exists)

	if err != nil {
		return domain.WebsiteStats{}, fmt.Errorf(
			"check website existence: %w", err,
		)
	}
	if !exists {
		return domain.WebsiteStats{}, repository.ErrNotFound
	}

	query := `
	SELECT
    COUNT(*) AS total_checks,
    COUNT(*) FILTER (WHERE is_up = TRUE) AS successful_checks,
    COUNT(*) FILTER (WHERE is_up = FALSE) AS failed_checks,
    COALESCE(AVG(duration_ms), 0) AS avg_response_ms
	FROM checks
	WHERE website_id = $1
 	AND checked_at >= NOW() - INTERVAL '24 hours';
	`

	var website domain.WebsiteStats

	err = r.pool.QueryRow(ctx, query, websiteID).Scan(&website.TotalChecks,
		&website.SuccessfulChecks,
		&website.FailedChecks,
		&website.AvgResponseMs,
	)
	if err != nil {
		return domain.WebsiteStats{}, fmt.Errorf(
			"get website %d stats: %w",
			websiteID,
			err,
		)
	}

	website.WebsiteID = websiteID
	return website, nil
}
