package monitor

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type CheckResult struct {
	URL        string
	StatusCode int
	Duration   time.Duration
	IsUp       bool
	Error      string
	CheckedAt  time.Time
}

func (r *CheckResult) String() string {
	return fmt.Sprintf(
		"URL: %s | UP: %t | Status: %d | Duration: %s | Error: %s",
		r.URL, r.IsUp, r.StatusCode, r.Duration, r.Error,
	)
}

func checkWebsite(ctx context.Context, url string, client *http.Client) CheckResult {
	result := CheckResult{URL: url, CheckedAt: time.Now()}
	request, err := http.NewRequestWithContext(ctx, "GET", url, nil)

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

func checkAll(ctx context.Context, urls []string, maxConcurrent int, client *http.Client) []CheckResult {
	if maxConcurrent <= 0 {
		return nil
	}

	workers := min(maxConcurrent, len(urls))

	urlsChan := make(chan string, len(urls))
	results := make([]CheckResult, 0, len(urls))
	var mx sync.Mutex

	go func() {
		defer close(urlsChan)
		for _, url := range urls {
			urlsChan <- url
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
				case url, ok := <-urlsChan:
					if !ok {
						return
					}
					if ctx.Err() != nil {
						return
					}
					result := checkWebsite(ctx, url, client)
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
