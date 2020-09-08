package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
)

// WechatStatus 微信运行状态
var WechatStatus = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Namespace: "base",
		Name:      "wechat_status",
		Help:      "微信运行状态",
	},
)

// KeaimaoStatus 可爱猫运行状态
var KeaimaoStatus = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Namespace: "base",
		Name:      "keaimao_status",
		Help:      "可爱猫运行状态",
	},
)

// WechatToolsStatus wecaht-tools 运行状态
var WechatToolsStatus = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Namespace: "base",
		Name:      "wechat_tools_status",
		Help:      "wecaht_tools 运行状态",
	},
)

// InitMetrics init custom metrics
func InitMetrics() {
	// init base metrics
	prometheus.MustRegister(WechatStatus)
	prometheus.MustRegister(KeaimaoStatus)
	prometheus.MustRegister(WechatToolsStatus)

	// init sendReceive metrics
	prometheus.MustRegister(SendReceiveStatus)
	prometheus.MustRegister(ReceiveMsgs)
	prometheus.MustRegister(ExecuteMsgs)
	prometheus.MustRegister(TransTexts)
	prometheus.MustRegister(TransImages)
	prometheus.MustRegister(TransVideos)
	prometheus.MustRegister(TransFailedMsgs)

	// init inviteManager metrics
	prometheus.MustRegister(InviteManagerStatus)
	prometheus.MustRegister(AlertUsers)
	prometheus.MustRegister(RemoveUsers)

	// init autoAgreeFriend metrics
	prometheus.MustRegister(AutoAgreeStatus)
	prometheus.MustRegister(AgreeSuccessUsers)
	prometheus.MustRegister(AgreeFailedUsers)
	prometheus.MustRegister(SendWelcomeTexts)
	prometheus.MustRegister(SendWelcomeImages)

	// init taoLiJin metrics
	prometheus.MustRegister(TaoLiJinStatus)
	prometheus.MustRegister(CreatedSuccessTaoLiJinCounts)
	prometheus.MustRegister(CreatedFailedTaoLiJinCounts)

	// init autoRemover metrics
	prometheus.MustRegister(AutoRemoveStatus)
	prometheus.MustRegister(AutoRemoveUsers)
	prometheus.MustRegister(AutoBlackUsers)
}