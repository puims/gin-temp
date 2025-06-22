package controller

type UserCreate struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password"`
	Email    string `json:"email" binding:"required"`
}

type UserLogin struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UserRole struct {
	Role string `json:"role" binding:"required"`
}

type UserInfo struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required"`
}

type UserPassword struct {
	Password    string `json:"password" binding:"required"`
	NewPassword string `json:"newpassword" binding:"required"`
}
