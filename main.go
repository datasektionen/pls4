package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"

	_ "github.com/datasektionen/pls4/admin"
	_ "github.com/datasektionen/pls4/api"
	"github.com/datasektionen/pls4/database"
)

func main() {
	var address string
	flag.StringVar(&address, "address",
		envOr("ADDRESS", "0.0.0.0:3000"),
		"The address to listen to requests on",
	)
	flag.Parse()

	ctx := context.Background()

	if err := database.Migrate(ctx); err != nil {
		panic(err)
	}

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
