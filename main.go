package main

import (
	"context"
	"database/sql"
	"flag"
	"log/slog"
	"net/http"
	"os"

	_ "github.com/lib/pq"

	"github.com/datasektionen/pls4/admin"
	"github.com/datasektionen/pls4/api"
	"github.com/datasektionen/pls4/database"
)

func main() {
	var address, loginURL, loginAPIKey, hodisURL, databaseURL string
	flag.StringVar(&address, "address",
		envOr("ADDRESS", "0.0.0.0:3000"),
		"The address to listen to requests on",
	)
	flag.StringVar(&loginURL, "login-url",
		os.Getenv("LOGIN_URL"),
		"URL to login",
	)
	flag.StringVar(&loginAPIKey, "login-api-key",
		os.Getenv("LOGIN_API_KEY"),
		"API token for login. Funnily enough this service verifies the token",
	)
	flag.StringVar(&hodisURL, "hodis-url",
		os.Getenv("HODIS_URL"),
		"API token for login. Funnily enough this service verifies the token",
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

	if err := database.Migrate(db, ctx); err != nil {
		panic(err)
	}

	apiService := api.New(db)
	adminService, err := admin.New(db, loginURL, loginAPIKey, hodisURL)
	if err != nil {
		panic(err)
	}

	api.Mount(apiService)
	admin.Mount(adminService)

	slog.Info("Started", "address", address)
	slog.Error("Server crashed", "error", http.ListenAndServe(address, nil))
}

func envOr(key string, fallback string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	return value
}
