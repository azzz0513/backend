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
type ResponseData struct {
	Code ResCode     `json:"code"`
	Msg  interface{} `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

// ResponseError 返回错误
func ResponseError(c *gin.Context, code ResCode) {
	c.JSON(http.StatusOK, &ResponseData{
		Code: code,
		Msg:  code.Msg(),
		Data: nil,
	})
}

// ResponseErrorWithMsg 返回带有指定信息的错误
func ResponseErrorWithMsg(c *gin.Context, code ResCode, msg interface{}) {
	c.JSON(http.StatusOK, &ResponseData{
		Code: code,
		Msg:  msg,
		Data: nil,
	})
}

// ResponseSuccess 成功响应
func ResponseSuccess(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, &ResponseData{
		Code: CodeSuccess,
		Msg:  CodeSuccess.Msg(),
		Data: data,
	})
}
