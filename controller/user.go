package controller

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"web_app/dao/mysql"
	"web_app/logic"
	"web_app/models"
)

// SignUpHandler 处理注册请求
// @Tags 用户管理
// @Summary 用户注册
// @Description 接收前端数据注册一个新用户
// @Param request body models.ParamSignUp  true  "注册凭证"
// @Router /api/v1/signup [post]
// @Accept json
// @Produce json
// @Success 200 {object} ResponseData "成功响应示例：{"code":1000,"msg":"业务处理成功","data":null}"
func SignUpHandler(c *gin.Context) {
	// 1.获取参数和参数校验
	p := new(models.ParamSignUp)
	if err := c.ShouldBindJSON(p); err != nil {
		// 请求参数有误，直接返回响应
		zap.L().Error("SignUp with invalid param", zap.Error(err))
		// 判断errs是不是validator.ValidationErrors类型
		errs, ok := err.(validator.ValidationErrors)
		if !ok {
			ResponseError(c, CodeInvalidParam)
			return
		}
		ResponseErrorWithMsg(c, CodeInvalidParam, removeTopStruct(errs.Translate(trans)))
		return
	}
	// 手动对请求参数进行详细的业务规则校验
	//if len(p.Username) == 0 || len(p.Password) == 0 || len(p.RePassword) == 0 || p.RePassword != p.Password {
	//	// 请求参数有误，直接返回响应
	//	zap.L().Error("SignUp with invalid param")
	//	c.JSON(http.StatusOK, gin.H{
	//		"msg": "请求参数有误",
	//	})
	//	return
	//}
	fmt.Println(p)
	// 2.业务处理
	if err := logic.SignUp(p); err != nil {
		zap.L().Error("logic.SignUp failed", zap.Error(err))
		if errors.Is(err, mysql.ErrorUserExist) {
			ResponseError(c, CodeUserExist)
		}
		ResponseError(c, CodeServerBusy)
		return
	}
	// 3.返回响应
	ResponseSuccess(c, nil)
}

// LoginHandler 处理登录请求
// @Tags 用户管理
// @Summary 用户登录
// @Description 接收前端数据登录用户账户
// @Param request body models.ParamLogin  true  "登录凭证"
// @Router /api/v1/login [post]
// @Accept json
// @Produce json
// @Success 200 {object} ResponseData "成功响应示例：{"code":1000,"msg":"业务处理成功","data":models.LoginResponse}"
func LoginHandler(c *gin.Context) {
	// 1.获取请求参数和参数校验
	p := new(models.ParamLogin)
	if err := c.ShouldBindJSON(p); err != nil {
		// 请求参数有误，直接返回响应
		zap.L().Error("Login with invalid param", zap.Error(err))
		// 判断errs是不是validator.ValidationErrors类型
		errs, ok := err.(validator.ValidationErrors)
		if !ok {
			ResponseError(c, CodeInvalidParam)
			return
		}
		ResponseErrorWithMsg(c, CodeInvalidParam, removeTopStruct(errs.Translate(trans)))
		return
	}
	// 2.业务处理
	user, err := logic.Login(p)
	if err != nil {
		zap.L().Error("logic.Login failed", zap.Error(err))
		if errors.Is(err, mysql.ErrorUserNotExist) {
			ResponseError(c, CodeUserNotExist)
			return
		}
		ResponseError(c, CodeInvalidPassword)
		return
	}

	// 3.返回响应
	ResponseSuccess(c, &models.LoginResponse{
		UserID:   fmt.Sprintf("%d", user.UserID), // id值1<<53-1，int64类型的最大值为1<<63-1
		UserName: user.Username,
		Token:    user.Token,
	})
}

// UpdateUserHandler 处理修改用户数据
// @Tags 用户管理
// @Summary 用户数据修改
// @Description 接收前端数据修改用户数据
// @Param request body models.UpdateUser  true  "用户数据修改参数"
// @Router /api/v1/change_data [post]
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} ResponseData "成功响应示例：{"code":1000,"msg":"业务处理成功","data":null}"
func UpdateUserHandler(c *gin.Context) {
	// 获取请求参数
	u := new(models.UpdateUser)
	if err := c.ShouldBindJSON(&u); err != nil {
		zap.L().Debug("c.ShouldBindJSON(l) err", zap.Any("err", err))
		zap.L().Error("UpdateUser with invalid param", zap.Error(err))
		ResponseError(c, CodeInvalidParam)
		return
	}
	// 获取用户id
	userID, err := getCurrentUserID(c)
	if err != nil {
		ResponseError(c, CodeNeedLogin)
		return
	}
	u.UserID = userID
	// 修改成员数据的具体逻辑
	if err := logic.UpdateUser(u); err != nil {
		zap.L().Error("logic.UpdateUser failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	// 返回响应
	ResponseSuccess(c, nil)
}
