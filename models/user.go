package models

import "github.com/jinzhu/gorm"

type User struct {
	gorm.Model
	Username string `gorm:"type:varchar(20);uniqueIndex:idx_username_deleted"`
	Password string
	Email    string `gorm:"type:varchar(100);uniqueIndex:idx_email_deleted"`
	Roles    []Role `gorm:"many2many:user_roles"`
}

type Role struct {
	gorm.Model
	Name  string `gorm:"type:varchar(20);unique;not null"`
	Users []User `gorm:"many2many:user_roles"`
}

func (User) Indexes(db *gorm.DB) error {
	err := db.Exec(`
        CREATE UNIQUE INDEX idx_username_deleted ON users(username, deleted_at) 
        WHERE deleted_at IS NULL
    `).Error
	if err != nil {
		return err
	}

	err = db.Exec(`
        CREATE UNIQUE INDEX idx_email_deleted ON users(email, deleted_at) 
        WHERE deleted_at IS NULL
    `).Error
	return err
}

type UserLogin struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UserRegister struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email" binding:"required"`
}

type UserProfile struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email" binding:"required"`
	AdminKey string `json:"adminkey"`
}
