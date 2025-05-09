package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"strconv"
	"web_app/logic"
	"web_app/models"
)

// CreateMemberListHandler 创建成员列表的处理函数
func CreateMemberListHandler(c *gin.Context) {
	// 1.获取参数及参数校验
	l := new(models.MemberList)
	if err := c.ShouldBindJSON(l); err != nil {
		zap.L().Debug("c.ShouldBindJSON(l) err", zap.Any("err", err))
		zap.L().Error("create member list failed with invalid param", zap.Error(err))
		ResponseError(c, CodeInvalidParam)
		return
	}
	// 从c中获取当前发送请求的用户的ID
	userID, err := getCurrentUserID(c)
	if err != nil {
		ResponseError(c, CodeNeedLogin)
		return
	}
	l.AuthorID = userID
	// 2.创建成员列表
	if err := logic.CreateMemberList(l); err != nil {
		zap.L().Error("logic.CreateMemberList failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	// 3.返回响应
	ResponseSuccess(c, nil)
}

// AddMemberHandler 往成员列表添加成员的处理函数
func AddMemberHandler(c *gin.Context) {
	// 参数校验
	m := new(models.UpdateMember)
	if err := c.ShouldBindJSON(&m); err != nil {
		zap.L().Debug("c.ShouldBindJSON(l) err", zap.Any("err", err))
		errs, ok := err.(validator.ValidationErrors) // 类型断言
		if !ok {
			ResponseError(c, CodeInvalidParam)
			return
		}
		errData := removeTopStruct(errs.Translate(trans)) // 翻译并去除掉错误提示中的结构体标识
		ResponseErrorWithMsg(c, CodeInvalidParam, errData)
		return
	}
	// 具体的添加用户的业务逻辑
	if err := logic.AddMember(m); err != nil {
		zap.L().Error("logic.AddMember failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	// 返回响应
	ResponseSuccess(c, nil)
}

// DeleteMemberHandler 将指定成员从成员名单中删除
func DeleteMemberHandler(c *gin.Context) {
	// 参数校验
	m := new(models.UpdateMember)
	if err := c.ShouldBindJSON(&m); err != nil {
		zap.L().Debug("c.ShouldBindJSON(l) err", zap.Any("err", err))
		errs, ok := err.(validator.ValidationErrors) // 类型断言
		if !ok {
			ResponseError(c, CodeInvalidParam)
			return
		}
		errData := removeTopStruct(errs.Translate(trans))
		ResponseErrorWithMsg(c, CodeInvalidParam, errData)
		return
	}
	// 具体的删除用户的业务逻辑
	if err := logic.DeleteMember(m); err != nil {
		zap.L().Error("logic.DeleteMember failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	// 返回响应
	ResponseSuccess(c, nil)
}

// GetListListHandler 获取当前用户创建的用户列表的处理函数
func GetListListHandler(c *gin.Context) {
	// 1.获取用户id
	userID, err := getCurrentUserID(c)
	if err != nil {
		ResponseError(c, CodeNeedLogin)
		return
	}
	// 获取分页参数
	page, size := getPageInfo(c)
	// 2.根据用户id获取创建的用户列表
	data, err := logic.GetListList(userID, page, size)
	if err != nil {
		zap.L().Error("logic.GetListList failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	// 3.返回响应
	ResponseSuccess(c, data)
}

// GetListDetailHandler 查看用户列表详情的处理函数
func GetListDetailHandler(c *gin.Context) {
	// 1.获取参数（从URL中获取列表的id）
	pidStr := c.Param("id")
	pid, err := strconv.ParseInt(pidStr, 10, 64)
	if err != nil {
		zap.L().Error("get list detail with invalid param", zap.Error(err))
		ResponseError(c, CodeInvalidParam)
		return
	}
	// 获取分页参数
	page, size := getPageInfo(c)
	// 2.根据id取出列表数据
	data, err := logic.GetListDetail(pid, page, size)
	if err != nil {
		zap.L().Error("logic.GetListDetail failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	// 3.返回响应
	ResponseSuccess(c, data)
}
