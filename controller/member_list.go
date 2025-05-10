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
// @Tags 用户列表管理
// @Summary 创建用户列表
// @Description 接收前端数据创建用户列表
// @Param request body models.MemberList true  "创建用户列表参数"
// @Router /api/v1/create_member_list [post]
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} ResponseData "成功响应示例：{"code":1000,"msg":"业务处理成功","data":null}"
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
// @Tags 用户列表管理
// @Summary 添加用户
// @Description 接收前端数据添加用户
// @Param request body models.UpdateMember true  "修改用户列表参数"
// @Router /api/v1/add_member [post]
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} ResponseData "成功响应示例：{"code":1000,"msg":"业务处理成功","data":null}"
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
// @Tags 用户列表管理
// @Summary 删除用户
// @Description 接收前端数据删除用户
// @Param request body models.UpdateMember true  "修改用户列表参数"
// @Router /api/v1/delete_member [post]
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} ResponseData "成功响应示例：{"code":1000,"msg":"业务处理成功","data":null}"
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
	// 具体地删除用户的业务逻辑
	if err := logic.DeleteMember(m); err != nil {
		zap.L().Error("logic.DeleteMember failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	// 返回响应
	ResponseSuccess(c, nil)
}

// GetListListHandler 获取当前用户创建的用户列表的处理函数
// @Tags 用户列表管理
// @Summary 获取当前用户创建的用户列表
// @Description 接收前端数据查找当前用户创建的用户列表
// @Router /api/v1/member_list [get]
// @Param page query int true "页码"
// @Param size query int true "页面大小参数"
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} ResponseData "成功响应示例：{"code":1000,"msg":"业务处理成功","data":[]*models.MemberList}"
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
// @Tags 用户列表管理
// @Summary 获取当前用户列表的详情
// @Description 接收前端数据查找当前用户列表的详情
// @Router /api/v1/member_list/{id} [get]
// @Param id path string true "当前列表id"
// @Param page query int true "页码"
// @Param size query int true "页面大小参数"
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} ResponseData "成功响应示例：{"code":1000,"msg":"业务处理成功","data":[]*models.ListDetail}"
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
