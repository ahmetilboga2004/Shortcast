package model

import "gorm.io/gorm"

type User struct {
	gorm.Model
	FirstName string `gorm:"type:varchar(100);not null"`
	LastName  string `gorm:"type:varchar(100);not null"`
	Username  string `gorm:"type:varchar(100);uniqueIndex;not null"`
	Email     string `gorm:"type:varchar(100);uniqueIndex;not null"`
	Password  string `gorm:"type:varchar(255);not null"`
	Podcasts []Podcast `gorm:"foreignKey:UserID"`
}
