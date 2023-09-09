package api

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

type API interface {
	CheckUser(ctx context.Context, kthID, system, permission string) (bool, error)
	ListForUser(ctx context.Context, kthID, system string) ([]string, error)
	CheckToken(ctx context.Context, secret uuid.UUID, system, permission string) (bool, error)
}

type service struct {
	db *sql.DB
}

func New(db *sql.DB) API {
	return &service{db}
}
