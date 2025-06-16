package models

import "gorm.io/gorm"

type UserInfo struct {
	gorm.Model
	Name     string `form:"username" gorm:"type:varchar(20);unique;not null"`
	Password string `form:"password"`
}

func (UserInfo) TableName() string {
	return "users"
}
