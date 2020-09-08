package controller

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/labstack/echo/v4"
	"github.com/wppzxc/wechat-tools/pkg/config"
	"github.com/wppzxc/wechat-tools/pkg/database"
	"github.com/wppzxc/wechat-tools/pkg/prometheus"
	"github.com/wppzxc/wechat-tools/pkg/types"
	"github.com/wppzxc/wechat-tools/pkg/utils"
	"github.com/wppzxc/wechat-tools/pkg/wechat"
	"k8s.io/klog"
)

func (ctl *Controller) server(c echo.Context) error {
	defer prometheus.ReceiveMsgs.Inc()

	// 程序还未开始运行，则跳过 http 请求
	if !config.GetRunning() {
		klog.Info("程序未开始运行，跳过请求")
		return nil
	}

	reqParams := getRequestParams(c)
	if reqParams == nil {
		return fmt.Errorf("Error in get request query data ")
	}

	klog.Infof("Get request data : %+v", reqParams)
	if err := ctl.doForward(reqParams); err != nil {
		return err
	}
	return nil
}

// get RequestParams from request body
func getRequestParams(c echo.Context) *types.RequestParam {
	defer c.Request().Body.Close()
	reqData := new(types.RequestParam)
	data, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		klog.Error(err)
		return nil
	}
	if err := json.Unmarshal(data, reqData); err != nil {
		return nil
	}

	return reqData
}

// 处理不通的消息类型
func (ctl *Controller) doForward(reqParam *types.RequestParam) error {
	var execute = false
	defer func() {
		if execute {
			prometheus.ExecuteMsgs.Inc()
		}
	}()

	// 处理群成员增减消息
	if reqParam.Event == types.EventGroupMemberAdd || reqParam.Event == types.EventGroupMemberDecrease {
		execute = true
		return ctl.execGroupMemberChange(reqParam)
	}

	// 处理群消息
	if reqParam.Event == types.EventGroupMsg {
		execute = true
		// return ctl.execGroupMsg(reqParam)
		// 处理创建淘礼金消息
		return ctl.execGroupCreateTaolijin(reqParam)
	}

	// 处理私聊消息
	if reqParam.Event == types.EventFriendMsg {
		execute = true
		return ctl.execFriendMsg(reqParam)
	}

	// 处理好友请求
	if reqParam.Event == types.EventFriendVerify {
		execute = true
		go func(reqParams *types.RequestParam) {
			if err := ctl.execFriendVerify(reqParam); err != nil {
				klog.Error(err)
				prometheus.AgreeFailedUsers.Inc()
				return
			}
			prometheus.AgreeSuccessUsers.Inc()
		}(reqParam)
		return nil
	}

	// 其他类型不处理
	klog.Infof("unknown msg request : %+v", reqParam)
	return nil
}

func (ctl *Controller) execGroupMsg(reqParams *types.RequestParam) error {
	reqUser, err := database.GetGroupUserByWxid(reqParams.FromWxid, reqParams.FinalFromWxid)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			reqUser = &database.User{
				GroupWxid:        reqParams.FromWxid,
				NickName:         reqParams.FinalFromName,
				Wxid:             reqParams.FinalFromWxid,
				InviteUserNumber: 0,
				Alerted:          false,
				Role:             database.UserRoleNormal,
			}
		} else {
			klog.Error(err)
			return err
		}
	}

	// 是否是管理员消息
	// at消息，且是管理员或者全局管理员发出的消息
	_, _, _, yes := utils.IsAtMsg(reqParams.Msg)
	if yes && utils.UsersContain(config.GlobalConfig.InviteMangerConf.ManageGroups, reqParams.FromWxid) &&
		(reqUser.Role == database.UserRoleManager || reqUser.Role == database.UserRoleOwner) {
		return ctl.execManagerReq(reqParams)
	}

	// 发言是否违规：管理群的消息，管理的群，且普通用户发出
	if reqParams.Event == types.EventGroupMsg && utils.UsersContain(config.GlobalConfig.InviteMangerConf.ManageGroups, reqParams.FromWxid) &&
		reqUser.Role == database.UserRoleNormal {
		if kickOut := ctl.judgeMsgKickOut(reqParams); kickOut {
			msg := fmt.Sprintf("此人%s(%s)违反群规已被踢出群聊，并永久加入黑名单，如有误踢请联系群主或管理员用户", reqParams.FinalFromName, reqParams.FinalFromWxid)
			// 发出踢出提示
			ctl.enqueueSendMsg(utils.TextMsgSendParam(msg, reqParams.FromWxid))
			// 踢出用户
			ctl.enqueueSendAction(utils.RemoveMsgSendParam(reqParams.FromWxid, reqParams.FinalFromWxid))
			prometheus.AutoRemoveUsers.Inc()
			// 并拉黑
			err := database.CreateBlackList(&database.User{
				Wxid: reqParams.FinalFromWxid,
			})
			if err != nil {
				klog.Error(err)
			} else {
				prometheus.AutoBlackUsers.Inc()
			}
			return nil
		}
	}

	// 是否需要设置活跃用户
	if utils.UsersContain(config.GlobalConfig.InviteMangerConf.ManageGroups, reqParams.FromWxid) {
		// 如果发言内容包含 “已拍”，且未设置活跃用户
		if strings.Index(reqParams.Msg, "已拍") >= 0 && !reqUser.Active {
			reqUser.Active = true
			if err := database.UpdateGroupUserByWxid(*reqUser, reqUser.GroupWxid, reqUser.Wxid); err != nil {
				klog.Errorf("设置活跃用户%s(%s)失败: %s", reqUser.NickName, reqUser.Wxid, err)
			}
			klog.Infof("设置活跃用户%s(%s)成功", reqUser.NickName, reqUser.Wxid)
		}
	}

	// 是否需要转发群消息
	if config.GlobalConfig.SendReceiveConf.StartSendReceiver {
		if needToSend := ctl.needToSend(reqParams); needToSend {
			return ctl.transGroupMsg(reqParams)
		}
		klog.V(3).Info("消息不满足过滤条件，跳过消息: %+v", reqParams)
	}
	return nil
}

