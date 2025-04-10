package models

import "gorm.io/gorm"

type UserInfo struct {
	gorm.Model
	Name          string `gorm:"type:varchar(20)"`
	Password      string
	Phone         string
	Email         string
	Identity      string
	HeartBeatTime uint64
	ClientIP      string
	ClientPort    string
	LoginTime     uint64
	LogoutTime    uint64
	IsLogout      bool
	DeviceInfo    string
}

func (table *UserInfo) TableName() string {
	return "users"
}
