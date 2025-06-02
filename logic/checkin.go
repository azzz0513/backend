package logic

import (
	"errors"
	"fmt"
	"github.com/skip2/go-qrcode"
	"go.uber.org/zap"
	"math"
	"strconv"
	"web_app/dao/mysql"
	"web_app/models"
	"web_app/pkg/jwt"
	"web_app/pkg/snowflake"
)

// CreateCheckin 根据作者id创建打卡活动
func CreateCheckin(ck *models.Checkin) (err error) {
	// 1.生成checkin id
	ck.ID = snowflake.GenID()
	// 2.保存到数据库
	err = mysql.CreateCheckin(ck)
	if err != nil {
		zap.L().Error("mysql.CreateCheckin failed", zap.Error(err))
		return
	}
	// 3.返回
	return
}

// DeleteCheckin 根据打卡活动id删除指定的打卡活动
func DeleteCheckin(checkID int64) (err error) {
	zap.L().Debug("Delete Checkin", zap.Int64("checkID", checkID))
	return mysql.DeleteCheckin(checkID)
}

// GetCheckinDetailByID 根据活动id查询活动详情数据
func GetCheckinDetailByID(id, page, size int64) (data *models.CheckinDetail, err error) {
	data = &models.CheckinDetail{
		Checkin:          new(models.Checkin),
		UnCheckedMembers: make([]*models.UserEasyDetail, 0),
		CheckedMembers:   make([]*models.UserEasyDetail, 0),
	}
	// 根据活动id获取活动基础信息
	data.Checkin, err = mysql.GetCheckinMsg(id)
	if err != nil {
		zap.L().Error("mysql.GetCheckinMsg failed", zap.Error(err))
		return
	}
	// 根据checkin_id获取未完成打卡活动的成员
	data.UnCheckedCount, data.UnCheckedMembers, err = mysql.UnCheckedMember(id, page, size)
	if err != nil {
		zap.L().Error("mysql.UnCheckedMember failed", zap.Error(err))
		return
	}
	// 根据checkin_id获取已完成打卡活动的成员
	data.CheckedCount, data.CheckedMembers, err = mysql.CheckedMember(id, page, size)
	if err != nil {
		zap.L().Error("mysql.CheckedMember failed", zap.Error(err))
		return
	}
	return
}

// Participate 参与当前活动
func Participate(p *models.ParticipateMsg) (err error) {
	zap.L().Debug("Participate",
		zap.Int64("checkinID", p.CheckinID),
		zap.Int64("userID", p.UserID))
	// 判断用户验证码是否正确
	ok, err := mysql.CheckCheckinPassword(p.CheckinID, p.PassWord)
	if err != nil {
		zap.L().Error("mysql.CheckCheckinPassword failed", zap.Error(err))
		return
	}
	if !ok {
		return errors.New("活动验证码错误")
	}
	// 修改数据库
	return mysql.Participate(p.UserID, p.CheckinID)
}

// GetCheckinList 根据用户id获取当前用户需要参与的活动列表
func GetCheckinList(userID, page, size int64) (data []*models.MsgParticipant, err error) {
	data = make([]*models.MsgParticipant, 0)
	// 获取当前用户需要参与的活动列表
	checkins, err := mysql.GetCheckinList(userID, page, size)
	if err != nil {
		zap.L().Error("mysql.GetCheckinList failed", zap.Error(err))
		return
	}
	// 整合结构体
	for _, ck := range checkins {
		// 根据作者id查询作者信息
		author, err := mysql.GetUserByID(ck.AuthorID)
		if err != nil {
			zap.L().Error("mysql.GetUserByID failed", zap.Error(err))
			continue
		}
		// 根据活动种类id查询活动种类详细信息
		typeDetail, err := mysql.GetTypeDetailByID(ck.TypeID)
		if err != nil {
			zap.L().Error("mysql.GetTypeDetailByID failed", zap.Error(err))
			continue
		}
		// 根据打卡方式id查询打卡方式详细信息
		wayDetail, err := mysql.GetWayDetailByID(ck.WayID)
		if err != nil {
			zap.L().Error("mysql.GetWayDetailByID failed", zap.Error(err))
			continue
		}
		// 整合结构体
		checkinMsg := &models.CheckinMsg{
			CheckinID:  strconv.FormatInt(ck.ID, 10),
			Type:       typeDetail,
			Way:        wayDetail,
			Title:      ck.Title,
			Content:    ck.Content,
			CreateTime: ck.CreateTime,
			UpdateTime: ck.UpdateTime,
		}
		checkin := &models.MsgParticipant{
			AuthorName: author.Username,
			CheckinMsg: checkinMsg,
		}
		data = append(data, checkin)
	}
	return
}

