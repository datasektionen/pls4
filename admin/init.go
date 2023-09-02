package admin

import (
	"database/sql"
	"embed"
	"flag"
	"html/template"
	"os"

	"github.com/datasektionen/pls4/database"
)

//go:embed templates/*.html
var templates embed.FS

type service struct {
	db          *sql.DB
	t           *template.Template
	loginURL    string
	loginAPIKey string
	hodisURL    string
	sessions    map[string]session
}

var s service

func init() {
	flag.StringVar(&s.loginURL, "login-url",
		os.Getenv("LOGIN_URL"),
		"URL to login",
	)
	flag.StringVar(&s.loginAPIKey, "login-api-key",
		os.Getenv("LOGIN_API_KEY"),
		"API token for login. Funnily enough this service verifies the token",
	)
	flag.StringVar(&s.hodisURL, "hodis-url",
		os.Getenv("HODIS_URL"),
		"API token for login. Funnily enough this service verifies the token",
	)

	var err error
	s.t, err = template.ParseFS(templates, "templates/*.html")
	if err != nil {
		panic(err)
	}

	s.db = database.DB
	s.sessions = make(map[string]session)
}
