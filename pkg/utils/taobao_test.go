package utils

import (
	"fmt"
	"github.com/wppzxc/wechat-tools/pkg/config"
	"k8s.io/klog"
	"testing"
)

func TestCreateTaoLiJin(t *testing.T) {
	config.GlobalConfig = new(config.Config)
	config.GlobalConfig.TaoLiJinConf = new(config.TaoLiJinConf)
	config.GlobalConfig.TaoLiJinConf.TBAppKey = "30029015"
	config.GlobalConfig.TaoLiJinConf.TBAppSecret = "b9f71d954e6a331a160c6c33956f7c44"
	config.GlobalConfig.TaoLiJinConf.TBAdzoneID = "110409800054"
	config.GlobalConfig.TaoLiJinConf.TBTotalNum = "2"

	itemId := "610742526459"
	tbc := NewTaoBaoClient()
	sendUrl, err := tbc.CreateTaoLiJinUrl(itemId, "1", "1", "")
	if err != nil {
		fmt.Println("Failed")
		fmt.Println(err)
	} else {
		fmt.Println("OK")
		fmt.Println(sendUrl)
	}
}

func TestCreateTaoKouLing(t *testing.T) {
	config.GlobalConfig = new(config.Config)
	config.GlobalConfig.TaoLiJinConf = new(config.TaoLiJinConf)
	config.GlobalConfig.TaoLiJinConf.TBAppKey = "30029015"
	config.GlobalConfig.TaoLiJinConf.TBAppSecret = "b9f71d954e6a331a160c6c33956f7c44"
	config.GlobalConfig.TaoLiJinConf.TBAdzoneID = "110409800054"
	config.GlobalConfig.TaoLiJinConf.TBTotalNum = "2"

	itemId := "610742526459"
	tbc := NewTaoBaoClient()
	sendUrl, err := tbc.CreateTaoLiJinUrl(itemId, "1", "", "")
	if err != nil {
		fmt.Printf("Error : %s\n", err)
		return
	}

	taokouling, err := tbc.CreateTaoKouLing("一二三", sendUrl)
	if err != nil {
		fmt.Printf("Error : %s\n", err)
	}
	fmt.Printf("get taokouling : %s", taokouling)
}

func TestGetRealTimeListItem(t *testing.T) {
	config.GlobalConfig = new(config.Config)
	config.GlobalConfig.TaoLiJinConf = new(config.TaoLiJinConf)
	config.GlobalConfig.TaoLiJinConf.DTKAppKey = "5e9d2dbadc286"
	config.GlobalConfig.TaoLiJinConf.DTKAppSecret = "8f3c81484fdf7bd2695ddbbc6a128201"

	dtkClient := NewDataokeClient()
	data, err := dtkClient.GetRealTimeListItem()
	if err != nil {
		klog.Error(err)
		return
	}
	fmt.Println(data)
}
