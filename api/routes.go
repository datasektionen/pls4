package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
)

func init() {
	http.HandleFunc("/api/check-user", checkUser)
	http.HandleFunc("/api/list-for-user", listForUser)
	http.HandleFunc("/api/check-token", checkToken)
}

func checkUser(w http.ResponseWriter, r *http.Request) {
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
	ok, err := CheckUser(ctx, body.KTHID, body.System, body.Permission)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.ErrorContext(ctx, "Error checking user permission", "error", err)
		return
	}
	if err := json.NewEncoder(w).Encode(ok); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.ErrorContext(ctx, "Error writing body", "error", err)
		return
	}
}

func listForUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var body struct {
		KTHID  string `json:"kth_id"`
		System string `json:"system"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	perms, err := ListForUser(ctx, body.KTHID, body.System)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.ErrorContext(ctx, "Error listing permissions for user", "error", err)
		return
	}
	if err := json.NewEncoder(w).Encode(perms); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.ErrorContext(ctx, "Error writing body", "error", err)
		return
	}
}

func checkToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var body struct {
		Secret     uuid.UUID `json:"secret"`
		System     string    `json:"system"`
		Permission string    `json:"permission"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	perms, err := CheckToken(ctx, body.Secret, body.System, body.Permission)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.ErrorContext(ctx, "Error checking token permission", "error", err)
		return
	}
	if err := json.NewEncoder(w).Encode(perms); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.ErrorContext(ctx, "Error writing body", "error", err)
		return
	}
}
