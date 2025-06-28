package controller

import (
	"gin-temp/util"
	"regexp"
	"strings"
	"unicode"

	"github.com/go-playground/validator/v10"
)

var (
	roleList       = getRoleList()
	lowercaseRegex = regexp.MustCompile(`[a-z]`)
	numberRegex    = regexp.MustCompile(`[0-9]`)
)

type UserController struct {
	DB *util.MysqlDB
}

type UserCreate struct {
	Username string `json:"username" binding:"required,username"`
	Password string `json:"password" binding:"required,password"`
	Email    string `json:"email" binding:"required,email"`
}

type UserLogin struct {
	Account  string `json:"account" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UserRole struct {
	Role string `json:"role" binding:"required"`
}

type UserInfo struct {
	Username string `json:"username" binding:"required,username"`
	Email    string `json:"email" binding:"required,email"`
}

type UserPassword struct {
	Password    string `json:"password" binding:"required"`
	NewPassword string `json:"newpassword" binding:"required,password"`
}

func UsernameValidator(fl validator.FieldLevel) bool {
	username := fl.Field().String()
	if len(username) < 3 || len(username) > 20 {
		return false
	}

	hasLetter := false
	for _, char := range username {
		if unicode.IsLetter(char) {
			hasLetter = true
		} else if !unicode.IsNumber(char) {
			return false
		}
	}

	return hasLetter
}

func PasswordValidator(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	if len(password) < 6 {
		return false
	}

	lowerPassword := strings.ToLower(password)

	hasLetter := lowercaseRegex.MatchString(lowerPassword)
	hasNum := numberRegex.MatchString(password)

	return hasLetter && hasNum
}
