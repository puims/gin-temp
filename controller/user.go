package controller

import (
	"errors"
	"gin-temp/model"
	"gin-temp/util"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

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
	var users []model.User
	query := uc.setupSelect().Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize)

	if err := query.Find(&users).Error; err != nil {
		model.ErrorResponse(ctx, http.StatusInternalServerError, errors.New("database error: "+err.Error()))
		return
	}

	// 3. 获取总数
	var total int64
	if err := uc.DB.Model(&model.User{}).Count(&total).Error; err != nil {
		model.ErrorResponse(ctx, http.StatusInternalServerError, errors.New("database error: "+err.Error()))
		return
	}

	// 4. 返回响应(隐藏敏感信息)
	model.SuccessPaginateResponse(ctx, 200, users, page, int(total), pageSize)
}

func (uc *UserController) SearchUserByKeyward(ctx *gin.Context) {
	keyword := ctx.Query("keyword")
	keyword = "%" + keyword + "%"

	var users []model.User
	if err := uc.setupSelect().Where("username LIKE ? OR email LIKE ?", keyword, keyword).
		Find(&users).Error; err != nil {
		model.ErrorResponse(ctx, http.StatusInternalServerError, errors.New("database error: "+err.Error()))
		return
	}

	model.SuccessResponse(ctx, 200, users)
}

func (uc *UserController) GetUserById(ctx *gin.Context) {
	// 1. 参数验证
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil || id <= 0 {
		model.ErrorResponse(ctx, http.StatusBadRequest, errors.New("invalid user ID"))
		return
	}

	// 2. 查询用户(优化查询字段)
	var user model.User
	err = uc.setupSelect().First(&user, "id = ?", id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			model.ErrorResponse(ctx, http.StatusNotFound, errors.New("user not found"))
		} else {
			model.ErrorResponse(ctx, http.StatusInternalServerError, errors.New("database error: "+err.Error()))
			return
		}
		return
	}

	// 3. 返回响应
	model.SuccessResponse(ctx, 200, user)
}

func (uc *UserController) CreateUser(ctx *gin.Context) {
	userIn := UserCreate{}
	if err := ctx.ShouldBindJSON(&userIn); err != nil {
		model.ErrorResponse(ctx, http.StatusBadRequest, errors.New("invalid request payload"))
		return
	}

	queryUser := model.User{}
	if err := uc.DB.First(&queryUser, "username = ? OR email = ?", userIn.Username, userIn.Email).
		Error; err == nil {
		model.ErrorResponse(ctx, http.StatusConflict, errors.New("username or email already exists"))
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		model.ErrorResponse(ctx, http.StatusInternalServerError, errors.New("database error: "+err.Error()))
		return
	}

	user := model.User{
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
		model.ErrorResponse(ctx, http.StatusInternalServerError, errors.New("database error: "+err.Error()))
		return
	}

	userResponse := model.UserResponse{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Username:  user.Username,
		Email:     user.Email,
		Role:      user.Role,
	}

	model.SuccessResponse(ctx, 201, gin.H{
		"user": userResponse,
	})
}

func (uc *UserController) Login(ctx *gin.Context) {
	userIn := UserLogin{}
	if err := ctx.ShouldBindJSON(&userIn); err != nil {
		model.ErrorResponse(ctx, http.StatusBadRequest, errors.New("invalid request payload"))
		return
	}

	user := model.User{}
	if err := uc.DB.Select("id", "username", "password", "email", "role").
		First(&user, "username = ? OR email = ?", userIn.Account, userIn.Account).
		Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			model.ErrorResponse(ctx, http.StatusNotFound, errors.New("user not found"))
		} else {
			model.ErrorResponse(ctx, http.StatusInternalServerError, errors.New("database error: "+err.Error()))
		}
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(userIn.Password)); err != nil {
		model.ErrorAbortResponse(ctx, http.StatusConflict, errors.New("invalid credentials"))
		return
	}

	expires := util.Viper.GetInt("app.expires")
	accessToken, err := util.GenerateToken(&user, expires, util.JwtKey)
	if err != nil {
		model.ErrorResponse(ctx, http.StatusInternalServerError, errors.New("failed to generate token"))
		return
	}

	refreshToken, err := util.GenerateToken(&user, 168, util.JwtKeyRefresh)
	if err != nil {
		model.ErrorResponse(ctx, http.StatusInternalServerError, errors.New("failed to generate refresh token"))
		return
	}

	userResponse := model.UserResponse{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Username:  user.Username,
		Email:     user.Email,
		Role:      user.Role,
	}

	model.SuccessResponse(ctx, 200, gin.H{
		"access-token":  accessToken,
		"refresh-token": refreshToken,
		"user":          userResponse,
		"expires":       time.Now().Add(time.Duration(expires) * time.Hour),
	})
}

