package main

import (
	"log"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/quangtran666/simple-social-golang/internal/db"
	"github.com/quangtran666/simple-social-golang/internal/env"
	"github.com/quangtran666/simple-social-golang/internal/store"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file, error: %v", err)
	}

	add := env.GetString("DB_ADDR", "postgresql://postgres:postgres@localhost:5432/simple_social?sslmode=disable")
	maxOpenConns := env.GetInt("DB_MAX_OPEN_CONNS", 30)
	maxIdleConns := env.GetInt("DB_MAX_IDLE_CONNS", 10)
	maxIdleTime := env.GetString("DB_MAX_IDLE_TIME", "15m")

	conn, err := db.New(add, maxOpenConns, maxIdleConns, maxIdleTime)

	if err != nil {
		log.Fatalf("Error connecting to database, error: %v", err)
	}

	store := store.NewStorage(conn)
	db.Seed(store, conn)
}
