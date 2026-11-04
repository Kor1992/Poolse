package monitor

import (
	"context"
	"net/http"
	"sync"
	"time"

	"pulse/internal/domain"
)

func checkWebsite(ctx context.Context, website domain.Website, client *http.Client) domain.CheckResult {
	result := domain.CheckResult{
		WebsiteID: website.ID,
		URL:       website.URL,
		CheckedAt: time.Now(),
	}
	request, err := http.NewRequestWithContext(ctx, "GET", website.URL, nil)

	if err != nil {
		result.Error = err.Error()
		return result
	}
	resp, err := client.Do(request)
	result.Duration = time.Since(result.CheckedAt)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	defer resp.Body.Close()
	dostup := false
	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		dostup = true
	}
	result.StatusCode = resp.StatusCode
	result.IsUp = dostup

	return result

}

func checkAll(ctx context.Context, websites []domain.Website, maxConcurrent int, client *http.Client) []domain.CheckResult {
	if maxConcurrent <= 0 {
		return nil
	}

	workers := min(maxConcurrent, len(websites))

	websitesChan := make(chan domain.Website, len(websites))
	results := make([]domain.CheckResult, 0, len(websites))
	var mx sync.Mutex

	go func() {
		defer close(websitesChan)
		for _, website := range websites {
			websitesChan <- website
		}
	}()
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case website, ok := <-websitesChan:
					if !ok {
						return
					}
					if ctx.Err() != nil {
						return
					}
					result := checkWebsite(ctx, website, client)
					mx.Lock()
					results = append(results, result)
					mx.Unlock()
				}

			}
		}()
	}

	wg.Wait()

	return results

}