func (ctl *Controller) execFriendMsg(reqParams *types.RequestParam) error {
	klog.V(3).Infof("收到私聊消息: %s", reqParams.Msg)
	// 添加好友成功之后
	if reqParams.Type == types.FriendWelcomeMsg {
		// 自动回复消息和发送图片
		ctl.enqueueSendMsg(utils.TextMsgSendParam(config.GlobalConfig.AutoAgreeFriendVerifyConf.WelcomeMsg, reqParams.FromWxid))
		prometheus.SendWelcomeTexts.Inc()

		welcomeImagePath, _ := os.Getwd()
		// welcomeImagePath := "C:\\Users\\Administrator\\Desktop\\可爱猫4.4.0"
		ctl.enqueueSendAction(utils.ImageMsgSendParam(fmt.Sprintf("%s\\invite.jpg", welcomeImagePath), reqParams.FromWxid))
		prometheus.SendWelcomeImages.Inc()
		// 自动拉到群聊
		for index, g := range config.GlobalConfig.AutoAgreeFriendVerifyConf.AutoInviteGroups {
			users, err := wechat.GetGroupUserList(config.GlobalConfig.LocalUser.Wxid, g.Wxid)
			if err != nil {
				klog.Errorf("获取群成员失败: %s", err)
				continue
			}
			if len(users) == 0 || len(users) > 480 {
				klog.Warningf("自动邀请群组%s(%s)人员超过480，跳过该群", g.Name, g.Wxid)
				newAutoInviteGroups := make([]config.CommonUserInfo, 0)
				for i, ag := range config.GlobalConfig.AutoAgreeFriendVerifyConf.AutoInviteGroups {
					if i == index {
						continue
					}
					newAutoInviteGroups = append(newAutoInviteGroups, ag)
				}
				config.GlobalConfig.AutoAgreeFriendVerifyConf.AutoInviteGroups = newAutoInviteGroups
				continue
			}
			if err := wechat.InviteInGroup(config.GlobalConfig.LocalUser.Wxid, g.Wxid, reqParams.FromWxid); err != nil {
				klog.Errorf("自动邀请用户%s(%s)进入群组%s(%s)失败: %s", reqParams.FromName, reqParams.FromWxid, g.Name, g.Wxid, err)
				continue
			}
			time.Sleep(time.Duration(config.GlobalConfig.SendReceiveConf.ActionInterval) * time.Second)
		}
	} else {
		// TODO: 自动回复消息
	}
	return nil
}

func (ctl *Controller) execFriendVerify(reqParams *types.RequestParam) error {
	ctl.agreeLock.Lock()
	defer ctl.agreeLock.Unlock()
	klog.Infof("收到好友请求: %s", reqParams.JsonMsg)
	klog.Info("暂停3秒钟")
	time.Sleep(3 * time.Second)
	// 自动同意好友请求
	if err := wechat.AgreeFriendVerify(config.GlobalConfig.LocalUser.RobotWxid, reqParams.JsonMsg); err != nil {
		klog.Error(err)
		return err
	}
	return nil
}

// 转发群消息
func (ctl *Controller) transGroupMsg(reqParams *types.RequestParam) error {
	switch reqParams.Type {
	case types.SystemMsg:
		// 系统消息已经被 type=400 过滤掉了，其他系统消息此处并不进行处理
		return nil
	case types.TextMsg:
		return ctl.SendText(reqParams.Msg)
	case types.ImageMsg:
		return ctl.SendImageFile(reqParams.Msg)
	case types.VideoMsg:
		return ctl.SendVideoFile(reqParams.Msg)
	default:
		return fmt.Errorf("Unhandled msg type : %d ", reqParams.Type)
	}
}

