package model

import "gorm.io/gorm"

type Comment struct {
	gorm.Model
	Content   string  `gorm:"type:text;not null"`
	PodcastID uint    `gorm:"not null;index"`
	UserID    uint    `gorm:"not null;index"`
	User      User    `gorm:"foreignKey:UserID"`
	Podcast   Podcast `gorm:"foreignKey:PodcastID"`
} 