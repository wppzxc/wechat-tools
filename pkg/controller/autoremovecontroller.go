package controller

import (
	"github.com/wppzxc/wechat-tools/pkg/config"
	"github.com/wppzxc/wechat-tools/pkg/types"
	"github.com/wppzxc/wechat-tools/pkg/utils"
	"k8s.io/klog"
	"strings"
	"unicode/utf8"
)

func (ctl *Controller) judgeMsgKickOut(reqParam *types.RequestParam) bool {
	if !config.GlobalConfig.AutoRemoveConf.Start {
		klog.Info("未启用自动踢人程序，跳过")
		return false
	}

	if config.GlobalConfig.AutoRemoveConf.Applets {
		if reqParam.Type == types.AppletsMsg {
			klog.Warning("违规发言：发送小程序！")
			return true
		}
	}

	if config.GlobalConfig.AutoRemoveConf.FilterNames && len(config.GlobalConfig.AutoRemoveConf.FilterNamesString) > 0 {
		filterNames := strings.Split(config.GlobalConfig.AutoRemoveConf.FilterNamesString, "/")
		if contain := utils.StringsIndex(filterNames, reqParam.FinalFromName); contain {
			klog.Warning("用户名称违规：包含过滤词！")
			return true
		}
	}

	if config.GlobalConfig.AutoRemoveConf.FilterWords && len(config.GlobalConfig.AutoRemoveConf.FilterWordsString) > 0 {
		filterWords := strings.Split(config.GlobalConfig.AutoRemoveConf.FilterWordsString, "/")
		if contain := utils.StringsIndex(filterWords, reqParam.Msg); contain {
			klog.Warning("发言违规：包含过滤词！")
			return true
		}
	}

	if config.GlobalConfig.AutoRemoveConf.MsgLength {
		count := utf8.RuneCountInString(reqParam.Msg)
		if reqParam.Type == types.TextMsg && count > config.GlobalConfig.AutoRemoveConf.MaxMsgLength {
			klog.Warningf("违规发言：发送文字数量%d大于设定值%d！", count, config.GlobalConfig.AutoRemoveConf.MaxMsgLength)
			return true
		}
	}

	if config.GlobalConfig.AutoRemoveConf.SendCard {
		if reqParam.Type == types.CardMsg {
			klog.Warning("违规发言：发送名片！")
			return true
		}
	}

	if config.GlobalConfig.AutoRemoveConf.SendLink {
		if strings.Index(reqParam.Msg, "https//") >= 0 || strings.Index(reqParam.Msg, "http") >= 0 {
			klog.Warning("违规发言：发送内容包含链接！")
			return true
		}
	}

	if config.GlobalConfig.AutoRemoveConf.SendQRCode {
		if reqParam.Type == types.ImageMsg {
			if ok := utils.CheckQeCode(reqParam.Msg); ok {
				klog.Warning("违规发言：发送二维码！")
				return true
			}
		}
	}

	if config.GlobalConfig.AutoRemoveConf.SendVideo {
		if reqParam.Type == types.VideoMsg {
			klog.Warning("违规发言：发送视频！")
			return true
		}
	}

	if config.GlobalConfig.AutoRemoveConf.SendVoice {
		if reqParam.Type == types.VoiceMsg {
			klog.Warning("违规发言：发送语音！")
			return true
		}
	}

	if config.GlobalConfig.AutoRemoveConf.ShareLink {
		if reqParam.Type == types.ShareLinkMsg {
			klog.Warning("违规发言：发送分享链接！")
			return true
		}
	}
	return false
}
