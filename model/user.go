package model

import (
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username string `gorm:"type:varchar(20);uniqueIndex"`
	Password string `gorm:"type:varchar(255)"`
	Email    string `gorm:"type:varchar(100);uniqueIndex"`
	Role     string `gorm:"type:varchar(20);default:'user'"`
}

func (u *User) BeforeSave(tx *gorm.DB) (err error) {
	if u.Password == "" {
		return
	}

	if len(u.Password) == 60 && (strings.HasPrefix(u.Password, "$2a$") ||
		strings.HasPrefix(u.Password, "$2b$") ||
		strings.HasPrefix(u.Password, "$2y$")) {
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	u.Password = string(hashed)
	return nil
}

func (u *User) AfterDelete(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		u.Username = fmt.Sprintf("DEL_%s", u.Username)
		u.Email = fmt.Sprintf("DEL_%s", u.Email)
		return tx.Save(&u).Error
	})
}
