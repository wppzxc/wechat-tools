package controller

import (
	"encoding/json"
	"fmt"
	"path"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/wppzxc/wechat-tools/pkg/config"
	"github.com/wppzxc/wechat-tools/pkg/database"
	"github.com/wppzxc/wechat-tools/pkg/front"
	"github.com/wppzxc/wechat-tools/pkg/types"
	"github.com/wppzxc/wechat-tools/pkg/utils"
	"github.com/wppzxc/wechat-tools/pkg/wechat"
	"k8s.io/klog"
)

var TaobaoClient *utils.TaoBaoClient
var DataokeClient *utils.DaTaoKeClient

func (ctl *Controller) execManagerReq(reqParam *types.RequestParam) error {
	_, _, msg, _ := utils.IsAtMsg(reqParam.Msg)
	switch msg {
	case types.ManageMsgRemoveUser:
		return ctl.manageRemoveGroupUser(reqParam)
	case types.ManageMsgSetManager:
		return ctl.setGroupUserManager(reqParam)
	case types.ManageMsgRemoveManager:
		return ctl.removeGroupUserManager(reqParam)
	case types.ManageMsgSetVip:
		return ctl.setGroupUserVip(reqParam)
	case types.ManageMsgSetUserInviteNum:
		return ctl.SetUserInviteNum(reqParam)
	case types.ManageMsgHealthCheck:
		return ctl.healthCheck(reqParam)
	case types.ManageMsgInviteNumCheck:
		return ctl.inviteNumCheck(reqParam)
	default:
		klog.Infof("不支持的指令: %s", msg)
		return nil
	}
}

func (ctl *Controller) execGroupMemberChange(param *types.RequestParam) error {
	switch param.Event {
	case types.EventGroupMemberAdd:
		return ctl.execGroupUserAdd(param.JsonMsg)
	case types.EventGroupMemberDecrease:
		return execGroupUserDec(param.JsonMsg)
	}
	return nil
}

func (ctl *Controller) execGroupUserAdd(jsonMsg string) error {
	addMsg := new(types.GroupUserAddJsonMsg)
	if err := json.Unmarshal([]byte(jsonMsg), addMsg); err != nil {
		klog.Errorf("解析群成员增加事件 '%s' 失败: %s", jsonMsg, err)
		return err
	}

	// 统计所有进群人数、需要添加人数、需要踢出人数
	allUsers := make([]*database.User, 0)
	addUsers := make([]*database.User, 0)
	removeUsers := make([]*database.User, 0)
	for _, u := range addMsg.Guest {
		dataUser := database.User{
			GroupWxid:        addMsg.GroupWxid,
			NickName:         u.Nickname,
			Wxid:             u.Wxid,
			InviteUserNumber: 0,
			Alerted:          false,
			Role:             database.UserRoleNormal,
		}
		allUsers = append(allUsers, &dataUser)
		_, err := database.GetBlackListByWxid(u.Wxid)
		if err == nil {
			klog.Infof("用户%s(%s)在黑名单中", u.Nickname, u.Wxid)
			removeUsers = append(removeUsers, &dataUser)
		} else {
			addUsers = append(addUsers, &dataUser)
		}
	}

	// 获取邀请者信息
	inviter, err := database.GetGroupUserByWxid(addMsg.GroupWxid, addMsg.Inviter.Wxid)
	if err != nil {
		klog.Errorf("邀请者不在数据库中: %s(%s) '%s'", addMsg.Inviter.Nickname, addMsg.Inviter.Wxid, err)
		return err
	}

	// 管理员邀请的，不会踢掉，并且移除黑名单
	if inviter.Role == database.UserRoleOwner || inviter.Role == database.UserRoleManager {
		if err := database.DeleteBlackLists(removeUsers); err != nil {
			klog.Errorf("移出黑名单'%+v'失败: %s", removeUsers, err)
			return err
		}
		addUsers = append(addUsers, removeUsers...)
	} else {
		// 否则踢掉黑名单用户
		if len(removeUsers) > 0 {
			rmUsers := ""
			for _, u := range removeUsers {
				ctl.enqueueSendAction(utils.RemoveMsgSendParam(addMsg.GroupWxid, u.Wxid))
				rmUsers = path.Join(rmUsers, fmt.Sprintf("%s(%s)", u.NickName, u.Wxid))
			}
			msg := fmt.Sprintf("用户%s在黑名单中，立即踢出", rmUsers)
			ctl.enqueueSendMsg(utils.TextMsgSendParam(msg, addMsg.GroupWxid))
		}
	}

	// 添加数据库记录
	if err := database.CreateUsers(addUsers); err != nil {
		klog.Errorf("添加群成员'%+v'失败: %s", addUsers, err)
		return err
	}

	// 更新邀请者的邀请人数
	inviter.InviteUserNumber = inviter.InviteUserNumber + len(addUsers)
	if err := database.UpdateGroupUserByWxid(*inviter, addMsg.GroupWxid, inviter.Wxid); err != nil {
		klog.Errorf("群成员'%s | %s(%s)'更新邀请人数失败: %s", addMsg.GroupWxid, inviter.NickName, inviter.Wxid, err)
		return err
	}
	klog.Info("处理群成员增加事件成功！")
	if len(config.GlobalConfig.InviteMangerConf.WelcomeMsg) > 0 {
		ctl.enqueueSendMsg(utils.TextMsgSendParam(config.GlobalConfig.InviteMangerConf.WelcomeMsg, addMsg.GroupWxid))
	}
	return nil
}

