package service

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

type Session struct {
	KTHID       string
	DisplayName string
}

func (ui *UI) deleteOldSessionsForever() {
	// TODO: context?
	for {
		if _, err := ui.db.Exec(`
			delete from sessions
			where last_used_at <= now() - interval '1' hour
		`); err != nil {
			slog.Error("Could not delete old sessions", "error", err)
		}
		time.Sleep(time.Hour)
	}
}

func (ui *UI) Login(code string) (string, error) {
	res, err := http.Get(ui.loginAPIURL + "/verify/" + code + "?api_key=" + ui.loginAPIKey)
	if err != nil {
		return "", err
	}
	var body struct {
		KTHID     string `json:"user"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		return "", err
	}
	r := ui.db.QueryRow(`
		insert into sessions (kth_id, display_name, last_used_at)
		values ($1, $2, now())
		returning id
	`, body.KTHID, body.FirstName+" "+body.LastName)
	var id string
	if err := r.Scan(&id); err != nil {
		return "", err
	}
	return id, nil
}

func (ui *UI) DeleteSession(sessionID string) error {
	_, err := ui.db.Exec(`--sql
		delete from sessions
		where id = $1
	`, sessionID)
	return err
}

// Returns the kth id and display name of the logged in user, or the zero value if no user is
// logged in.
func (ui *UI) GetSession(r *http.Request) (Session, error) {
	cookie, _ := r.Cookie("session")
	if cookie == nil {
		return Session{}, nil
	}
	id := cookie.Value
	tx, err := ui.db.BeginTx(r.Context(), nil)
	defer tx.Rollback()
	if err != nil {
		return Session{}, err
	}
	row := tx.QueryRow(`--sql
		select kth_id, display_name
		from sessions
		where id = $1
		and last_used_at > now() - interval '1' hour
	`, id)
	var session Session
	if err := row.Scan(
		&session.KTHID,
		&session.DisplayName,
	); err == sql.ErrNoRows {
		return Session{}, nil
	} else if err != nil {
		slog.ErrorContext(r.Context(), "Could not get session from database", "id", id, "error", err)
		return Session{}, err
	}
	if _, err := tx.Exec(`--sql
		update sessions
		set last_used_at = now()
		where id = $1
	`, id); err != nil {
		return Session{}, err
	}
	return session, tx.Commit()
}
