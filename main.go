package main

import (
	"context"
	"database/sql"
	"flag"
	"log/slog"
	"net/http"

	_ "github.com/lib/pq"

	"github.com/datasektionen/pls4/database"
)

func main() {
	var (
		address     string
		databaseURL string
	)
	flag.StringVar(&address, "address", "0.0.0.0:3000", "The address to listen to requests on")
	flag.StringVar(&databaseURL, "database-url", "postgresql://pls4:pls4@localhost:5432/pls4?sslmode=disable", "URL to the postgresql database to use")
	flag.Parse()

	ctx := context.Background()

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		panic(err)
	}

	if err := database.Migrate(ctx, db); err != nil {
		panic(err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("You no have access >:^("))
	})

	slog.Info("Started", "address", address)
	slog.Error("Server crashed", "error", http.ListenAndServe(address, nil))
}
