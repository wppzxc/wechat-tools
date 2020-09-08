package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
)

// AutoRemoveStatus 是否启用自动通过好友请求
var AutoRemoveStatus = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Namespace: "auto_remove",
		Name:      "auto_remove_status",
		Help:      "是否启用自动拉黑",
	},
)

// AutoRemoveUsers 自动踢出用户数量
var AutoRemoveUsers = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "auto_remove",
		Name:      "auto_remove_users",
		Help:      "自动踢出用户数量",
	},
)

// AutoBlackUsers 自动拉黑用户数量
var AutoBlackUsers = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "auto_remove",
		Name:      "auto_black_users",
		Help:      "自动拉黑用户数量",
	},
)