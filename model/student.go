package model

type Student struct {
	Name         string `redis:"name"`
	SID          string `redis:"sid"`
	Password     string `redis:"password"`
	NormalCount  string `redis:"normalCount"`
	LeaveCount   string `redis:"leaveCount"`
	LateCount    string `redis:"lateCount"`
	EarlyCount   string `redis:"earlyCount"`
	AbsenceCount string `redis:"absenceCount"`
}
