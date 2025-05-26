package controller

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"strconv"
	"web_app/logic"
	"web_app/models"
	"web_app/pkg/jwt"
)

// CreateCheckinHandler 处理打卡活动发布
// @Tags 打卡活动管理
// @Summary 发布新的打卡活动
// @Description 接收前端数据创建新的打卡活动
// @Router /api/v1/checkin [post]
// @Param request body models.Checkin true  "创建打卡活动参数"
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} ResponseData "成功响应示例：{"code":1000,"msg":"业务处理成功","data":null}"
func CreateCheckinHandler(c *gin.Context) {
	// 1.获取当前用户的请求参数和参数校验
	ck := new(models.Checkin)
	if err := c.ShouldBindJSON(ck); err != nil {
		zap.L().Debug("c.ShouldBindJSON(p) err", zap.Any("err", err))
		zap.L().Error("create checkin activity failed with invalid param", zap.Error(err))
		ResponseError(c, CodeInvalidParam)
		return
	}
	// 从c中取得当前发送请求的用户的id
	userID, err := getCurrentUserID(c)
	if err != nil {
		ResponseError(c, CodeNeedLogin)
		return
	}
	ck.AuthorID = userID
	// 2.业务处理
	if err := logic.CreateCheckin(ck); err != nil {
		zap.L().Error("logic.Publish(ck) err", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	// 3.返回响应
	if ck.WayID == 2 {
		// 处理生成二维码的具体逻辑
		qrBytes, err := logic.QRCode(ck.ID, ck.DurationMinutes)
		if err != nil {
			zap.L().Error("logic.QRCode err", zap.Error(err))
			ResponseError(c, CodeServerBusy)
			return
		}
		// 返回响应
		ResponseSuccessWithPng(c, qrBytes)
	}
	ResponseSuccess(c, nil)
}

// DeleteCheckinHandler 处理打卡活动删除
// @Tags 打卡活动管理
// @Summary 删除打卡活动
// @Description 接收前端数据删除指定的打卡活动
// @Router /api/v1/delete_checkin/{id} [post]
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} ResponseData "成功响应示例：{"code":1000,"msg":"业务处理成功","data":null}"
func DeleteCheckinHandler(c *gin.Context) {
	// 1.获取参数（从URL中获取当前打卡活动的id）
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		zap.L().Error("controller.DeleteCheckinHandler failed", zap.Error(err))
		ResponseError(c, CodeInvalidParam)
		return
	}
	// 2.处理打卡活动删除的具体逻辑
	if err := logic.DeleteCheckin(id); err != nil {
		zap.L().Error("logic.DeleteCheckin failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	// 返回响应
	ResponseSuccess(c, nil)
}

// GetCheckinDetailHandler 处理获取当前打卡活动的详情
// @Tags 打卡活动管理
// @Summary 获取当前打卡活动的详情（已打卡人数以及未打卡人员名单）
// @Description 获取当前打卡活动的详情并发送到前端
// @Router /api/v1/checkin/{id} [get]
// @Param id path string true "活动ID"
// @Param page query int true "页码"
// @Param size query int true "页面大小参数"
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} ResponseData "成功响应示例：{"code":1000,"msg":"业务处理成功","data":models.CheckinDetail}"
func GetCheckinDetailHandler(c *gin.Context) {
	// 1.获取参数（从URL中获取当前打卡活动的id）
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		zap.L().Error("controller.GetCheckinDetailHandler err", zap.Error(err))
		ResponseError(c, CodeInvalidParam)
		return
	}
	// 2.获取分页参数
	page, size := getPageInfo(c)
	// 2.根据id取出打卡活动当前的活动详情
	data, err := logic.GetCheckinDetailByID(id, page, size)
	if err != nil {
		zap.L().Error("logic.GetCheckinDetailByID(id) err", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	// 3.返回响应
	ResponseSuccess(c, data)
}

// GetParticipateDetailHandler 处理获取用户当前参与的活动的详情
// @Tags 打卡活动管理
// @Summary 获取用户当前参与的活动的详情
// @Description 获取当前打卡活动的详情并发送到前端
// @Router /api/v1/participate_detail/{id} [get]
// @Param id path string true "活动ID"
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} ResponseData "成功响应示例：{"code":1000,"msg":"业务处理成功","data":models.MsgParticipant}"
func GetParticipateDetailHandler(c *gin.Context) {
	// 1.获取参数（从URL中获取当前打卡活动的id）
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		zap.L().Error("controller.GetParticipateDetailHandler err", zap.Error(err))
		ResponseError(c, CodeInvalidParam)
		return
	}
	// 2.根据id取出打卡活动当前的活动详情
	data, err := logic.GetParticipateDetail(id)
	if err != nil {
		zap.L().Error("logic.GetParticipateDetailByID(id) err", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	// 3.返回响应
	ResponseSuccess(c, data)
}

// ParticipateHandler 处理参与打卡活动
// @Tags 打卡活动管理
// @Summary 参与打卡活动
// @Description 接收前端数据参与指定打卡活动
// @Router /api/v1/participate/{id} [post]
// @Param id path string true "活动ID"
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} ResponseData "成功响应示例：{"code":1000,"msg":"业务处理成功","data":null}"
func ParticipateHandler(c *gin.Context) {
	// 获取请求参数
	p := new(models.ParticipateMsg)
	if err := c.ShouldBindJSON(p); err != nil {
		zap.L().Error("controller.ParticipateHandler err", zap.Error(err))
		ResponseError(c, CodeInvalidParam)
		return
	}
	// 获取参数（从URL中获取当前活动id）
	checkinIDStr := c.Param("id")
	checkinID, err := strconv.ParseInt(checkinIDStr, 10, 64)
	if err != nil {
		zap.L().Error("controller.ParticipateHandler err", zap.Error(err))
		ResponseError(c, CodeInvalidParam)
		return
	}
	p.CheckinID = checkinID
	// 获取当前用户的id
	userID, err := getCurrentUserID(c)
	if err != nil {
		ResponseError(c, CodeNeedLogin)
		return
	}
	p.UserID = userID
	// 处理参与打卡活动
	if err := logic.Participate(p); err != nil {
		zap.L().Error("logic.Participate err", zap.Error(err))
		ResponseError(c, CodeTimeOut)
		return
	}
	// 返回响应
	ResponseSuccess(c, nil)
}

// GetCheckinListHandler 处理获取当前用户参与的打卡活动列表
// @Tags 打卡活动管理
// @Summary 获取当前用户需要参与的打卡活动列表
// @Description 获取当前用户需要参与的打卡活动列表并发送到前端
// @Router /api/v1/checkin_list [get]
// @Param page query int true "页码"
// @Param size query int true "页面大小参数"
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} ResponseData "成功响应示例：{"code":1000,"msg":"业务处理成功","data":[]*models.MsgParticipant}"
func GetCheckinListHandler(c *gin.Context) {
	// 1.获取当前用户的id
	userID, err := getCurrentUserID(c)
	if err != nil {
		ResponseError(c, CodeNeedLogin)
		return
	}
	// 2.获取分页参数
	page, size := getPageInfo(c)
	// 3.根据用户id获取当前用户需要参与的打卡活动列表
	data, err := logic.GetCheckinList(userID, page, size)
	if err != nil {
		zap.L().Error("logic.GetCheckinList(userID, page, size) err", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	// 返回响应
	ResponseSuccess(c, data)
}

// GetCreatedCheckinListHandler 处理获取当前用户创建的打卡活动列表
// @Tags 打卡活动管理
// @Summary 获取当前用户创建的打卡活动列表
// @Description 获取当前用户创建的打卡活动列表并发送到前端
// @Router /api/v1/created_list [get]
// @Param page query int true "页码"
// @Param size query int true "页面大小参数"
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} ResponseData "成功响应示例：{"code":1000,"msg":"业务处理成功","data":[]*models.MsgCreator}"
func GetCreatedCheckinListHandler(c *gin.Context) {
	// 1.获取当前用户的id
	userID, err := getCurrentUserID(c)
	if err != nil {
		ResponseError(c, CodeNeedLogin)
		return
	}
	// 2.获取分页参数
	page, size := getPageInfo(c)
	// 3.根据用户id获取当前用户创建的打卡活动列表
	data, err := logic.GetCreatedCheckinList(userID, page, size)
	if err != nil {
		zap.L().Error("logic.GetCreatedCheckinList(userID, page, size) err", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	// 返回响应
	ResponseSuccess(c, data)
}

// GetHistoryListHandler 处理获取当前用户已参与过的打卡活动历史记录列表
// @Tags 打卡活动管理
// @Summary 获取当前用户参与的打卡活动历史记录
// @Description 获取当前用户参与的打卡活动历史记录并发送到前端
// @Router /api/v1/created_list [get]
// @Param page query int true "页码"
// @Param size query int true "页面大小参数"
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} ResponseData "成功响应示例：{"code":1000,"msg":"业务处理成功","data":[]*models.MsgHistory}"
func GetHistoryListHandler(c *gin.Context) {
	// 1.获取当前用户的id
	userID, err := getCurrentUserID(c)
	if err != nil {
		ResponseError(c, CodeNeedLogin)
		return
	}
	// 2.获取分页参数
	page, size := getPageInfo(c)
	// 3.根据用户id获取当前用户已打卡的打卡活动历史记录列表
	data, err := logic.GetHistoryList(userID, page, size)
	if err != nil {
		zap.L().Error("logic.GetHistoryList(userID, page, size) err", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	// 返回响应
	ResponseSuccess(c, data)
}

// GetHistoryDetailHandler 处理获取用户当前参与的活动的详情
// @Tags 打卡活动管理
// @Summary 获取用户参与的活动历史记录的详情
// @Description 获取当前历史记录活动的详情并发送到前端
// @Router /api/v1/history/{id} [get]
// @Param id path string true "活动ID"
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} ResponseData "成功响应示例：{"code":1000,"msg":"业务处理成功","data":models.MsgParticipant}"
func GetHistoryDetailHandler(c *gin.Context) {
	// 1.获取参数（从URL中获取当前打卡活动的id）
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		zap.L().Error("controller.GetHistoryDetailHandler err", zap.Error(err))
		ResponseError(c, CodeInvalidParam)
		return
	}
	// 2.根据id取出打卡活动当前的活动详情
	data, err := logic.GetHistoryDetail(id)
	if err != nil {
		zap.L().Error("logic.GetHistoryDetailByID(id) err", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	// 3.返回响应
	ResponseSuccess(c, data)
}

// GetStatisticsHandler 处理获取用户创建的长期考勤活动的数据
// @Tags 打卡活动管理
// @Summary 获取当前打卡活动的统计数据
// @Description 获取当前打卡活动的统计数据并发送到前端
// @Router /api/v1/statistics/{id} [get]
// @Param id path string true "活动ID"
// @Param type query string true "数据种类"
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} ResponseData "成功响应示例：{"code":1000,"msg":"业务处理成功","data":[]*models.MsgStatistics}"
func GetStatisticsHandler(c *gin.Context) {
	// 获取请求参数
	t := new(models.StatisticsType)
	if err := c.ShouldBindQuery(t); err != nil {
		zap.L().Error("controller.GetStatisticsHandler err", zap.Error(err))
		ResponseError(c, CodeInvalidParam)
		return
	}
	// 获取参数（从URL中获取当前打卡活动的id）
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		zap.L().Error("controller.GetStatisticsHandler err", zap.Error(err))
		ResponseError(c, CodeInvalidParam)
		return
	}
	// 根据id取出打卡活动的数据统计
	data, err := logic.GetStatistics(id, t.Type)
	if err != nil {
		zap.L().Error("logic.GetStatistics err", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	// 返回响应
	ResponseSuccess(c, data)
}

// QRCheckinHandler 处理扫描二维码后签到
// @Tags 打卡活动管理
// @Summary 处理二维码签到
// @Description 处理二维码签到
// @Router /api/v1/qr_checkin [post]
// @Param token query string true "Token"
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} ResponseData "成功响应示例：{"code":1000,"msg":"业务处理成功","data":null}"
func QRCheckinHandler(c *gin.Context) {
	// 获取请求参数
	p := new(models.ParticipateMsg)
	token := c.Query("token")
	// 解析JWT令牌
	ckc, err := jwt.ParseCheckinToken(token)
	if err != nil {
		zap.L().Error("controller.QRCheckinHandler err", zap.Error(err))
		ResponseError(c, CodeInvalidToken)
		return
	}
	p.CheckinID = ckc.CheckinID
	// 获取当前用户的id
	userID, err := getCurrentUserID(c)
	if err != nil {
		zap.L().Error("controller.QRCheckinHandler err", zap.Error(err))
		ResponseError(c, CodeNeedLogin)
		return
	}
	p.UserID = userID
	// 执行签到的具体逻辑
	if err := logic.QRCheckin(p); err != nil {
		zap.L().Error("controller.QRCheckin err", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	// 返回响应
	ResponseSuccess(c, nil)
}

// PositionCheckinHandler 处理定位打卡活动
// @Tags 打卡活动管理
// @Summary 处理定位打卡
// @Description 接收前端数据处理定位打卡
// @Router /api/v1/position_checkin/{id} [post]
// @Param id path string true "活动ID"
// @Param request body models.PosCheckin true  "定位打卡请求参数"
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} ResponseData "成功响应示例：{"code":1000,"msg":"业务处理成功","data":null}"
func PositionCheckinHandler(c *gin.Context) {
	// 获取请求参数
	r := new(models.PosCheckin)
	if err := c.ShouldBindJSON(r); err != nil {
		zap.L().Error("controller.GeoCheckinHandler err", zap.Error(err))
		ResponseError(c, CodeInvalidParam)
		return
	}
	// 获取当前用户id
	userID, err := getCurrentUserID(c)
	if err != nil {
		zap.L().Error("controller.GeoCheckinHandler err", zap.Error(err))
		ResponseError(c, CodeNeedLogin)
		return
	}
	r.UserID = userID
	// 处理定位签到的具体逻辑
	if err := logic.PositionCheckin(r); err != nil {
		zap.L().Error("controller.GeoCheckin err", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	// 返回响应
	ResponseSuccess(c, nil)
}