// 预处理消息内容，
// 是否包含关键词，
// 是否包含过滤关键词
func (ctl *Controller) preExecuteMsg(msg string) (string, error) {
	containKeywords := false
	if len(config.GlobalConfig.SendReceiveConf.Keywords) == 1 && config.GlobalConfig.SendReceiveConf.Keywords[0] == "" {
		containKeywords = true
	} else if len(config.GlobalConfig.SendReceiveConf.Keywords) > 0 {
		for _, key := range config.GlobalConfig.SendReceiveConf.Keywords {
			if index := strings.Index(msg, key); index >= 0 {
				containKeywords = true
			}
		}
	} else {
		containKeywords = true
	}

	if !containKeywords {
		return "", fmt.Errorf("Skip it, msg %s don't container keywords %+v ", msg, config.GlobalConfig.SendReceiveConf.Keywords)
	}

	containFilterKeywords := false
	if len(config.GlobalConfig.SendReceiveConf.FilterKeywords) == 1 && config.GlobalConfig.SendReceiveConf.FilterKeywords[0] == "" {
		containFilterKeywords = false
	} else if len(config.GlobalConfig.SendReceiveConf.FilterKeywords) > 0 {
		for _, filterKey := range config.GlobalConfig.SendReceiveConf.FilterKeywords {
			if index := strings.Index(msg, filterKey); index >= 0 {
				containFilterKeywords = true
			}
		}
	}

	if containFilterKeywords {
		return "", fmt.Errorf("Skip it, msg %s container keywords %+v ", msg, config.GlobalConfig.SendReceiveConf.FilterKeywords)
	}

	// 转换 ￥ $
	if config.GlobalConfig.SendReceiveConf.TranMoneySep {
		msg = utils.TranMoneySep(msg)
	}
	return msg, nil
}

// SendText 发送文字，包含预处理
func (ctl *Controller) SendText(msg string) error {
	atNickname, atWwxid, sendMsg, yes := utils.IsAtMsg(msg)
	if yes {
		// TODO: 待开发管理员功能
		klog.Infof("TODO: receive group at msg to_nickname : %s, to_wxid : %s, msg : %s", atNickname, atWwxid, sendMsg)
		return nil
	}

	result, err := ctl.preExecuteMsg(msg)
	if err != nil {
		klog.Error(err)
		return nil
	}
	defer prometheus.TransTexts.Inc()
	return ctl.sendTextMsg(result)
}

// SendTextMsg 发送文字消息
func (ctl *Controller) sendTextMsg(msg string) error {
	for _, user := range config.GlobalConfig.SendReceiveConf.SendToUsers {
		sendParam := types.SendParam{
			Api:       types.SendTextMsgApi,
			Msg:       msg,
			RobotWxid: config.GlobalConfig.LocalUser.Wxid,
			ToWxid:    user.Wxid,
		}
		ctl.enqueueSendMsg(sendParam)
	}
	klog.Infof("finish send text msg to users %+v", config.GlobalConfig.SendReceiveConf.SendToUsers)
	return nil
}

// SendImageFile 发送图片消息
func (ctl *Controller) SendImageFile(filePath string) error {
	for _, user := range config.GlobalConfig.SendReceiveConf.SendToUsers {
		sendParam := types.SendParam{
			Api:       types.SendImageMsgApi,
			Path:      filePath,
			RobotWxid: config.GlobalConfig.LocalUser.Wxid,
			ToWxid:    user.Wxid,
		}
		ctl.enqueueSendMsg(sendParam)
	}
	defer prometheus.TransImages.Add(float64(len(config.GlobalConfig.SendReceiveConf.SendToUsers)))
	klog.Infof("finish send image msg to users %+v", config.GlobalConfig.SendReceiveConf.SendToUsers)
	return nil
}

// SendVideoFile 发送视频消息
func (ctl *Controller) SendVideoFile(filePath string) error {
	for _, user := range config.GlobalConfig.SendReceiveConf.SendToUsers {
		sendParam := types.SendParam{
			Api:       types.SendVideoMsgApi,
			Path:      filePath,
			RobotWxid: config.GlobalConfig.LocalUser.Wxid,
			ToWxid:    user.Wxid,
		}
		ctl.enqueueSendMsg(sendParam)
	}
	defer prometheus.TransVideos.Add(float64(len(config.GlobalConfig.SendReceiveConf.SendToUsers)))
	klog.Infof("finish send video msg to users %+v", config.GlobalConfig.SendReceiveConf.SendToUsers)
	return nil
}

func (ctl *Controller) needToSend(reqParams *types.RequestParam) bool {
	if reqParams.Event != types.EventGroupMsg {
		klog.Info("当前版本只处理群消息")
		return false
	}
	if reqParams.Type == types.SystemMsg {
		klog.Warning("当前版本不处理群系统消息")
		return false
	}

	groupUserInfo, ok := config.GlobalConfig.SendReceiveConf.ReceiveFromGroup[reqParams.FromWxid]
	if !ok {
		return false
	}

	send := false
	for _, u := range groupUserInfo.Users {
		if reqParams.FinalFromWxid == u.Wxid {
			send = true
			break
		}
	}
	return send
}
