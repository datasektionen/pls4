package main

import (
	"context"
	"database/sql"
	"flag"
	"log/slog"
	"net/http"
	"os"

	_ "github.com/lib/pq"

	"github.com/datasektionen/pls4/database"
	"github.com/datasektionen/pls4/handlers"
	"github.com/datasektionen/pls4/routes"
)

func envOr(key string, fallback string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	return value
}

func main() {
	var (
		address     string
		databaseURL string
	)
	flag.StringVar(&address, "address",
		envOr("ADDRESS", "0.0.0.0:3000"),
		"The address to listen to requests on",
	)
	flag.StringVar(&databaseURL, "database-url",
		os.Getenv("DATABASE_URL"),
		"URL to the postgresql database to use",
	)
	flag.Parse()

	ctx := context.Background()

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		panic(err)
	}

	if err := database.Migrate(ctx, db); err != nil {
		panic(err)
	}

	s := handlers.NewService(db)

	routes.Mount(s)

	slog.Info("Started", "address", address)
	slog.Error("Server crashed", "error", http.ListenAndServe(address, nil))
}
