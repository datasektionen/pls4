package main

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	_ "github.com/lib/pq"

	"github.com/datasektionen/pls4/api"
	"github.com/datasektionen/pls4/database"
	uiService "github.com/datasektionen/pls4/ui/service"
	uiViews "github.com/datasektionen/pls4/ui/views"
)

func main() {
	address := getenv("ADDRESS", "0.0.0.0:3000")
	loginFrontendURL := getenv("LOGIN_FRONTEND_URL")
	loginAPIURL := getenv("LOGIN_API_URL")
	loginAPIKey := getenv("LOGIN_API_KEY") // "API token for login. Funnily enough this service verifies the token",
	databaseURL := getenv("DATABASE_URL")

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		panic(err)
	}

	{
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		if err := database.Migrate(db, ctx); err != nil {
			panic(err)
		}
		cancel()
	}

	ctx, cancel := context.WithCancel(context.Background())

	apiService := api.New(db)
	uiService := uiService.New(ctx, db, apiService, loginFrontendURL, loginAPIURL, loginAPIKey)

	mux := http.NewServeMux()
	api.Mount(mux, apiService)
	uiViews.Mount(mux, uiService)

	server := http.Server{Addr: address, Handler: mux}

	beginShutdown := make(chan os.Signal)
	signal.Notify(beginShutdown, os.Interrupt)
	shutdownComplete := make(chan struct{})
	go func() {
		<-beginShutdown
		cancel()
		server.Shutdown(context.Background())
		shutdownComplete <- struct{}{}
	}()
	slog.Info("Started", "address", address)
	if err := server.ListenAndServe(); errors.Is(err, http.ErrServerClosed) {
		slog.Info("Shutting down...")
		<-shutdownComplete
		slog.Info("Clean shutdown finished")
	} else {
		slog.Error("Server crashed", "error", err)
	}
}

func getenv(key string, fallback ...string) string {
	if len(fallback) > 1 {
		panic("Supplied multiple fallbacks")
	}
	value, ok := os.LookupEnv(key)
	if !ok {
		if len(fallback) == 0 {
			panic("Missing required environment variable $" + key)
		}
		return fallback[0]
	}
	return value
}
