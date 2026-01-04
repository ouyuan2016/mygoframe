// Package utils 提供通用的工具函数，包括统一API响应格式
package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一API响应结构体
type Response struct {
	Code    int         `json:"code"`    // 业务状态码，0表示成功
	Message string      `json:"message"` // 响应消息
	Data    interface{} `json:"data"`    // 响应数据
}

// Success 成功响应（极简版）
// 支持多种调用方式：
// 1. Success(c) - 只返回默认消息
// 2. Success(c, data) - 返回默认消息和数据
// 3. Success(c, "自定义消息") - 返回自定义消息
// 4. Success(c, "自定义消息", data) - 返回自定义消息和数据
func Success(c *gin.Context, args ...interface{}) {
	msg := "Success"
	var data interface{}

	switch len(args) {
	case 0:
		// Success(c)
		c.JSON(http.StatusOK, Response{Code: 0, Message: msg, Data: nil})
	case 1:
		// Success(c, data) 或 Success(c, "message")
		if str, ok := args[0].(string); ok {
			msg = str
			data = nil
		} else {
			data = args[0]
		}
		c.JSON(http.StatusOK, Response{Code: 0, Message: msg, Data: data})
	default:
		// Success(c, "message", data)
		if str, ok := args[0].(string); ok {
			msg = str
			data = args[1]
		} else {
			data = args[0]
		}
		c.JSON(http.StatusOK, Response{Code: 0, Message: msg, Data: data})
	}
}

// SuccessWithStatus 带HTTP状态码的成功响应（扩展版）
// 支持：
// 1. SuccessWithStatus(c, http.StatusCreated) - 只返回状态码和默认消息
// 2. SuccessWithStatus(c, http.StatusCreated, data) - 返回状态码、默认消息和数据
// 3. SuccessWithStatus(c, http.StatusCreated, "消息", data) - 返回状态码、自定义消息和数据
func SuccessWithStatus(c *gin.Context, status int, args ...interface{}) {
	msg := "Success"
	var data interface{}

	switch len(args) {
	case 0:
		c.JSON(status, Response{Code: 0, Message: msg, Data: nil})
	case 1:
		if str, ok := args[0].(string); ok {
			msg = str
			data = nil
		} else {
			data = args[0]
		}
		c.JSON(status, Response{Code: 0, Message: msg, Data: data})
	default:
		if str, ok := args[0].(string); ok {
			msg = str
			data = args[1]
		} else {
			data = args[0]
		}
		c.JSON(status, Response{Code: 0, Message: msg, Data: data})
	}
}

// Error 错误响应（极简版）
// 支持：
// 1. Error(c, "错误消息") - 默认500错误
// 2. Error(c, "错误消息", 400) - 自定义HTTP状态码（业务码自动同步）
// 3. Error(c, "错误消息", 400, 10001) - 自定义HTTP状态码和业务码
func Error(c *gin.Context, msg string, codes ...int) {
	status := http.StatusInternalServerError
	code := 500

	if len(codes) > 0 {
		status = codes[0]
		code = codes[0]
	}
	if len(codes) > 1 {
		code = codes[1]
	}

	c.JSON(status, Response{Code: code, Message: msg, Data: nil})
}

// BadRequest 400错误响应
func BadRequest(c *gin.Context, msg string) {
	c.JSON(http.StatusBadRequest, Response{Code: 400, Message: msg, Data: nil})
}

// Unauthorized 401错误响应
func Unauthorized(c *gin.Context, msg string) {
	c.JSON(http.StatusUnauthorized, Response{Code: 401, Message: msg, Data: nil})
}

// Forbidden 403错误响应
func Forbidden(c *gin.Context, msg string) {
	c.JSON(http.StatusForbidden, Response{Code: 403, Message: msg, Data: nil})
}

// NotFound 404错误响应
func NotFound(c *gin.Context, msg string) {
	c.JSON(http.StatusNotFound, Response{Code: 404, Message: msg, Data: nil})
}

// ServerError 500错误响应
func ServerError(c *gin.Context, msg string) {
	c.JSON(http.StatusInternalServerError, Response{Code: 500, Message: msg, Data: nil})
}

// ValidationError 验证错误响应（400状态码，422业务码）
func ValidationError(c *gin.Context, msg string) {
	c.JSON(http.StatusBadRequest, Response{Code: 422, Message: msg, Data: nil})
}
