package admin

import (
	"log/slog"
	"net/http"
)

func Mount(admin Admin) {
	http.HandleFunc("/", route(admin, index))
	http.HandleFunc("/login", route(admin, login))
	http.HandleFunc("/logout", route(admin, logout))
}

func route(admin Admin, handler func(t Admin, w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler(admin, w, r)
	}
}

func index(admin Admin, w http.ResponseWriter, r *http.Request) {
	if err := admin.RenderIndex(w, IndexParameters{
		UserName: admin.LoggedInName(r),
	}); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("Could not render template", "error", err, "template", "index.html")
	}
}

func login(admin Admin, w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	sessionToken, err := admin.Login(code)
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

func logout(admin Admin, w http.ResponseWriter, r *http.Request) {
	cookie, _ := r.Cookie("session")
	if cookie != nil {
		admin.DeleteSession(cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name:   "session",
		MaxAge: -1,
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
