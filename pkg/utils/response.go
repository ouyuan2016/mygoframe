package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

func SuccessWithStatus(c *gin.Context, status int, data interface{}) {
	c.JSON(status, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

func Error(c *gin.Context, message string, status int) {
	c.JSON(status, Response{
		Code:    status,
		Message: message,
		Data:    nil,
	})
}

func BadRequest(c *gin.Context, message string) {
	Error(c, message, http.StatusBadRequest)
}

func Unauthorized(c *gin.Context, message string) {
	Error(c, message, http.StatusUnauthorized)
}

func Forbidden(c *gin.Context, message string) {
	Error(c, message, http.StatusForbidden)
}

func NotFound(c *gin.Context, message string) {
	Error(c, message, http.StatusNotFound)
}

func ServerError(c *gin.Context, message string) {
	Error(c, message, http.StatusInternalServerError)
}