func execGroupUserDec(jsonMsg string) error {
	decMsg := new(types.GroupUserDecJsonMsg)
	if err := json.Unmarshal([]byte(jsonMsg), decMsg); err != nil {
		klog.Errorf("解析群成员增加事件 '%s' 失败: %s", jsonMsg, err)
		return err
	}
	if err := database.DeleteGroupUserByWxid(decMsg.GroupWxid, decMsg.MemberWxid); err != nil {
		klog.Infof("删除群用户信息'%+v'失败: %s", decMsg, err)
		return err
	}
	klog.Info("处理群成员减少事件成功！")
	return nil
}

func (ctl *Controller) healthCheck(reqParam *types.RequestParam) error {
	msg := "机器人运行正常"

	_, err := wechat.GetLocalUserInfo(0)
	if err != nil {
		klog.Errorf("可爱猫健康检查失败: %s", err)
		msg = fmt.Sprintf("可爱猫健康检查失败: %s\n", err)
	}

	_, err = database.GetGroupUserByWxid(reqParam.FromWxid, reqParam.RobotWxid)
	if err != nil {
		klog.Errorf("数据库健康检查失败: %s", err)
		msg = fmt.Sprintf("数据库健康检查失败: %s\n", err)
	}

	sendP := types.SendParam{
		Api:       types.SendTextMsgApi,
		Msg:       msg,
		RobotWxid: reqParam.RobotWxid,
		ToWxid:    reqParam.FromWxid,
	}
	ctl.enqueueSendMsg(sendP)
	return nil
}

func (ctl *Controller) manageRemoveGroupUser(reqParam *types.RequestParam) error {
	result := ""
	reqUser, err := database.GetGroupUserByWxid(reqParam.FromWxid, reqParam.FinalFromWxid)
	if err != nil {
		result = fmt.Sprintf("查询失败: %s", err)
	} else if reqUser.Role != database.UserRoleOwner && reqUser.Role != database.UserRoleManager {
		result = fmt.Sprintf("您不是群管理员，不能执行此操作！")
	} else {
		atNickname, atWxid, _, _ := utils.IsAtMsg(reqParam.Msg)
		// 获取用户信息
		user, getErr := database.GetGroupUserByWxid(reqParam.FromWxid, atWxid)
		if getErr == gorm.ErrRecordNotFound {
			klog.Infof("用户%s(%s)不存在", atNickname, atWxid)
			user = &database.User{
				Wxid:      atWxid,
				GroupWxid: reqParam.FromWxid,
			}
		} else if getErr != nil {
			result = fmt.Sprintf("踢出用户失败, 获取用户信息错误：%s", getErr)
			ctl.enqueueSendMsg(utils.TextMsgSendParam(result, reqParam.FromWxid))
			return getErr
		}

		// 删除用户信息
		if getErr != gorm.ErrRecordNotFound {
			if err := database.DeleteGroupUserByWxid(reqParam.FromWxid, user.Wxid); err != nil {
				result = fmt.Sprintf("踢出用户失败, 删除用户信息错误：%s", err)
				ctl.enqueueSendMsg(utils.TextMsgSendParam(result, reqParam.FromWxid))
				return err
			}
		}

		// 添加黑名单
		if err := database.CreateBlackList(user); err != nil {
			result = fmt.Sprintf("踢出用户失败, 添加黑名单错误：%s", err)
			ctl.enqueueSendMsg(utils.TextMsgSendParam(result, reqParam.FromWxid))
			return err
		}

		// 踢掉用户
		ctl.enqueueSendAction(utils.RemoveMsgSendParam(reqParam.FromWxid, user.Wxid))
		result = fmt.Sprintf("踢出用户%s(%s)成功", user.NickName, user.Wxid)
	}
	ctl.enqueueSendMsg(utils.TextMsgSendParam(result, reqParam.FromWxid))
	return nil
}

