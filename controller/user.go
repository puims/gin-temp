package controller

import (
	"errors"
	"gin-temp/config"
	"gin-temp/models"
	"gin-temp/utils"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	roleList = getRoleList()
)

type UserController struct {
	DB *utils.MysqlDB
}

func (uc *UserController) setupSelect() *gorm.DB {
	return uc.DB.Select("id", "username", "email", "role",
		"created_at", "updated_at")
}

func (uc *UserController) GetAllUsers(ctx *gin.Context) {
	// 1. 分页参数处理
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// 2. 查询用户列表(优化查询字段)
	var users []models.User
	query := uc.setupSelect().Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize)

	if err := query.Find(&users).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	// 3. 获取总数
	var total int64
	if err := uc.DB.Model(&models.User{}).Count(&total).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count users"})
		return
	}

	// 4. 返回响应(隐藏敏感信息)
	ctx.JSON(http.StatusOK, gin.H{
		"data":  users,
		"total": total,
		"page":  page,
		"size":  pageSize,
	})
}

func (uc *UserController) SearchUserByKeyward(ctx *gin.Context) {
	keyword := ctx.Query("keyword")
	keyword = "%" + keyword + "%"

	var users []models.User
	if err := uc.setupSelect().Where("username LIKE ? OR email LIKE ?", keyword, keyword).
		Find(&users).Error; err != nil {
		ctx.JSON(500, gin.H{"error": "failed to search user"})
		return
	}

	ctx.JSON(200, users)
}

func (uc *UserController) GetUserById(ctx *gin.Context) {
	// 1. 参数验证
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil || id <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// 2. 查询用户(优化查询字段)
	var user models.User
	err = uc.setupSelect().First(&user, "id = ?", id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
	}

	// 3. 返回响应
	ctx.JSON(http.StatusOK, user)
}

func (uc *UserController) CreateUser(ctx *gin.Context) {
	userIn := UserCreate{}
	if err := ctx.ShouldBindJSON(&userIn); err != nil {
		ctx.JSON(403, gin.H{"error": err.Error()})
		return
	}

	queryUser := models.User{}
	if err := uc.DB.First(&queryUser, "username = ? OR email = ?", userIn.Username, userIn.Email).
		Error; err == nil {
		ctx.JSON(http.StatusConflict, gin.H{"error": "Username or email already exists"})
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	user := models.User{
		Username: userIn.Username,
		Password: userIn.Password,
		Email:    userIn.Email,
	}

	err := uc.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&user).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
		},
	})
}

func (uc *UserController) LoginCheck(ctx *gin.Context) {
	userIn := UserLogin{}
	if err := ctx.ShouldBindJSON(&userIn); err != nil {
		ctx.JSON(403, gin.H{"error": err.Error()})
		return
	}

	user := models.User{}
	if err := uc.DB.Select("id", "username", "password", "email", "role").
		First(&user, "username = ? OR email = ?", userIn.Account, userIn.Account).
		Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(403, gin.H{"error": "user not found"})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(userIn.Password)); err != nil {
		ctx.JSON(http.StatusConflict,
			gin.H{"error": "Invalid credentials"})
		return
	}

	token, err := utils.GenerateToken(&user)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError,
			gin.H{"error": "Failed to generate token"})
		return
	}

	ctx.JSON(200, gin.H{
		"token": token,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"role":     user.Role,
		},
	})
}

func (uc *UserController) ChangeUserRole(ctx *gin.Context) {
	// 1. 参数绑定与验证
	var userIn UserRole
	if err := ctx.ShouldBindJSON(&userIn); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// 2. 获取当前用户(用于权限验证)
	id := ctx.Param("id")
	claims, exists := ctx.Get("claims")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	if id == strconv.Itoa(int(claims.(*utils.Claims).ID)) {
		ctx.JSON(403, gin.H{"error": "could not change role by yourself"})
		return
	}

	user := models.User{}
	if err := uc.DB.First(&user, "id = ?", id).Error; err != nil {
		ctx.JSON(403, gin.H{"error": "failed to find user"})
		return
	}

	if !hasMorePermission(claims.(*utils.Claims).Role, user.Role) ||
		!hasEquelPermission(claims.(*utils.Claims).Role, userIn.Role) ||
		!hasRole(userIn.Role) {
		ctx.JSON(403, gin.H{"error": "no permissions"})
		return
	}

	// 5. 更新用户信息(使用事务)
	if err := uc.DB.Transaction(func(tx *gorm.DB) error {
		user.Role = userIn.Role
		return tx.Save(&user).Error
	}); err != nil {
		ctx.JSON(500, gin.H{"error": "Failed to update user"})
		return
	}

	ctx.JSON(200, gin.H{
		"success": true,
		"msg":     "user role has changed",
	})
}

