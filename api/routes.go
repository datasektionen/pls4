package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

func Mount(api *API) {
	http.Handle("/api/user/get-permissions", route(api, userGetPermissions))
	http.Handle("/api/user/check", route(api, userCheckPermission))
	http.Handle("/api/user/get-scopes", route(api, userGetScopes))
}

func route(api *API, handler func(api *API, w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler(api, w, r)
	}
}

func userGetPermissions(api *API, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var body struct {
		KTHID      string `json:"kth_id"`
		System     string `json:"system"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	permissions, err := api.UserGetPermissions(ctx, body.KTHID, body.System)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.ErrorContext(ctx, "Error checking user permission", "error", err)
		return
	}
	if err := json.NewEncoder(w).Encode(permissions); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.ErrorContext(ctx, "Error writing body", "error", err)
		return
	}
}

func userCheckPermission(api *API, w http.ResponseWriter, r *http.Request) {
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
	ok, err := api.UserCheckPermission(ctx, body.KTHID, body.System, body.Permission)
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

func userGetScopes(api *API, w http.ResponseWriter, r *http.Request) {
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
	ok, err := api.UserGetScopes(ctx, body.KTHID, body.System, body.Permission)
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