func (uc *UserController) Logout(ctx *gin.Context) {
	token := ctx.GetHeader("Authorization")

	claims, exists := ctx.Get("claims")
	if !exists {
		model.ErrorResponse(ctx, 401, errors.New("unauthorization"))
		return
	}

	if err := util.AddToBlackList(token,
		time.Unix(claims.(*util.Claims).ExpiresAt.Unix(), 0),
		util.Redis); err != nil {
		model.ErrorResponse(ctx, 500, errors.New("failed to add token to blacklist: "+err.Error()))
	}

	model.SuccessResponse(ctx, 200, gin.H{"message": "user has logout"})
}

func (uc *UserController) TokenRefresh(ctx *gin.Context) {
	refreshToken := ctx.PostForm("refresh-token")
	if refreshToken == "" {
		model.ErrorResponse(ctx, 403, errors.New("failed to get refresh-token"))
		return
	}

	claims, err := util.ParseToken(refreshToken, util.JwtKeyRefresh)
	if err != nil {
		model.ErrorResponse(ctx, 403, errors.New("failed to parse refresh-token: "+err.Error()))
		return
	}

	user := model.User{}
	if err := uc.DB.Select("id", "username", "role").First(&user, "id = ?", claims.ID).
		Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			model.ErrorResponse(ctx, 403, errors.New("user not found"))
		} else {
			model.ErrorResponse(ctx, 500, errors.New("database error: "+err.Error()))
		}
		return
	}

	accessToken := ""
	expires := util.Viper.GetInt("app.expires")
	accessToken, err = util.GenerateToken(&user, expires, util.JwtKey)
	if err != nil {
		model.ErrorResponse(ctx, 500, errors.New("failed to generate access token: "+err.Error()))
		return
	}

	model.SuccessResponse(ctx, 200, gin.H{
		"code":          200,
		"state":         "success",
		"data":          user,
		"message":       "token refreshed successfully",
		"access-token":  accessToken,
		"refresh-token": refreshToken,
		"expires":       expires,
	})
}

func (uc *UserController) ChangeUserRole(ctx *gin.Context) {
	// 1. 参数绑定与验证
	var userIn UserRole
	if err := ctx.ShouldBindJSON(&userIn); err != nil {
		model.ErrorResponse(ctx, http.StatusBadRequest, errors.New("invalid request payload"))
		return
	}

	// 2. 获取当前用户(用于权限验证)
	id := ctx.Param("id")
	claims, exists := ctx.Get("claims")
	if !exists {
		model.ErrorResponse(ctx, 401, errors.New("unauthorization"))
		return
	}
	if id == strconv.Itoa(int(claims.(*util.Claims).ID)) {
		model.ErrorResponse(ctx, 403, errors.New("could not change role by yourself"))
		return
	}

	user := model.User{}
	if err := uc.DB.First(&user, "id = ?", id).Error; err != nil {
		model.ErrorResponse(ctx, http.StatusNotFound, errors.New("user not found"))
		return
	}

	if !hasPermission(claims.(*util.Claims).Role, user.Role) {
		model.ErrorResponse(ctx, 403, errors.New("no permissions"))
		return
	}

	if indexOf(roleList, userIn.Role) < 0 {
		model.ErrorResponse(ctx, 403, errors.New("entries incorrect"))
		return
	}

	if err := uc.DB.Transaction(func(tx *gorm.DB) error {
		user.Role = userIn.Role
		return tx.Save(&user).Error
	}); err != nil {
		model.ErrorResponse(ctx, 500, errors.New("failed to update user role: "+err.Error()))
		return
	}

	model.SuccessResponse(ctx, 200, gin.H{
		"code":    200,
		"state":   "success",
		"message": "user role has changed",
		"data":    user,
	})
}