func (ctl *Controller) setGroupUserManager(reqParam *types.RequestParam) error {
	result := ""
	reqUser, err := database.GetGroupUserByWxid(reqParam.FromWxid, reqParam.FinalFromWxid)
	if err != nil {
		result = fmt.Sprintf("查询失败: %s", err)
	} else if reqUser.Role != database.UserRoleOwner {
		result = fmt.Sprintf("您不是全局管理员，不能执行此操作！")
	} else {
		_, atWxid, _, _ := utils.IsAtMsg(reqParam.Msg)
		user, err := database.GetGroupUserByWxid(reqParam.FromWxid, atWxid)
		if err != nil {
			result = fmt.Sprintf("设置管理员失败：%s", err)
			ctl.enqueueSendMsg(utils.TextMsgSendParam(result, reqParam.FromWxid))
			return err
		}
		user.Role = database.UserRoleManager
		if err := database.UpdateGroupUserByWxid(*user, reqParam.FromWxid, user.Wxid); err != nil {
			result = fmt.Sprintf("设置管理员失败：%s", err)
			ctl.enqueueSendMsg(utils.TextMsgSendParam(result, reqParam.FromWxid))
			return err
		}
		result = fmt.Sprintf("设置管理员成功！")
	}
	ctl.enqueueSendMsg(utils.TextMsgSendParam(result, reqParam.FromWxid))
	return nil
}

func (ctl *Controller) removeGroupUserManager(reqParam *types.RequestParam) error {
	result := ""
	reqUser, err := database.GetGroupUserByWxid(reqParam.FromWxid, reqParam.FinalFromWxid)
	if err != nil {
		result = fmt.Sprintf("查询失败: %s", err)
	} else if reqUser.Role != database.UserRoleOwner {
		result = fmt.Sprintf("您不是全局管理员，不能执行此操作！")
	} else {
		_, atWxid, _, _ := utils.IsAtMsg(reqParam.Msg)
		user, err := database.GetGroupUserByWxid(reqParam.FromWxid, atWxid)
		if err != nil {
			result = fmt.Sprintf("取消管理员失败：%s", err)
			ctl.enqueueSendMsg(utils.TextMsgSendParam(result, reqParam.FromWxid))
			return err
		}
		user.Role = database.UserRoleNormal
		if err := database.UpdateGroupUserByWxid(*user, reqParam.FromWxid, user.Wxid); err != nil {
			result = fmt.Sprintf("取消管理员失败：%s", err)
			ctl.enqueueSendMsg(utils.TextMsgSendParam(result, reqParam.FromWxid))
			return err
		}
		result = fmt.Sprintf("取消管理员成功！")
	}
	ctl.enqueueSendMsg(utils.TextMsgSendParam(result, reqParam.FromWxid))
	return nil
}

func (ctl *Controller) setGroupUserVip(reqParam *types.RequestParam) error {
	result := ""
	reqUser, err := database.GetGroupUserByWxid(reqParam.FromWxid, reqParam.FinalFromWxid)
	if err != nil {
		result = fmt.Sprintf("查询失败: %s", err)
	} else if reqUser.Role != database.UserRoleOwner && reqUser.Role != database.UserRoleManager {
		result = fmt.Sprintf("您不是群管理员，不能执行此操作！")
	} else {
		_, atWxid, _, _ := utils.IsAtMsg(reqParam.Msg)
		user, err := database.GetGroupUserByWxid(reqParam.FromWxid, atWxid)
		if err != nil {
			result = fmt.Sprintf("设置白名单失败：%s", err)
			ctl.enqueueSendMsg(utils.TextMsgSendParam(result, reqParam.FromWxid))
			return err
		}
		user.Role = database.UserRoleVip
		if err := database.UpdateGroupUserByWxid(*user, reqParam.FromWxid, user.Wxid); err != nil {
			result = fmt.Sprintf("设置白名单失败：%s", err)
			ctl.enqueueSendMsg(utils.TextMsgSendParam(result, reqParam.FromWxid))
			return err
		}
		result = fmt.Sprintf("设置白名单成功！")
	}
	ctl.enqueueSendMsg(utils.TextMsgSendParam(result, reqParam.FromWxid))
	return nil
}

