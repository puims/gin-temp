package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username string `gorm:"type:varchar(20);not null;index:idx_username,unique,where:deleted_at IS NULL"`
	Password string `gorm:"type:varchar(255);not null"`
	Email    string `gorm:"type:varchar(100);not null;index:idx_email,unique,where:deleted_at IS NULL"`
	Roles    []Role `gorm:"many2many:user_roles;joinForeignKey:user_id;joinReferences:role_id"`
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

	if len(u.Roles) == 0 {
		var defaultRole Role
		if err := tx.Where("name = ?", "user").First(&defaultRole).Error; err != nil {
			defaultRole = Role{Name: "user"}
			if err := tx.Create(&defaultRole).Error; err != nil {
				return err
			}
		}
		u.Roles = []Role{defaultRole}
	}

	return
}

type Role struct {
	gorm.Model
	Name        string `gorm:"type:varchar(20);unique;not null"`
	Description string `gorm:"type:varchar(100)"`
	Users       []User `gorm:"many2many:user_roles;joinForeignKey:role_id;joinReferences:user_id"`
}

type UserRole struct {
	UserID    uint `gorm:"primaryKey"`
	RoleID    uint `gorm:"primaryKey"`
	CreatedAt time.Time
}
