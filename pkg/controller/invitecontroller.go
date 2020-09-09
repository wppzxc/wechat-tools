package controller

import (
	"fmt"
	"time"

	"github.com/wppzxc/wechat-tools/pkg/config"
	"github.com/wppzxc/wechat-tools/pkg/database"
	"github.com/wppzxc/wechat-tools/pkg/prometheus"
	"github.com/wppzxc/wechat-tools/pkg/utils"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog"
)

const (
	defaultTimeUnit = time.Hour
)

// StartInviteManger 开始群裂变任务
func (ctl *Controller) StartInviteManger(stopCh chan struct{}) {
	if config.GlobalConfig.InviteMangerConf.AlertHours == 0 || config.GlobalConfig.InviteMangerConf.RemoveHours == 0 {
		klog.Info("群裂变模块：提醒时间或踢除时间为0，不启动群裂变功能")
		return
	}
	klog.Info("开始强制邀请任务")
	wait.Until(ctl.startInviteManagerWork, 1*defaultTimeUnit, stopCh)
	klog.Info("停止强制邀请任务")
}

func (ctl *Controller) startInviteManagerWork() {
	klog.Info("执行强制邀请...")
	hour := time.Now().Hour()

	// 判断是否在工作时间
	if hour >= config.GlobalConfig.InviteMangerConf.AlertTimeBegin && hour <= config.GlobalConfig.InviteMangerConf.AlertTimeEnd {
		for _, group := range config.GlobalConfig.InviteMangerConf.ManageGroups {

			users, err := database.GetGroupUsersByRole(group.Wxid, database.UserRoleNormal)
			if err != nil {
				errMsg := fmt.Sprintf("警告：强制邀请任务：获取群 %s 全部用户失败: %s", group.Name, err)
				klog.Error(errMsg)
				// ctl.enqueueSendMsg(utils.TextMsgSendParam(errMsg, group.Wxid))
			}

			// 移除用户
			removeUser := ctl.getNeedRemoveUser(users)
			removeMsg := ""
			if removeUser != nil {
				// 删除数据库用户信息
				if err := database.DeleteGroupUserByWxid(group.Wxid, removeUser.Wxid); err != nil {
					klog.Errorf("用户'%+v'移除数据库失败: err", removeUser, err)
				}

				// 生成消息
				removeMsg = fmt.Sprintf("@%s 在规定时间内邀请不足3人已被移除群聊！", removeUser.NickName)
				// 发送消息
				ctl.enqueueSendMsg(utils.TextMsgSendParam(removeMsg, group.Wxid))
				// 移除出群组
				ctl.enqueueSendAction(utils.RemoveMsgSendParam(group.Wxid, removeUser.Wxid))
				prometheus.RemoveUsers.Inc()
			} else {
				klog.Warningf("群组%s未找到需要踢出的用户", group.Name)
			}

			// 提醒用户
			alertUser := ctl.getNeedAlertUser(users)
			alertMsg := ""
			if alertUser != nil {
				alertMsg = fmt.Sprintf(" 邀请不足3人即将移除群聊")
				alertUser.Alerted = true
				// 更新数据库已提醒状态
				if err := database.UpdateGroupUserByWxid(*alertUser, group.Wxid, alertUser.Wxid); err != nil {
					klog.Errorf("更新用户%s(%s)提醒状态失败: %s", alertUser.NickName, alertUser.Wxid, err)
				}

				// 发送提醒信息
				ctl.enqueueSendMsg(utils.AlertAtMsgSendParam(group.Wxid, alertUser.Wxid, alertUser.NickName, alertMsg))
				prometheus.AlertUsers.Inc()
			} else {
				klog.Warningf("群组%s未找到需要提醒的用户", group.Name)
			}
		}
	} else {
		klog.Info("当前时间不在设定的工作时间内，跳过...")
	}
}

// getNeedRemoveUser 判断该群组中的用户是否超过踢出时间
func (ctl *Controller) getNeedRemoveUser(users []database.User) *database.User {
	for _, u := range users {
		if _, err := database.GetWhiteListByWxid(u.Wxid); err == nil {
			klog.V(6).Infof("跳过白名单用户%s(%s)", u.NickName, u.Wxid)
			continue
		}
		// 如果未警告过，则跳过
		if !u.Alerted {
			continue
		}
		// 如果是活跃用户，则跳过
		if u.Active {
			continue
		}
		if u.InviteUserNumber < config.DefaultInviteNum {
			// 超过踢人时间，踢掉
			if u.CreatedAt.Add(time.Duration(config.GlobalConfig.InviteMangerConf.AlertHours+
				config.GlobalConfig.InviteMangerConf.RemoveHours) * defaultTimeUnit).Before(time.Now()) {
				return &u
			}
		}
	}

	return nil
}

// getNeedAlertUser 判断该群组中的用户是否超过警告时间
func (ctl *Controller) getNeedAlertUser(users []database.User) *database.User {
	for _, u := range users {
		if _, err := database.GetWhiteListByWxid(u.Wxid); err == nil {
			klog.V(6).Infof("跳过白名单用户:%s(%s)", u.NickName, u.Wxid)
			continue
		}
		// 如果是活跃用户，则跳过
		if u.Active {
			continue
		}
		if u.InviteUserNumber < config.DefaultInviteNum {
			// 超过警告时间，且没警告过，警告他
			if u.CreatedAt.Add(time.Duration(config.GlobalConfig.InviteMangerConf.AlertHours)*defaultTimeUnit).Before(time.Now()) && !u.Alerted {
				return &u
			}
		}
	}

	return nil
}
