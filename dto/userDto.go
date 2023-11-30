package dto

import "dning.com/pro02/model"

type StudentDto struct {
	Name         string
	SID          string
	NormalCount  string
	LeaveCount   string
	LateCount    string
	EarlyCount   string
	AbsenceCount string
}

func ToStudentDto(student model.Student) StudentDto {
	return StudentDto{
		Name:         student.Name,
		SID:          student.SID,
		NormalCount:  student.NormalCount,
		LeaveCount:   student.LeaveCount,
		LateCount:    student.LateCount,
		EarlyCount:   student.EarlyCount,
		AbsenceCount: student.AbsenceCount,
	}
}