func (uc *UserController) ChangePassword(ctx *gin.Context) {
	// 1. 参数绑定与验证
	var userIn UserPassword
	if err := ctx.ShouldBindJSON(&userIn); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	user, err := uc.getCurrentUser(ctx)
	if err != nil {
		return // 错误已在getCurrentUser中处理
	}

	// 5. 更新用户信息(使用事务)
	if err := uc.DB.Transaction(func(tx *gorm.DB) error {
		user.Password = userIn.NewPassword
		return tx.Save(&user).Error
	}); err != nil {
		ctx.JSON(500, gin.H{"error": "failed to update password"})
		return
	}

	ctx.JSON(200, gin.H{
		"success": true,
		"msg":     "password has changed",
	})
}

func (uc *UserController) ChangeUserinfo(ctx *gin.Context) {
	// 1. 参数绑定与验证
	var userIn UserInfo
	if err := ctx.ShouldBindJSON(&userIn); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// 2. 获取认证用户
	user, err := uc.getCurrentUser(ctx)
	if err != nil {
		return
	}

	// 5. 更新用户信息(使用事务)
	err = uc.DB.Transaction(func(tx *gorm.DB) error {
		user.Username = userIn.Username
		user.Email = userIn.Email
		if err := tx.Save(&user).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	ctx.JSON(200, gin.H{
		"success": true,
		"msg":     "userinfo has changed",
	})
}

func (uc *UserController) DeleteUser(ctx *gin.Context) {
	// 1. 参数验证
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil || id <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// 2. 获取当前用户(用于权限验证)
	claims, exists := ctx.Get("claims")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// 3. 检查是否在删除自己
	if id == int(claims.(*utils.Claims).ID) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Cannot delete yourself"})
		return
	}

	// 4. 查询要删除的用户
	var user models.User
	if err := uc.DB.First(&user, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
	}

	if !hasMorePermission(claims.(*utils.Claims).Role, user.Role) {
		ctx.JSON(403, gin.H{"error": "no permissions"})
		return
	}

	// 5. 删除用户(使用事务)
	err = uc.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&user).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	// 6. 返回成功响应
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "User deleted successfully",
	})
}

func (uc *UserController) getCurrentUser(ctx *gin.Context) (*models.User, error) {
	claims, exists := ctx.Get("claims")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return nil, errors.New("Unauthorization")
	}

	var user models.User
	if err := uc.DB.First(&user, "id = ?", claims.(*utils.Claims).ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		} else {
			ctx.JSON(500, gin.H{"error": "Failed to retrieve user: " + err.Error()})
		}
		return nil, err
	}
	return &user, nil
}

func hasMorePermission(current, target string) bool {
	iCurrent := indexOf(roleList, current)
	iTarget := indexOf(roleList, target)

	if iCurrent < 0 || iTarget < 0 {
		return false
	}
	return iTarget < iCurrent
}

func hasEquelPermission(current, target string) bool {
	iCurrent := indexOf(roleList, current)
	iTarget := indexOf(roleList, target)

	if iCurrent < 0 || iTarget < 0 {
		return false
	}
	return iTarget <= iCurrent
}

func hasRole(target string) bool {
	for _, rl := range roleList {
		if rl == target {
			return true
		}
	}
	return false
}

func getRoleList() []string {
	rolesStr := config.Viper.GetString("app.roles")
	return strings.Split(rolesStr, ",")
}

func indexOf(slice []string, target string) int {
	for i, v := range slice {
		if v == target {
			return i
		}
	}
	return -1
}
