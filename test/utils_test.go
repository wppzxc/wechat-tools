package test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/wppzxc/wechat-tools/pkg/types"
	"github.com/wppzxc/wechat-tools/pkg/utils"
)

func TestCheckQRCode(t *testing.T) {
	filepath := "C:\\Users\\wpp\\Desktop\\微信图片_20200601221124.jpg"
	if ok := utils.CheckQeCode(filepath); ok {
		fmt.Println("二维码图片!")
	} else {
		fmt.Println("不是二维码图片")
	}
}

func TestJsonStr(t *testing.T) {
	jsonMsg := `{"to_wxid":"wxid_lluoefrhpwlc22","to_name":"小毛","msgid":1093340227,"from_wxid":"wxid_monydd6sszyg22","from_nickname":"孔令北","v1":"v1_2531bb53db3fdd63148f2756edf01792845a0e00ff2f72411322673dd35b1c174b846eff933dd062132be99db1258f5b@stranger","v2":"v4_000b708f0b040000010000000000e4fcd37ba816275f794f2ab5f05e1000000050ded0b020927e3c97896a09d47e6e9e936974c2fe88d013a4c8b2c5a418b388b8ee11b046a0baf0d487f6ac248b9a5e816e990119ca5308237979536cfcfefffdf7a2c8098e4b2f1ff89c304e60e515ecb0c3f57dc555e9dc0e4c9e5d3f8062b939bb52c3a15bf27c264509424ac32b19c207ceb1ffb35901@stranger","sex":1,"from_content":"我是孔令北，办执照","headimgurl":"http:\/\/wx.qlogo.cn\/mmhead\/ver_1\/AbznYoVpaq27ic2FDibhxogAA6zk8zk26ENLJ97dgWmic1uSH2UfVOSn6ic66JsfTaJDIbyQ0lF0dj5F9hEqpznvgg\/96","type":30}`
	str := strings.Replace(jsonMsg, "\"", "\\\"", -1)
	fmt.Printf(str)
}

func TestPWD(t *testing.T) {
	dir, _ := os.Getwd()
	imgPath := fmt.Sprintf("%s\\invite.jpg", dir)
	fmt.Println(imgPath)
}

func TestDownloadImg(t *testing.T) {
	err := os.MkdirAll("./tmp/image", 0755)
	if err != nil {
		panic(err)
	}
	item := new(types.DaTaoKeItem)
	item.MainPic = "https://img.alicdn.com/imgextra/i4/2200542991284/O1CN01HuS8k01LM78uLxhzo_!!2200542991284.jpg"
	imgPath, err := utils.DownloadGoodsImage(item)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(imgPath)
}

func TestTransMoney(t *testing.T) {
	str := `测试一二三c8fJ1D2f4it)`
	newStr := utils.TranMoneySep(str)
	fmt.Println(newStr)
}
