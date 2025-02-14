package main

import (
	"github.com/joho/godotenv"
	"github.com/quangtran666/simple-social-golang/internal/env"
	"log"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file, error: %v", err)
	}

	cfg := config{
		addr: env.GetString("ADDR", ":8080"),
	}

	app := &application{
		config: cfg,
	}

	mux := app.mount()
	log.Fatal(app.run(mux))
}
