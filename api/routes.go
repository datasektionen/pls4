package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
)

func Mount(api *API) {
	http.Handle("/api/user/check", route(api, checkUser))
	http.Handle("/api/user/filter", route(api, filterForUser))
	http.Handle("/api/user/raw", route(api, getUserRaw))
	http.Handle("/api/token/check", route(api, checkToken))
}

func route(api *API, handler func(api *API, w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler(api, w, r)
	}
}

func checkUser(api *API, w http.ResponseWriter, r *http.Request) {
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
	ok, err := api.CheckUser(ctx, body.KTHID, body.System, body.Permission)
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

func filterForUser(api *API, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var body struct {
		KTHID       string   `json:"kth_id"`
		System      string   `json:"system"`
		Permissions []string `json:"permissions"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	granted, err := api.FilterForUser(ctx, body.KTHID, body.System, body.Permissions)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.ErrorContext(ctx, "Error checking user permission", "error", err)
		return
	}
	if err := json.NewEncoder(w).Encode(granted); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.ErrorContext(ctx, "Error writing body", "error", err)
		return
	}
}

func checkToken(api *API, w http.ResponseWriter, r *http.Request) {
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
	perms, err := api.CheckToken(ctx, body.Secret, body.System, body.Permission)
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

func getUserRaw(api *API, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var body struct {
		KTHID  string `json:"kth_id"`
		System string `json:"system"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	perms, err := api.GetUserRaw(ctx, body.KTHID, body.System)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.ErrorContext(ctx, "Error getting permissions for user", "error", err)
		return
	}
	if err := json.NewEncoder(w).Encode(perms); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.ErrorContext(ctx, "Error writing body", "error", err)
		return
	}
}
