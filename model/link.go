package model

import "gorm.io/gorm"

type ShortLink struct {
	gorm.Model
	ShortID     string `gorm:"type:varchar(20);uniqueIndex;not null"`
	OriginalURL string `gorm:"type:text;not null"`
}
