package models

import "time"

type Department struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"not null"`
	ParentID  *int      `json:"parent_id"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`

	Employees []Employee   `json:"employees,omitempty" gorm:"foreignKey:ParentID"`
	Children  []Department `json:"children,omitempty" gorm:"foreignKey:ParentID"`
}
