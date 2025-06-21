package controller

import (
	"gin-temp/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

type RoleController struct {
	DB *models.MysqlDB
}

func NewRoleController(db *models.MysqlDB) *RoleController {
	return &RoleController{DB: db}
}

// AssignRole 为用户分配角色
func (rc *RoleController) AssignRole(c *gin.Context) {
	var input struct {
		UserID uint   `json:"user_id" binding:"required"`
		Role   string `json:"role" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 查找用户和角色
	var user models.User
	var role models.Role

	if err := rc.DB.First(&user, input.UserID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if err := rc.DB.Where("name = ?", input.Role).First(&role).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
		return
	}

	// 检查是否已分配该角色
	var count int64
	rc.DB.Model(&models.UserRole{}).Where("user_id = ? AND role_id = ?", user.ID, role.ID).Count(&count)
	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "User already has this role"})
		return
	}

	// 分配角色
	if err := rc.DB.Model(&user).Association("Roles").Append(&role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Role assigned successfully",
		"user": gin.H{
			"id":    user.ID,
			"name":  user.Username,
			"roles": getRoleNames(user.Roles),
		},
	})
}

func getRoleNames(roles []models.Role) []string {
	var names []string
	for _, role := range roles {
		names = append(names, role.Name)
	}
	return names
}
