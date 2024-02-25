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

type SystemPermissionInstances struct {
	System      string
	Permissions []PermissionInstance
}

type PermissionInstance struct {
	ID           uuid.UUID
	PermissionID string
	Scope        string
}

type Permission struct {
	ID       string
	HasScope bool
}
