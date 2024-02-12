package models

import (
	"time"

	"github.com/google/uuid"
)

type Role struct {
	ID           string
	DisplayName  string
	Description  string
	SubroleCount int
	MemberCount  int
}

type Member struct {
	MemberID   uuid.UUID
	KTHID      string
	ModifiedBy string
	ModifiedAt time.Time
	StartDate  time.Time
	EndDate    time.Time
}

type SystemPermissions struct {
	System      string
	Permissions []string
}
