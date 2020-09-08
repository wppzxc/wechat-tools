package controller

import (
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/wppzxc/wechat-tools/pkg/config"
	"github.com/wppzxc/wechat-tools/pkg/prometheus"
	"github.com/wppzxc/wechat-tools/pkg/types"
	"github.com/wppzxc/wechat-tools/pkg/utils"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog"
)

const (
	defaultAlertMsg = `%s
口令: %s
拍下: %.2f 元`
)

type dataokeItemSlice []types.DaTaoKeItem

func (d dataokeItemSlice) Len() int           { return len(d) }
func (d dataokeItemSlice) Swap(i, j int)      { d[i], d[j] = d[j], d[i] }
func (d dataokeItemSlice) Less(i, j int) bool { return d[i].ActualPrice < d[j].ActualPrice }

// StartTaoLiJinWorker 开始淘礼金任务
func (ctl *Controller) StartTaoLiJinWorker(stopCh chan struct{}) {
	// 创建临时目录
	err := os.MkdirAll("./tmp/image", 0755)
	if err != nil {
		klog.Fatal(err)
	}
	klog.Info("临时目录'./tmp/image'创建完成")
	klog.Info("启动淘礼金任务")
	wait.Until(ctl.runTaoLijinworker, time.Duration(config.GlobalConfig.TaoLiJinConf.Interval)*time.Second, stopCh)
	klog.Info("结束淘礼金任务")
}

func (ctl *Controller) runTaoLijinworker() {
	ctl.dtkClient = utils.NewDataokeClient()
	ctl.tbClient = utils.NewTaoBaoClient()

	// 计时器，当淘礼金任务运行24小时，清空发送记录一次
	hour := time.Now().Hour()
	if hour > ctl.taolijinNowHour {
		ctl.taolijinNowHour = hour
		ctl.taolijinRunHours++
	}
	if hour == 0 && ctl.taolijinNowHour != 0 {
		ctl.taolijinNowHour = 0
		ctl.taolijinRunHours++
	}

	// 每天清空一次发送记录
	if ctl.taolijinRunHours == 24 {
		klog.Info("清空淘礼金发送记录...")
		ctl.taolijinRunHours = 0
		ctl.taolijinSendItems = make(map[string]types.DaTaoKeItem, 0)
	}

	items, err := ctl.dtkClient.GetRealTimeListItem()
	if err != nil {
		klog.Error(err)
		return
	}
	if len(items) == 0 {
		klog.Error("拉取大淘客实时榜单数据失败，拉取数据为 0")
		return
	}
	newItems := make(dataokeItemSlice, 0)
	for _, i := range items {
		if i.NewRankingGoods > 0 {
			if _, ok := ctl.taolijinSendItems[i.GoodsId]; !ok {
				newItems = append(newItems, i)
			} else {
				klog.Info("商品今天已发送，跳过")
			}
		}
	}
	if len(newItems) == 0 {
		klog.Info("没有新上榜商品")
		return
	}
	klog.Infof("获取新上榜商品: %v", newItems)
	sort.Sort(newItems)

	item, perFace, tkl := func(items dataokeItemSlice) (*types.DaTaoKeItem, float32, string) {
		for _, i := range items {
			perFace := i.ActualPrice * (i.CommissionRate / 100) * (float32(config.GlobalConfig.TaoLiJinConf.TBPerFaceRate) / 100)
			perFaceStr := fmt.Sprintf("%.2f", perFace)
			tljURL, err := ctl.tbClient.CreateTaoLiJinUrl(i.GoodsId, perFaceStr, "")
			if err != nil {
				klog.Warningf("创建淘礼金失败：%s", err)
				continue
			}
			tkl, err := ctl.tbClient.CreateTaoKouLing(i.DTitle, tljURL)
			if err != nil {
				klog.Warningf("创建口令失败：%s", err)
				continue
			}
			return &i, perFace, tkl
		}
		return nil, 0, ""
	}(newItems)

	if len(tkl) == 0 {
		klog.Error("获取淘礼金失败！")
		prometheus.CreatedFailedTaoLiJinCounts.Inc()
		return
	}
	prometheus.CreatedSuccessTaoLiJinCounts.Inc()

	alert := fmt.Sprintf(defaultAlertMsg, item.DTitle, utils.TranMoneySep(tkl), item.ActualPrice-perFace)
	imgPath, err := utils.DownloadGoodsImage(item)
	if err != nil {
		klog.Warningf("下载商品图片失败: %s", err)
	}
	for _, g := range config.GlobalConfig.InviteMangerConf.ManageGroups {
		// 发送淘口令
		ctl.enqueueSendMsg(utils.TextMsgSendParam(alert, g.Wxid))
		// 发送对应图片
		if len(imgPath) > 0 {
			ctl.enqueueSendAction(utils.ImageMsgSendParam(imgPath, g.Wxid))
		}
	}
	ctl.taolijinSendItems[item.GoodsId] = *item
}
