package admin

import (
	"log/slog"
	"net/http"
)

func init() {
	http.HandleFunc("/", index)
	http.HandleFunc("/login", login)
}

func index(w http.ResponseWriter, r *http.Request) {
	if err := s.t.ExecuteTemplate(w, "index.html", map[string]any{
		"login_url": s.loginURL,
		"user_name": LoggedInName(r),
	}); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("Could not render template", "error", err, "template", "index.html")
	}
}

func login(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	sessionToken, err := Login(code)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("Could not verify login code", "error", err)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    sessionToken,
		MaxAge:   60 * 60,
		Secure:   false,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