// GetCreatedCheckinList 根据用户id获取当前用户创建的打卡活动列表
func GetCreatedCheckinList(userID, page, size int64) (data []*models.MsgCreator, err error) {
	data = make([]*models.MsgCreator, 0)
	// 获取当前用户需要参与的活动列表
	checkins, err := mysql.GetCreatedCheckinList(userID, page, size)
	if err != nil {
		zap.L().Error("mysql.GetCreatedCheckinList failed", zap.Error(err))
		return
	}
	// 整合结构体
	for _, ck := range checkins {
		listName, err := mysql.GetListDetailByID(ck.ListID)
		// 根据活动种类id查询活动种类详细信息
		typeDetail, err := mysql.GetTypeDetailByID(ck.TypeID)
		if err != nil {
			zap.L().Error("mysql.GetTypeDetailByID failed", zap.Error(err))
			continue
		}
		// 根据打卡方式id查询打卡方式详细信息
		wayDetail, err := mysql.GetWayDetailByID(ck.WayID)
		if err != nil {
			zap.L().Error("mysql.GetWayDetailByID failed", zap.Error(err))
			continue
		}
		// 整合结构体
		checkinMsg := &models.CheckinMsg{
			CheckinID:  strconv.FormatInt(ck.ID, 10),
			Type:       typeDetail,
			Way:        wayDetail,
			Title:      ck.Title,
			Content:    ck.Content,
			CreateTime: ck.CreateTime,
			UpdateTime: ck.UpdateTime,
		}
		checkin := &models.MsgCreator{
			ListName:   listName,
			CheckinMsg: checkinMsg,
		}
		data = append(data, checkin)
	}
	return
}

// GetParticipateDetail 获取用户当前参与的活动的详情数据
func GetParticipateDetail(id int64) (data *models.MsgParticipant, err error) {
	// 获取用户当前参与的活动的详情数据
	checkin, err := mysql.GetCheckinMsg(id)
	if err != nil {
		zap.L().Error("mysql.GetCheckinMsg failed", zap.Error(err))
		return nil, err
	}
	// 整合结构体
	author, err := mysql.GetUserByID(checkin.AuthorID)
	if err != nil {
		zap.L().Error("mysql.GetUserByID failed", zap.Error(err))
		return nil, err
	}
	// 根据活动种类id查询活动种类详细信息
	typeDetail, err := mysql.GetTypeDetailByID(checkin.TypeID)
	if err != nil {
		zap.L().Error("mysql.GetTypeDetailByID failed", zap.Error(err))
		return nil, err
	}
	// 根据打卡方式id查询打卡方式详细信息
	wayDetail, err := mysql.GetWayDetailByID(checkin.WayID)
	if err != nil {
		zap.L().Error("mysql.GetWayDetailByID failed", zap.Error(err))
		return nil, err
	}
	checkinMsg := &models.CheckinMsg{
		CheckinID:  strconv.FormatInt(checkin.ID, 10),
		Type:       typeDetail,
		Way:        wayDetail,
		Title:      checkin.Title,
		Content:    checkin.Content,
		CreateTime: checkin.CreateTime,
		UpdateTime: checkin.UpdateTime,
	}
	if checkin.TypeID == 1 {
		checkinMsg.StartTime = checkin.StartTime
		checkinMsg.DurationMinutes = checkin.DurationMinutes
	} else if checkin.TypeID == 2 {
		checkinMsg.StartDate = checkin.StartDate
		checkinMsg.EndDate = checkin.EndDate
		checkinMsg.DailyDeadline = checkin.DailyDeadline
	}
	data = &models.MsgParticipant{
		AuthorName: author.Username,
		CheckinMsg: checkinMsg,
	}
	return
}

