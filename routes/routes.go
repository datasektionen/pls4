package routes

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/datasektionen/pls4/handlers"
)

func Mount(s *handlers.Service) {
	http.HandleFunc("/", index)
	http.HandleFunc("/api/check-user", checkUser(s))
}

func index(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("You no have access >:^("))
}

func checkUser(s *handlers.Service) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		var body struct {
			KTHID      string `json:"kth_id"`
			System     string `json:"system"`
			Permission string `json:"permission"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		ok, err := s.CheckUser(ctx, body.KTHID, body.System, body.Permission)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			slog.ErrorContext(ctx, "Error checking user permission", "error", err)
			return
		}
		if ok {
			w.Write([]byte("yes :)"))
		} else {
			w.Write([]byte("NO!"))
		}
	}
}
