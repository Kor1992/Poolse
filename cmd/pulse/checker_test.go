package main

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
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

	urls := make([]string, 10)
	for i := range urls {
		urls[i] = server.URL
	}

	results := checkAll(urls, 3)
	if len(results) != len(urls) {
		t.Fatalf("получено %d результатов, ожидалось %d", len(results), len(urls))
	}
	for i, result := range results {
		if !result.IsUp {
			t.Errorf("результат %d для %q имеет IsUp == false: %+v", i, result.URL, result)
		}
	}
	if got := maximum.Load(); got > 3 {
		t.Errorf("одновременно обрабатывалось %d запросов, максимум должен быть 3", got)
	}
}
