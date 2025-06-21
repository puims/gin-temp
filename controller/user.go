package controller

import (
	"errors"
	"gin-temp/models"
	"gin-temp/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AdminController struct {
	DB *models.MysqlDB
}

// GetAllUsers 获取所有用户(分页+权限控制)
func (ac *AdminController) GetAllUsers(ctx *gin.Context) {
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
	query := ac.DB.Select("id", "username", "email", "created_at", "updated_at").
		Preload("Roles", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "name")
		}).
		Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize)

	if err := query.Find(&users).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	// 3. 获取总数
	var total int64
	if err := ac.DB.Model(&models.User{}).Count(&total).Error; err != nil {
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

// GetUserById 根据ID获取用户详情
func (ac *AdminController) GetUserById(ctx *gin.Context) {
	// 1. 参数验证
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil || id <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// 2. 查询用户(优化查询字段)
	var user models.User
	err = ac.DB.Select("id", "username", "email", "created_at", "updated_at").
		Preload("Roles", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "name")
		}).
		First(&user, "id = ?", id).Error

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

// UpdateUser 更新用户信息
func (ac *AdminController) UpdateUser(ctx *gin.Context) {
	// 1. 参数绑定与验证
	var userIn UserProfile
	if err := ctx.ShouldBindJSON(&userIn); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// 2. 获取认证用户
	claims, exists := ctx.Get("claims")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// 3. 检查用户名/邮箱是否已存在(排除当前用户)
	var existingUser models.User
	err := ac.DB.Where("(username = ? OR email = ?) AND id != ?",
		userIn.Username, userIn.Email, claims.(*utils.Claims).ID).
		First(&existingUser).Error

	if err == nil {
		ctx.JSON(http.StatusConflict, gin.H{"error": "Username or email already exists"})
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// 4. 获取当前用户
	var user models.User
	if err := ac.DB.First(&user, "id = ?", claims.(*utils.Claims).ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
	}

	// 5. 更新用户信息(使用事务)
	err = ac.DB.Transaction(func(tx *gorm.DB) error {
		// 更新基本信息
		user.Username = userIn.Username
		user.Email = userIn.Email
		user.Password = userIn.Password

		if err := tx.Save(&user).Error; err != nil {
			return err
		}

		// 处理管理员权限升级
		if userIn.AdminKey == "happydays" {
			var adminRole models.Role
			if err := tx.Select("id").First(&adminRole, "name = ?", "admin").Error; err != nil {
				return err
			}

			// 检查是否已有admin角色
			var existingRoles []models.Role
			if err := tx.Model(&user).Association("Roles").Find(&existingRoles); err != nil {
				return err
			}

			hasAdminRole := false
			for _, role := range existingRoles {
				if role.Name == "admin" {
					hasAdminRole = true
					break
				}
			}

			if !hasAdminRole {
				if err := tx.Model(&user).Association("Roles").Append(&models.Role{Model: gorm.Model{ID: adminRole.ID}}); err != nil {
					return err
				}
			}
		}

		return nil
	})

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	// 6. 返回更新后的用户信息(重新加载)
	if err := ac.DB.Preload("Roles").First(&user, "id = ?", user.ID).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch updated user"})
		return
	}

	ctx.JSON(http.StatusOK, user)
}

// DeleteUser 删除用户
func (ac *AdminController) DeleteUser(ctx *gin.Context) {
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
	if err := ac.DB.Preload("Roles").First(&user, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
	}

	// 5. 删除用户(使用事务)
	err = ac.DB.Transaction(func(tx *gorm.DB) error {
		// 删除关联角色
		if err := tx.Model(&user).Association("Roles").Clear(); err != nil {
			return err
		}

		// 删除用户
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
