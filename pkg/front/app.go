package front

import (
	"fmt"
	"github.com/wppzxc/wechat-tools/pkg/config"
	"k8s.io/klog"
	"strings"
)

// ValidateConfig 验证参数是否正确
func (sr *SendReceiver) ValidateConfig() error {
	if sr.StartSendReceiver {
		if len(config.GlobalConfig.SendReceiveConf.ReceiveFromGroup) == 0 {
			return fmt.Errorf("必须指定监听群组！")
		}
		if len(config.GlobalConfig.SendReceiveConf.SendToUsers) == 0 {
			return fmt.Errorf("必须指定转发用户！")
		}

		// 提醒时间必须是 0点 ~ 23点，且开始时间不能大于结束时间
		if config.GlobalConfig.InviteMangerConf.AlertTimeBegin > config.GlobalConfig.InviteMangerConf.AlertTimeEnd {
			return fmt.Errorf("群邀请设置：提醒时间不正确，请重新输入")
		}
	}

	keywords := strings.Split(sr.Keywords, "/")
	filterKeywords := strings.Split(sr.FilterKeywords, "/")

	config.GlobalConfig.SendReceiveConf.Keywords = keywords
	config.GlobalConfig.SendReceiveConf.FilterKeywords = filterKeywords
	config.GlobalConfig.SendReceiveConf.TranMoneySep = sr.TranMoneySep
	config.GlobalConfig.SendReceiveConf.StartSendReceiver = sr.StartSendReceiver
	config.GlobalConfig.SendReceiveConf.SendInterval = sr.SendInterval
	config.GlobalConfig.SendReceiveConf.ActionInterval = sr.ActionInterval

	if err := config.SaveConfig(""); err != nil {
		klog.Error(err)
	}

	return nil
}
