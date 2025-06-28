package model

import (
	"time"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code  int         `json:"code"`
	State string      `json:"state"`
	Data  interface{} `json:"data,omitempty"`
	Error string      `json:"error,omitempty"`
	Page  int         `json:"page,omitempty"`
	Total int         `json:"total,omitempty"`
	Size  int         `json:"size,omitempty"`
}

type UserResponse struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
}

func SuccessResponse(ctx *gin.Context, code int, data interface{}) {
	ctx.JSON(200, Response{
		Code:  code,
		State: "success",
		Data:  data,
	})
}

func SuccessPaginateResponse(ctx *gin.Context, code int, data interface{}, page, total, size int) {
	ctx.JSON(200, Response{
		Code:  code,
		State: "success",
		Data:  data,
		Page:  page,
		Total: total,
		Size:  size,
	})
}

func ErrorResponse(ctx *gin.Context, code int, err error) {
	ctx.JSON(200, Response{
		Code:  code,
		State: "error",
		Error: err.Error(),
	})
}

func ErrorAbortResponse(ctx *gin.Context, code int, err error) {
	ctx.AbortWithStatusJSON(200, Response{
		Code:  code,
		State: "error",
		Error: err.Error(),
	})
}
