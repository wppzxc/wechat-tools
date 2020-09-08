package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
)

// AutoAgreeStatus 是否启用自动通过好友请求
var AutoAgreeStatus = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Namespace: "auto_agree",
		Name:      "auto_agree_status",
		Help:      "是否启用自动通过好友请求",
	},
)

// AgreeSuccessUsers 自动通过成功的用户数量
var AgreeSuccessUsers = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "auto_agree",
		Name:      "agree_success_users",
		Help:      "自动通过成功的用户数量",
	},
)

// AgreeFailedUsers 自动通过失败的用户数量
var AgreeFailedUsers = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "auto_agree",
		Name:      "agree_failed_users",
		Help:      "自动通过成功的用户数量",
	},
)

// SendWelcomeTexts 发送的欢迎消息
var SendWelcomeTexts = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "auto_agree",
		Name:      "send_welcome_texts",
		Help:      "发送的欢迎消息数量",
	},
)

// SendWelcomeImages 发送的欢迎图片
var SendWelcomeImages = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "auto_agree",
		Name:      "send_welcome_images",
		Help:      "发送的欢迎图片数量",
	},
)