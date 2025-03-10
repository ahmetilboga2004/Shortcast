package model

import "gorm.io/gorm"

type Podcast struct {
	gorm.Model
	Title    string `gorm:"type:varchar(255);not null"`
	Category string `gorm:"type:varchar(100);not null"`
	AudioKey string `gorm:"type:varchar(255);not null"`
	CoverKey string `gorm:"type:varchar(255);not null"`
	UserID   uint   `gorm:"not null;index"`
	User     User   `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}
