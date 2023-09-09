package models

import "time"

type Role struct {
	ID           string
	DisplayName  string
	Description  string
	SubroleCount int
	MemberCount  int
}

type Member struct {
	KTHID      string
	Comment    string
	ModifiedBy string
	ModifiedAt time.Time
	StartDate  time.Time
	EndDate    time.Time
	Indirect   bool
}
