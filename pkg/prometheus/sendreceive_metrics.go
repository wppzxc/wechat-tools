package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
)

// SendReceiveStatus 是否启用转发
var SendReceiveStatus = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Namespace: "sendreceive",
		Name:      "send_receive_status",
		Help:      "是否启用转发",
	},
)

// ReceiveMsgs 接收的消息
var ReceiveMsgs = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "sendreceive",
		Name:      "receive_msgs",
		Help:      "接收的总消息数量",
	},
)

// ExecuteMsgs 处理的消息
var ExecuteMsgs = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "sendreceive",
		Name:      "execute_msgs",
		Help:      "处理的消息数量",
	},
)

// TransTexts 发送的文字消息
var TransTexts = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "sendreceive",
		Name:      "trans_texts",
		Help:      "发送的文字消息数量",
	},
)

// TransImages 发送的图片消息
var TransImages = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "sendreceive",
		Name:      "trans_images",
		Help:      "发送的图片消息数量",
	},
)

// TransVideos 发送的视频消息
var TransVideos = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "sendreceive",
		Name:      "trans_videos",
		Help:      "发送的视频消息数量",
	},
)

// TransFailedMsgs 发送失败的消息
var TransFailedMsgs = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "sendreceive",
		Name:      "trans_failed_msgs",
		Help:      "发送的视频消息数量",
	},
)
