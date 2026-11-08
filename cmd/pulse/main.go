package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"pulse/internal/handler"
	"pulse/internal/monitor"
	"pulse/internal/repository/postgres"
	"pulse/internal/service"
)

func main() {
	dataBase := os.Getenv("DATABASE_URL")
	if dataBase == "" {
		dataBase = "postgres://postgres:postgres@localhost:5433/pulse?sslmode=disable"
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	connectCtx, cancelConnect := context.WithTimeout(ctx, 5*time.Second)
	pool, err := postgres.NewPostgres(connectCtx, dataBase)
	cancelConnect()
	if err != nil {
		log.Printf("database connection failed: %v", err)
		return
	}
	defer pool.Close()

	repo := postgres.NewPostgresRepository(pool)
	client := &http.Client{Timeout: 5 * time.Second}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		monitor.RunMonitor(ctx, repo, 30*time.Second, 5, client)
	}()
	websiteService := service.NewWebsiteService(repo)
	websiteHandler := handler.NewWebHandler(websiteService)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/websites", websiteHandler.CreateWebsite)
	mux.HandleFunc("GET /api/websites", websiteHandler.GetActiveWebsites)
	mux.HandleFunc("GET /api/websites/{id}/checks", websiteHandler.GetWebsiteChecks)
	mux.HandleFunc("GET /api/websites/{id}", websiteHandler.GetWebsiteByID)
	mux.HandleFunc("PATCH /api/websites/{id}", websiteHandler.UpdateWebsiteActive)
	mux.HandleFunc("DELETE /api/websites/{id}", websiteHandler.DeleteWebsite)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	serverErr := make(chan error, 1)
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	log.Print("server start on", server.Addr)

	select {
	case <-ctx.Done():
	case err := <-serverErr:
		log.Printf("HTTP server error: %v", err)
		stop()
	}

	shutdownCtx, cancelShutdown := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancelShutdown()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown failed: %v", err)
	}

	wg.Wait()
}
