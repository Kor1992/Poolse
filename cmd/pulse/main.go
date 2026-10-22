package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"pulse/internal/repository"
	"time"
)

func main() {
	dataBase := os.Getenv("DATABASE_URL")
	if dataBase == "" {
		dataBase = "postgres://postgres:postgres@localhost:5433/avito?sslmode=disable"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := repository.NewPostgres(ctx, dataBase)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer pool.Close()

	fmt.Println("data base connected")

}
