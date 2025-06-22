package controller

import (
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

type UserCreate struct {
	Username string `json:"username" binding:"required,min=3,max=20"`
	Password string `json:"password" binding:"required,min=6,password"`
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
	Username string `json:"username" binding:"required,min=3,max=20"`
	Email    string `json:"email" binding:"required,email"`
}

type UserPassword struct {
	Password    string `json:"password" binding:"required,min=6"`
	NewPassword string `json:"newpassword" binding:"required,min=6,password"`
}

func PasswordValidator(fl validator.FieldLevel) bool {
	lowercaseRegex := regexp.MustCompile(`[a-z]`)
	numberRegex := regexp.MustCompile(`[0-9]`)

	password := fl.Field().String()
	lowerPassword := strings.ToLower(password)

	hasLetter := lowercaseRegex.MatchString(lowerPassword)
	hasNum := numberRegex.MatchString(password)

	return hasLetter && hasNum
}
