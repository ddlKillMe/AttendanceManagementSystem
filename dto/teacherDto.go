package dto

import "dning.com/pro02/model"

type TeacherDto struct {
	Name string
	TID  string
}

func ToTeacherDto(teacher model.Teacher) TeacherDto {
	return TeacherDto{
		Name: teacher.Name,
		TID:  teacher.TID,
	}
}
