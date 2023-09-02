package admin

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type session struct {
	kthID      string
	validUntil time.Time
}

func init() {
	go func() {
		for {
			time.Sleep(time.Hour)
			for k, session := range s.sessions {
				if session.validUntil.Before(time.Now()) {
					delete(s.sessions, k)
				}
			}
		}
	}()
}

func Login(code string) (string, error) {
	slog.Info("logging in", "login url", s.loginURL)
	res, err := http.Get(
		s.loginURL + "/verify/" + code + "?api_key=" + s.loginAPIKey,
	)
	if err != nil {
		return "", err
	}
	var body struct {
		KTHID string `json:"user"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		return "", err
	}
	id := uuid.NewString()
	s.sessions[id] = session{
		kthID:      body.KTHID,
		validUntil: time.Now().Add(time.Hour),
	}
	return id, nil
}

// Returns the kth id of the logged in user, or the empty string if no user is
// logged in.
func LoggedInKTHID(r *http.Request) string {
	cookie, err := r.Cookie("session")
	if err != nil {
		return ""
	}
	id := cookie.Value
	session, ok := s.sessions[id]
	if !ok {
		return ""
	}
	if session.validUntil.Before(time.Now()) {
		delete(s.sessions, id)
		return ""
	}
	return session.kthID
}

// Fetches the name of the logged in user from hodis. Falls back to the kth id
// if the name cannot be fetched. Returns the empty string when no user is
// logged in.
func LoggedInName(r *http.Request) string {
	kthID := LoggedInKTHID(r)
	if kthID == "" {
		return ""
	}
	res, err := http.Get(s.hodisURL + "/uid/" + kthID)
	if err != nil {
		return kthID
	}
	var body struct {
		Name string `json:"displayName"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		return kthID
	}
	return body.Name
}
