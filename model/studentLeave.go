package model

type StudentLeave struct {
	//gorm.Model
	//SID    string `gorm:"type:varchar(11);not null;unique"`
	//Name   string `gorm:"type:varchar(20);not null"`
	//SClass string `gorm:"type:varchar(6);not null"`
	//Date   string `gorm:"type:varchar(20);not null"`
	//Reason string `gorm:"type:varchar(255);not null"`

	SID        string `redis:"sid" json:"SID"`
	Name       string `redis:"name" json:"Name"`
	SClass     string `redis:"sClass" json:"SClass"`
	Date       string `redis:"date" json:"Date"`
	Reason     string `redis:"reason" json:"Reason"`
	IsApproved string `redis:"isApproved" json:"IsApproved"`
}
