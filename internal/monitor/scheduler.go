package monitor

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

func RunMonitor(
	ctx context.Context,
	urls []string,
	interval time.Duration,
	maxConcurrent int,
	client *http.Client,
) {
	if err := ctx.Err(); err != nil {
		return
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	results := checkAll(ctx, urls, maxConcurrent, client)
	for _, result := range results {
		fmt.Println(result.String())
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if ctx.Err() != nil {
				return
			}

			results := checkAll(ctx, urls, maxConcurrent, client)
			for _, result := range results {
				fmt.Println(result.String())

			}
		}
	}
}
