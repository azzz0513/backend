package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

/*
{
	"code": 10001,	// 程序中的错误码
	"msg": xx,		// 提示信息
	"data": {},		// 存放数据
}
*/

// ResponseData 返回数据
// swagger:response responseData
type ResponseData struct {
	// 业务状态码:
	// - 1000: 业务处理成功
	// - 1001: 请求参数错误
	// - 1002: 用户名已存在
	// - 1003: 用户名不存在
	// - 1004: 用户名或密码错误
	// - 1005: 服务繁忙
	// - 1006: 需要登录
	// - 1007: 无效的Token
	// - 1008: 需要正确的验证码
	// - 1009: 活动超时或未开始
	Code ResCode `json:"code"`

	// 消息描述:
	// - 按照状态码内容返回具体信息
	Msg interface{} `json:"msg"`

	// 返回数据（成功时有效）
	Data interface{} `json:"data,omitempty"`
}

// ResponseError 返回错误
// swagger:response errorResponse
func ResponseError(c *gin.Context, code ResCode) {
	c.JSON(http.StatusOK, &ResponseData{
		Code: code,
		Msg:  code.Msg(),
		Data: nil,
	})
}

// ResponseErrorWithMsg 返回带有指定信息的错误
// swagger:response customErrorResponse
func ResponseErrorWithMsg(c *gin.Context, code ResCode, msg interface{}) {
	c.JSON(http.StatusOK, &ResponseData{
		Code: code,
		Msg:  msg,
		Data: nil,
	})
}

// ResponseSuccess 成功响应
// swagger:response successResponse
func ResponseSuccess(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, &ResponseData{
		Code: CodeSuccess,
		Msg:  CodeSuccess.Msg(),
		Data: data,
	})
}
