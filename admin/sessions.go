package admin

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type session struct {
	kthID      string
	validUntil time.Time
}

func (s *service) deleteOldSessionsForever() {
	for {
		time.Sleep(time.Hour)
		for k, session := range s.sessions {
			if session.validUntil.Before(time.Now()) {
				delete(s.sessions, k)
			}
		}
	}
}

func (s *service) Login(code string) (string, error) {
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

func (s *service) DeleteSession(sessionID string) {
	delete(s.sessions, sessionID)
}

// Returns the kth id of the logged in user, or the empty string if no user is
// logged in.
func (s *service) LoggedInKTHID(r *http.Request) string {
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
func (s *service) LoggedInName(r *http.Request) string {
	kthID := s.LoggedInKTHID(r)
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
