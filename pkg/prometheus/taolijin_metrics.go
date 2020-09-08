package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
)

// TaoLiJinStatus 是否启用淘礼金
var TaoLiJinStatus = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Namespace: "taolijin",
		Name:      "taolijin_status",
		Help:      "是否启用淘礼金",
	},
)

// CreatedSuccessTaoLiJinCounts 创建成功的淘礼金数量
var CreatedSuccessTaoLiJinCounts = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "taolijin",
		Name:      "created_success_taolijin_counts",
		Help:      "创建成功的淘礼金数量",
	},
)

// CreatedFailedTaoLiJinCounts 创建失败的淘礼金数量
var CreatedFailedTaoLiJinCounts = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "taolijin",
		Name:      "created_failed_taolijin_counts",
		Help:      "创建失败的淘礼金数量",
	},
)