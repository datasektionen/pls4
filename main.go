package main

import (
	"flag"
	"log/slog"
	"net/http"
)

func main() {
	var address string
	flag.StringVar(&address, "address", "0.0.0.0:3000", "The address to listen to requests on")
	flag.Parse()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("You no have access >:^("))
	})

	slog.Info("Started", "address", address)
	slog.Error("Server crashed", "error", http.ListenAndServe(address, nil))
}
