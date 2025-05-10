package logic

import (
	"go.uber.org/zap"
	"web_app/dao/mysql"
	"web_app/models"
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

// GetCheckinDetailByID 根据活动id查询活动详情数据
func GetCheckinDetailByID(id, page, size int64) (data *models.CheckinDetail, err error) {
	data = &models.CheckinDetail{
		Checkin: new(models.Checkin),
		Members: make([]*models.UserDetail, 0),
	}
	// 根据活动id获取活动基础信息
	data.Checkin, err = mysql.GetCheckinMsg(id)
	if err != nil {
		zap.L().Error("mysql.GetCheckinMsg failed", zap.Error(err))
		return
	}
	// 根据checkin_id获取未完成打卡活动的成员
	data.Count, data.Members, err = mysql.CheckMember(id, page, size)
	if err != nil {
		zap.L().Error("mysql.CheckMember failed", zap.Error(err))
		return
	}
	return
}

// Participate 参与当前活动
func Participate(userID, checkinID int64) (err error) {
	// 修改数据库
	return mysql.Participate(userID, checkinID)
}

// GetCheckinList 根据用户id获取当前用户需要参与的活动列表
func GetCheckinList(userID, page, size int64) (data []*models.MsgParticipant, err error) {
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
			CheckinID:  ck.ID,
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
			CheckinID:  ck.ID,
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

func GetHistoryList(userID, page, size int64) (data []*models.MsgHistory, err error) {
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
			CheckinID:  ck.ID,
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