// GetHistoryList 根据用户id获取当前用户参与过的打卡活动历史记录
func GetHistoryList(userID, page, size int64) (data []*models.MsgHistory, err error) {
	data = make([]*models.MsgHistory, 0)
	// 获取当前用户已参与的活动历史记录列表
	checkins, err := mysql.GetHistoryList(userID, page, size)
	if err != nil {
		zap.L().Error("mysql.GetCheckinList failed", zap.Error(err))
		return
	}
	// 整合结构体
	for _, ck := range checkins {
		// 根据作者id查询作者信息
		author, err := mysql.GetUserByID(ck.AuthorID)
		if err != nil {
			zap.L().Error("mysql.GetUserByID failed", zap.Error(err))
			continue
		}
		// 根据活动种类id查询活动种类详细信息
		typeDetail, err := mysql.GetTypeDetailByID(ck.TypeID)
		if err != nil {
			zap.L().Error("mysql.GetTypeDetailByID failed", zap.Error(err))
			continue
		}
		// 根据打卡方式id查询打卡方式详细信息
		wayDetail, err := mysql.GetWayDetailByID(ck.WayID)
		if err != nil {
			zap.L().Error("mysql.GetWayDetailByID failed", zap.Error(err))
			continue
		}
		// 根据打卡活动id和用户id获取打卡的时间
		checkTime, err := mysql.GetCheckTime(ck.ID, userID)
		if err != nil {
			zap.L().Error("mysql.GetCheckTime failed", zap.Error(err))
			continue
		}
		// 整合结构体
		checkinMsg := &models.CheckinMsg{
			CheckinID:  strconv.FormatInt(ck.ID, 10),
			Type:       typeDetail,
			Way:        wayDetail,
			Title:      ck.Title,
			Content:    ck.Content,
			CreateTime: ck.CreateTime,
			UpdateTime: ck.UpdateTime,
		}
		checkin := &models.MsgHistory{
			CheckTime: checkTime,
			MsgParticipant: &models.MsgParticipant{
				AuthorName: author.Username,
				CheckinMsg: checkinMsg,
			},
		}
		data = append(data, checkin)
	}
	return
}

// GetHistoryDetail 获取当前历史记录的详情
func GetHistoryDetail(id int64) (data *models.MsgParticipant, err error) {
	// 获取用户当前活动历史记录的详情数据
	checkin, err := mysql.GetCheckinMsg(id)
	if err != nil {
		zap.L().Error("mysql.GetCheckinMsg failed", zap.Error(err))
		return nil, err
	}
	// 整合结构体
	author, err := mysql.GetUserByID(checkin.AuthorID)
	if err != nil {
		zap.L().Error("mysql.GetUserByID failed", zap.Error(err))
		return nil, err
	}
	// 根据活动种类id查询活动种类详细信息
	typeDetail, err := mysql.GetTypeDetailByID(checkin.TypeID)
	if err != nil {
		zap.L().Error("mysql.GetTypeDetailByID failed", zap.Error(err))
		return nil, err
	}
	// 根据打卡方式id查询打卡方式详细信息
	wayDetail, err := mysql.GetWayDetailByID(checkin.WayID)
	if err != nil {
		zap.L().Error("mysql.GetWayDetailByID failed", zap.Error(err))
		return nil, err
	}
	checkinMsg := &models.CheckinMsg{
		CheckinID:  strconv.FormatInt(checkin.ID, 10),
		Type:       typeDetail,
		Way:        wayDetail,
		Title:      checkin.Title,
		Content:    checkin.Content,
		CreateTime: checkin.CreateTime,
		UpdateTime: checkin.UpdateTime,
	}
	if checkin.TypeID == 1 {
		checkinMsg.StartTime = checkin.StartTime
		checkinMsg.DurationMinutes = checkin.DurationMinutes
	} else if checkin.TypeID == 2 {
		checkinMsg.StartDate = checkin.StartDate
		checkinMsg.EndDate = checkin.EndDate
		checkinMsg.DailyDeadline = checkin.DailyDeadline
	}
	data = &models.MsgParticipant{
		AuthorName: author.Username,
		CheckinMsg: checkinMsg,
	}
	return
}

