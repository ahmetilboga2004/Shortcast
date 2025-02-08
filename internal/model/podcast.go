package model

import "gorm.io/gorm"

type Podcast struct {
	gorm.Model
	Title    string `gorm:"type:varchar(255);not null"`
	Category string `gorm:"type:varchar(100);not null"`
	AudioURL string `gorm:"type:varchar(255);not null"`
	CoverURL string `gorm:"type:varchar(255);not null"`
	UserID   uint   `gorm:"not null;index"`
	User     User   `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}
