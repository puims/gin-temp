package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username string `gorm:"type:varchar(20);not null;index:idx_username,unique,where:deleted_at IS NULL"`
	Password string `gorm:"type:varchar(255)"`
	Email    string `gorm:"type:varchar(100);not null;index:idx_email,unique,where:deleted_at IS NULL"`
	Roles    []Role `gorm:"many2many:user_roles;"`
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.Password != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}

		u.Password = string(hashed)
		return nil
	}
	return
}

type Role struct {
	gorm.Model
	Name        string `gorm:"type:varchar(20);unique;not null"`
	Description string `gorm:"type:varchar(100)"`
	Users       []User `gorm:"many2many:user_roles;"`
}

type UserRole struct {
	UserID    uint      `gorm:"primaryKey"`
	RoleID    uint      `gorm:"primaryKey"`
	CreatedAt time.Time `gorm:"autoCreateTime"`

	// User User `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	// Role Role `gorm:"foreignKey:RoleID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}
