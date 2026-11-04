package monitor

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"pulse/internal/domain"
)

func TestCheckAll(t *testing.T) {
	var active atomic.Int64
	var maximum atomic.Int64

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		current := active.Add(1)
		defer active.Add(-1)

		for observed := maximum.Load(); current > observed; observed = maximum.Load() {
			if maximum.CompareAndSwap(observed, current) {
				break
			}
		}

		time.Sleep(75 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	websites := make([]domain.Website, 10)
	expectedURLs := make(map[int64]string, len(websites))
	for i := range websites {
		websiteID := int64(i + 1)
		websites[i] = domain.Website{ID: websiteID, URL: server.URL}
		expectedURLs[websiteID] = server.URL
	}

	results := checkAll(context.Background(), websites, 3, server.Client())
	if len(results) != len(websites) {
		t.Fatalf("получено %d результатов, ожидалось %d", len(results), len(websites))
	}
	seenIDs := make(map[int64]bool, len(results))
	for i, result := range results {
		if !result.IsUp {
			t.Errorf("результат %d для %q имеет IsUp == false: %+v", i, result.URL, result)
		}
		expectedURL, ok := expectedURLs[result.WebsiteID]
		if !ok {
			t.Errorf("результат %d содержит неизвестный WebsiteID %d", i, result.WebsiteID)
		} else if result.URL != expectedURL {
			t.Errorf("для WebsiteID %d получен URL %q, ожидался %q", result.WebsiteID, result.URL, expectedURL)
		}
		if seenIDs[result.WebsiteID] {
			t.Errorf("WebsiteID %d встречается в результатах повторно", result.WebsiteID)
		}
		seenIDs[result.WebsiteID] = true
	}
	if got := maximum.Load(); got > 3 {
		t.Errorf("одновременно обрабатывалось %d запросов, максимум должен быть 3", got)
	}
}
