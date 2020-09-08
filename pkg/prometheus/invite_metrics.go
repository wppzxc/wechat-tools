package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
)

// InviteManagerStatus 是否启用强制邀请
var InviteManagerStatus = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Namespace: "invite_manager",
		Name:      "invite_manager_status",
		Help:      "是否启用强制邀请",
	},
)

// AlertUsers 警告数量
var AlertUsers = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "invite_manager",
		Name:      "alert_users",
		Help:      "提醒的用户数量",
	},
)

// RemoveUsers 踢出用户数量
var RemoveUsers = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "invite_manager",
		Name:      "remove_users",
		Help:      "踢出的用户数量",
	},
)