// GetStatistics 根据指定的类型获取统计数据
func GetStatistics(checkinID int64, statsType string) (data []*models.MsgStatistics, err error) {
	data = make([]*models.MsgStatistics, 0)
	// 从状态表中取出数据
	data, err = mysql.GetStatistics(checkinID, statsType)
	if err != nil {
		zap.L().Error("mysql.GetStatistics failed", zap.Error(err))
		return nil, err
	}
	// 整合结构体
	for _, stats := range data {
		user, err := mysql.GetUserByID(stats.UserID)
		if err != nil {
			zap.L().Error("mysql.GetUserByID failed", zap.Error(err))
			continue
		}
		stats.UserName = user.Username
	}
	return
}

// QRCode 生成二维码
func QRCode(checkinID int64) (data []byte, err error) {
	// 获取当前活动剩余时间
	duration, err := mysql.GetDuration(checkinID)
	if err != nil {
		zap.L().Error("mysql.GetDuration failed", zap.Error(err))
		return nil, err
	}
	// 生成活动指定的token
	token, err := jwt.GenCheckinToken(checkinID, duration)
	if err != nil {
		return
	}
	// 直接生成签到页面URL
	url := fmt.Sprintf("http://8.138.230.142:8087/qr_checkin.html?token=%s", token)
	return qrcode.Encode(url, qrcode.Medium, 256)
}

// GetUserInfo 检验当前用户是否有参与当前活动的权限
func GetUserInfo(checkinID, userID int64) (data *models.User, err error) {
	zap.L().Debug("GetUserInfo",
		zap.Int64("checkinID", checkinID),
		zap.Int64("userID", userID))
	return mysql.GetUserInfo(checkinID, userID)
}

// QRCheckin 二维码签到
func QRCheckin(p *models.ParticipateMsg) (err error) {
	zap.L().Debug("QRCheckin",
		zap.Int64("userID", p.UserID),
		zap.Int64("checkinID", p.CheckinID))
	return mysql.Participate(p.UserID, p.CheckinID)
}

// PositionCheckin 定位签到
func PositionCheckin(r *models.PosCheckin) (err error) {
	zap.L().Debug("GeoCheckin",
		zap.Int64("userID", r.UserID),
		zap.Int64("checkinID", r.CheckinID))
	// 获取当前活动的定位范围
	oLat, oLng, oRadius, err := mysql.GetRangeByID(r.CheckinID)
	if err != nil {
		zap.L().Error("mysql.GetRangeByID failed", zap.Error(err))
		return err
	}
	// 判断距离
	distance := calculateDistance(oLat, oLng, r.Lat, r.Lng)
	if distance <= oRadius {
		return mysql.Participate(r.UserID, r.CheckinID)
	}
	return errors.New("打卡距离超出范围")
}

// 计算距离
func calculateDistance(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371000
	φ1 := lat1 * math.Pi / 180
	φ2 := lat2 * math.Pi / 180
	Δφ := (lat2 - lat1) * math.Pi / 180
	Δλ := (lng2 - lng1) * math.Pi / 180

	a := math.Sin(Δφ/2)*math.Sin(Δφ/2) +
		math.Cos(φ1)*math.Cos(φ2)*
			math.Sin(Δλ/2)*math.Sin(Δλ/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}
