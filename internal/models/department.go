package models

import "time"

type Department struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"not null" validate:"required,min=1,max=200"`
	ParentID  *int      `json:"parent_id"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`

	Employees []Employee    `json:"employees,omitempty" gorm:"foreignKey:DepartmentID"`
	Children  []*Department `json:"children,omitempty" gorm:"foreignKey:ParentID"`
}
