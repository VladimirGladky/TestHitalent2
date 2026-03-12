package models

import "time"

type Employee struct {
	ID           int        `json:"id" gorm:"primaryKey"`
	DepartmentID int        `json:"department_id" gorm:"not null"`
	FullName     string     `json:"full_name" gorm:"not null"`
	Position     string     `json:"position" gorm:"not null"`
	HiredAt      *time.Time `json:"hired_at"`
	CreatedAt    time.Time  `json:"created_at" gorm:"autoCreateTime"`
}