func (ctl *Controller) inviteNumCheck(reqParam *types.RequestParam) error {
	result := ""
	u, err := database.GetGroupUserByWxid(reqParam.FromWxid, reqParam.FinalFromWxid)
	if err != nil {
		result = fmt.Sprintf("查询失败: %s", err)
	} else {
		result = fmt.Sprintf("您的当前积分为: %d", u.InviteUserNumber)
	}
	ctl.enqueueSendMsg(utils.AlertAtMsgSendParam(reqParam.FromWxid, reqParam.FinalFromWxid, reqParam.FinalFromName, result))
	return err
}

func (ctl *Controller) SetUserInviteNum(reqParam *types.RequestParam) error {
	result := ""
	reqUser, err := database.GetGroupUserByWxid(reqParam.FromWxid, reqParam.FinalFromWxid)
	if err != nil {
		result = fmt.Sprintf("设置积分失败：%s", err)
	} else if reqUser.Role != database.UserRoleOwner && reqUser.Role != database.UserRoleManager {
		result = fmt.Sprintf("您不是群管理员，不能执行此操作！")
	} else {
		_, atWxid, _, _ := utils.IsAtMsg(reqParam.Msg)
		user, err := database.GetGroupUserByWxid(reqParam.FromWxid, atWxid)
		if err != nil {
			result = fmt.Sprintf("设置积分失败：%s", err)
			ctl.enqueueSendMsg(utils.TextMsgSendParam(result, reqParam.FromWxid))
			return err
		}
		user.InviteUserNumber++
		if err := database.UpdateGroupUserByWxid(*user, reqParam.FromWxid, user.Wxid); err != nil {
			result = fmt.Sprintf("设置积分失败：%s", err)
			ctl.enqueueSendMsg(utils.TextMsgSendParam(result, reqParam.FromWxid))
			return err
		}
		result = fmt.Sprintf("设置积分成功，积分 +1")
	}
	ctl.enqueueSendMsg(utils.TextMsgSendParam(result, reqParam.FromWxid))
	return nil
}

func (ctl *Controller) execGroupCreateTaolijin(reqParam *types.RequestParam) error {
	// 检查是否是监听的群组
	if !isFromListionGroup(reqParam) {
		klog.Infof("群组未监听, 跳过")
		return nil
	}

	var tkl string
	var err error

	// 检查是否是商品ID
	// 检查是否是商品ID,淘礼金价格,淘礼金数量,时间
	inputs := strings.Split(reqParam.Msg, ",")
	if len(strings.TrimSpace(reqParam.Msg)) == 12 {
		tkl, err = ctl.execCreateTaolijinWithGoodsID(reqParam.Msg, reqParam.FromWxid)
	} else if len(inputs) == 4 {
		tkl, err = ctl.execCreateTaolijin(inputs[0], inputs[1], inputs[2], inputs[3])
	} else {
		err = fmt.Errorf("无法识别内容: '%s'", reqParam.Msg)
	}

	// 发送结果
	var msg string
	if err != nil {
		msg = err.Error()
	} else {
		msg = fmt.Sprintf(front.Ct.Keywords, tkl)
	}
	ctl.enqueueSendMsg(utils.TextMsgSendParam(msg, reqParam.FromWxid))
	return nil
}

func (ctl *Controller) execCreateTaolijinWithGoodsID(goodsID string, FromWxid string) (string, error) {
	// TODO: 实现创建淘礼金的淘口令内容
	return "", nil
}

func (ctl *Controller) execCreateTaolijin(goodsID string, perFace string, totalNum string, day string) (string, error) {
	// TODO: 实现创建淘礼金的淘口令内容
	taolijin, err := TaobaoClient.CreateTaoLiJinUrl(goodsID, perFace, totalNum, day)
	if err != nil {
		klog.Error(err)
		return "", err
	}
	tkl, err := TaobaoClient.CreateTaoKouLing("", taolijin)
	if err != nil {
		klog.Error(err)
		return "", err
	}
	klog.Infof("创建淘礼金成功: %s '%s'", goodsID, tkl)
	return tkl, nil
}

func isFromListionGroup(reqParam *types.RequestParam) bool {
	for _, g := range front.ListenGroups {
		if g.Wxid == reqParam.FromWxid {
			return true
		}
	}
	return false
}
