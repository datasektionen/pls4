package models

type Role struct {
	ID           string
	DisplayName  string
	Description  string
	SubroleCount int
	MemberCount  int
}
