package front

import (
	"fmt"
)

// ValidateConfig 验证参数是否正确
func (sr *SendReceiver) ValidateConfig() error {

	if len(Ct.TaoBaoApiKey) == 0 {
		return fmt.Errorf("必须指定淘宝API KEY！")
	}
	if len(Ct.TaoBaoApiSecret) == 0 {
		return fmt.Errorf("必须指定淘宝API SECRET")
	}
	if len(Ct.TaoBaoAdZoneID) == 0 {
		return fmt.Errorf("必须指定淘宝 AdzoneID！")
	}
	if len(ListenGroups) == 0 {
		return fmt.Errorf("必须指定监听群组！")
	}

	// keywords := strings.Split(sr.Keywords, "/")
	// filterKeywords := strings.Split(sr.FilterKeywords, "/")

	// config.GlobalConfig.SendReceiveConf.Keywords = keywords
	// config.GlobalConfig.SendReceiveConf.FilterKeywords = filterKeywords
	// config.GlobalConfig.SendReceiveConf.TranMoneySep = sr.TranMoneySep
	// config.GlobalConfig.SendReceiveConf.StartSendReceiver = sr.StartSendReceiver
	// config.GlobalConfig.SendReceiveConf.SendInterval = sr.SendInterval
	// config.GlobalConfig.SendReceiveConf.ActionInterval = sr.ActionInterval

	// if err := config.SaveConfig(""); err != nil {
	// 	klog.Error(err)
	// }

	return nil
}