func (uc *UserController) ChangePassword(ctx *gin.Context) {
	var userIn UserPassword
	if err := ctx.ShouldBindJSON(&userIn); err != nil {
		model.ErrorResponse(ctx, http.StatusBadRequest, errors.New("invalid request payload"))
		return
	}

	user, err := uc.getCurrentUser(ctx)
	if err != nil {
		return
	}

	if err := uc.DB.Transaction(func(tx *gorm.DB) error {
		user.Password = userIn.NewPassword
		return tx.Save(&user).Error
	}); err != nil {
		model.ErrorResponse(ctx, http.StatusInternalServerError, errors.New("failed to update password: "+err.Error()))
		return
	}

	model.SuccessResponse(ctx, 200, gin.H{
		"code":    200,
		"state":   "success",
		"message": "password has changed",
		"data":    user,
	})
}

func (uc *UserController) ChangeUserinfo(ctx *gin.Context) {
	var userIn UserInfo
	if err := ctx.ShouldBindJSON(&userIn); err != nil {
		model.ErrorResponse(ctx, http.StatusBadRequest, errors.New("invalid request payload"))
		return
	}

	user, err := uc.getCurrentUser(ctx)
	if err != nil {
		return
	}

	err = uc.DB.Transaction(func(tx *gorm.DB) error {
		user.Username = userIn.Username
		user.Email = userIn.Email
		if err := tx.Save(&user).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		model.ErrorResponse(ctx, http.StatusInternalServerError, errors.New("failed to update userinfo: "+err.Error()))
	}

	model.SuccessResponse(ctx, 200, gin.H{
		"code":    200,
		"state":   "success",
		"message": "userinfo has changed",
		"data":    user,
	})
}

func (uc *UserController) DeleteUser(ctx *gin.Context) {
	// 1. 参数验证
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil || id <= 0 {
		model.ErrorResponse(ctx, http.StatusBadRequest, errors.New("invalid user ID"))
		return
	}

	// 2. 获取当前用户(用于权限验证)
	claims, exists := ctx.Get("claims")
	if !exists {
		model.ErrorResponse(ctx, http.StatusUnauthorized, errors.New("unauthorization"))
		return
	}

	// 3. 检查是否在删除自己
	if id == int(claims.(*util.Claims).ID) {
		model.ErrorResponse(ctx, http.StatusForbidden, errors.New("could not delete yourself"))
		return
	}

	// 4. 查询要删除的用户
	var user model.User
	if err := uc.DB.First(&user, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			model.ErrorResponse(ctx, http.StatusNotFound, errors.New("user not found"))
		} else {
			model.ErrorResponse(ctx, http.StatusInternalServerError, errors.New("database error: "+err.Error()))
		}
		return
	}

	if !hasPermission(claims.(*util.Claims).Role, user.Role) {
		model.ErrorResponse(ctx, http.StatusForbidden, errors.New("no permissions to delete this user"))
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
		model.ErrorResponse(ctx, http.StatusInternalServerError, errors.New("failed to delete user: "+err.Error()))
		return
	}

	// 6. 返回成功响应
	model.SuccessResponse(ctx, 200, gin.H{
		"code":    200,
		"state":   "success",
		"message": "User deleted successfully",
		"data":    user,
	})
}

func (uc *UserController) setupSelect() *gorm.DB {
	return uc.DB.Select("id", "username", "email", "role",
		"created_at", "updated_at")
}

func (uc *UserController) getCurrentUser(ctx *gin.Context) (*model.User, error) {
	claims, exists := ctx.Get("claims")
	if !exists {
		err := errors.New("unauthorization")
		model.ErrorResponse(ctx, http.StatusUnauthorized, err)
		return nil, err
	}

	var user model.User
	if err := uc.DB.First(&user, "id = ?", claims.(*util.Claims).ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			model.ErrorResponse(ctx, http.StatusNotFound, errors.New("user not found"))
		} else {
			model.ErrorResponse(ctx, http.StatusInternalServerError, errors.New("database error: "+err.Error()))
		}
		return nil, err
	}
	return &user, nil
}

func hasPermission(current, target string) bool {
	iCurrent := indexOf(roleList, current)
	iTarget := indexOf(roleList, target)

	if iCurrent < 0 || iTarget < 0 {
		return false
	}
	return iTarget < iCurrent
}

func getRoleList() []string {
	rolesStr := util.Viper.GetString("app.roles")
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
