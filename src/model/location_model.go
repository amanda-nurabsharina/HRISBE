package model

import "time"

type Location struct {
	ID         string    `gorm:"primaryKey;type:varchar(50)" json:"id"`
	Name       string    `gorm:"type:varchar(255);not null" json:"name"`
	Department string    `gorm:"type:varchar(255);not null" json:"department"`
	Latitude   float64   `gorm:"type:double precision;not null" json:"latitude"`
	Longitude  float64   `gorm:"type:double precision;not null" json:"longitude"`
	Radius     float64   `gorm:"type:double precision;not null" json:"radius"` // in meters
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}
