package model

import "gorm.io/gorm"

type Like struct {
	gorm.Model
	PodcastID uint    `gorm:"not null;index"`
	UserID    uint    `gorm:"not null;index"`
	User      User    `gorm:"foreignKey:UserID"`
	Podcast   Podcast `gorm:"foreignKey:PodcastID"`
} 