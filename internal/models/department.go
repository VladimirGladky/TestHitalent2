package models

import "time"

type Department struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"not null"`
	ParentID  *int      `json:"parent_id"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}
