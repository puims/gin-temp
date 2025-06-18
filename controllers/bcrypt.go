package controllers

import "golang.org/x/crypto/bcrypt"

func HashPassword(password string) (string, error) {
	hashPwd, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashPwd), err
}

func VerifyPassword(password, hashPassword string) (bool, error) {
	if err := bcrypt.CompareHashAndPassword([]byte(hashPassword), []byte(password)); err != nil {
		return false, err
	}
	return true, nil
}
