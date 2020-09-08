package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/tealeg/xlsx"
)

// MeiGuangData 接口返回数据/excel行数据
type MeiGuangData struct {
	AlipayName               string
	BefID                    int64
	BefMobile                string
	CalcModeType             int
	CalcModeTypeName         string
	CreateTime               string
	ID                       int64
	InviteCode               string
	Mobile                   string
	ModeType                 int
	ModeTypeName             string
	MonthRate                int64
	MonthSelfBuyCommission   int64
	NextUserCount            int
	Pid                      string
	RelationID               int64
	RemainAmount             int64
	Remark                   string
	StrUserType              string
	ThirtyRate               int64
	ThirtySelfBuyCommission  int64
	TkID                     int64
	TodayOrderCount          int64
	TodayRate                int64
	TodaySelfBuyCommission   int64
	UPMonthBalance           int64
	UPMonthSelfBuyCommission int64
	UserAvator               string
	UserToken                string
	WeiXin                   string
	BefShopManager           string
}

// BefData 查询店长信息返回数据
type BefData struct {
	WxNickName string
	Mobile string
	AlipayName string
	Remark string
}

var client *http.Client = new(http.Client)
var zutuanweb string

func init() {
	fmt.Printf("请输入 zutuanweb 值并按回车...\n:")
	fmt.Scanln(&zutuanweb)
	if len(zutuanweb) == 0 {
		fmt.Println("未输入 zutuanweb, 将使用默认值")
		zutuanweb = "n1/qdkTaC1J5a1/YU7bdSKHYMC2ePc/pQWP6OAvX8Mq0/AxmYCVc4IOrmv2m5SU6r/fCOSZmyUGQ7AF894N4Ss24pmwV1XvMUV52SZ7JruBEYXlAjYH8qgznwZZphxAZjTfSHim/vvYmzY4JuvsDMo%2BkS6COGetSEOG%2BOm32hu1dkO/N7IZqkstm4ovFGcLmSRqP2V4i6dSCEpS/wfYTXa/nN2RTeyj%2BHRrOwR5jI9Tyo1VcvplY6M9tcCcQUBO2PvVjofFrKDM0XEzQVYGJ48dZHsrT0HLIGySGmn1u2QUSIrnYzyr%2BW21r2iVSfts5C0co/RaObRxwhWgIhBxXmg%3D%3D"
	}
}

// GetMeiguangData 获取指定页码的数据
func GetMeiguangData(pageNum int) ([]MeiGuangData, error) {
	url := fmt.Sprintf("http://subadmin.sitezt.cn/api/p/tk/getuserlist?page=%d&useridentity=-1", pageNum)
	var req *http.Request
	req, _ = http.NewRequest(http.MethodGet, url, nil)
	cookie1 := &http.Cookie{
		Domain:   "subadmin.sitezt.cn",
		Path:     "/",
		Name:     "_zutuanweb",
		Value:    zutuanweb,
		HttpOnly: true}
	req.AddCookie(cookie1)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, _ := ioutil.ReadAll(resp.Body)

	datas := make([]MeiGuangData, 0)
	if err := json.Unmarshal(data, &datas); err != nil {
		return nil, err
	}
	return datas, nil
}

// GetBefShopManager 查询上级店长
func GetBefShopManager(befID int64) (*BefData, error) {
	url := fmt.Sprintf("http://subadmin.sitezt.cn/api/p/tk/getbefuser_dz?userId=%d", befID)
	var req *http.Request
	req, _ = http.NewRequest(http.MethodGet, url, nil)
	cookie1 := &http.Cookie{
		Domain:   "subadmin.sitezt.cn",
		Path:     "/",
		Name:     "_zutuanweb",
		Value:    zutuanweb,
		HttpOnly: true}
	req.AddCookie(cookie1)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, _ := ioutil.ReadAll(resp.Body)

	info := new(BefData)
	if err := json.Unmarshal(data, info); err != nil {
		return nil, err
	}
	return info, nil
}

func run() error {
	AllDatas := make([]MeiGuangData, 0)
	timestamp := time.Now().Unix()
	filename := "美逛后台数据-" + strconv.FormatInt(timestamp, 10) + ".xlsx"
	var file *xlsx.File
	var sheet *xlsx.Sheet
	var row *xlsx.Row
	var cell *xlsx.Cell
	file = xlsx.NewFile()
	sheet, _ = file.AddSheet("美逛后台数据")
	row = sheet.AddRow()
	cell = row.AddCell()
	cell.Value = "App账号"
	cell = row.AddCell()
	cell.Value = "邀请码"
	cell = row.AddCell()
	cell.Value = "小程序账号"
	cell = row.AddCell()
	cell.Value = "注册时间"
	cell = row.AddCell()
	cell.Value = "rid"
	cell = row.AddCell()
	cell.Value = "身份"
	cell = row.AddCell()
	cell.Value = "邀请模式"
	cell = row.AddCell()
	cell.Value = "分佣模式"
	cell = row.AddCell()
	cell.Value = "所属上级"
	cell = row.AddCell()
	cell.Value = "上级店长"
	cell = row.AddCell()
	cell.Value = "今日订单量"
	cell = row.AddCell()
	cell.Value = "今日预估"
	cell = row.AddCell()
	cell.Value = "本月预估"
	cell = row.AddCell()
	cell.Value = "可提现金额"
	cell = row.AddCell()
	cell.Value = "近30天收益"

	cell = row.AddCell()
	cell.Value = "备注"

	cell = row.AddCell()
	cell.Value = "支付宝姓名"

	cell = row.AddCell()
	cell.Value = "客服微信号"

	cell = row.AddCell()
	cell.Value = "微信号"

	for i := 0; i <= 50; i++ {
		fmt.Printf("抓取数据 %d/50...\n", i)
		datas, err := GetMeiguangData(i)
		if err != nil {
			fmt.Println(err)
			continue
		}
		AllDatas = append(AllDatas, datas...)
	}

	if len(AllDatas) == 0 {
		return fmt.Errorf("抓取数据异常，获取数据为空，请检查 zutuanweb 是否可用！")
	}

	fmt.Println("数据抓取完毕，正在导入 Excel 表格...")

	for _, d := range AllDatas {
		row = sheet.AddRow()
		// App账号
		cell = row.AddCell()
		cell.Value = d.Mobile

		// 邀请码
		cell = row.AddCell()
		cell.Value = d.InviteCode

		// 小程序账号
		cell = row.AddCell()
		cell.Value = ""

		// 注册时间
		cell = row.AddCell()
		cell.Value = d.CreateTime

		// rid
		cell = row.AddCell()
		cell.Value = strconv.FormatInt(d.RelationID, 10)

		// 身份
		cell = row.AddCell()
		cell.Value = d.StrUserType

		// 邀请模式
		cell = row.AddCell()
		cell.Value = ""

		// 分佣模式
		cell = row.AddCell()
		cell.Value = ""

		// 所属上级
		cell = row.AddCell()
		cell.Value = d.BefMobile

		// 上级店长
		cell = row.AddCell()
		befshopManager, err := GetBefShopManager(d.BefID)
		if err != nil {
			fmt.Printf("查询店长失败: %s\n", err)
			cell.Value = "查询失败"
		} else {
			cell.Value = befshopManager.Mobile
		}

		// 今日订单量
		cell = row.AddCell()
		cell.Value = strconv.FormatInt(d.TodayOrderCount, 10)

		// 今日预估
		cell = row.AddCell()
		var today float32 = float32(d.TodayRate) / 100
		cell.Value = fmt.Sprintf("%.2f", today)

		// 本月预估
		cell = row.AddCell()
		var month float32 = float32(d.MonthRate) / 100
		cell.Value = fmt.Sprintf("%.2f", month)

		// 可提现金额
		cell = row.AddCell()
		var remainAmount float32 = float32(d.RemainAmount) / 100
		cell.Value = fmt.Sprintf("%.2f", remainAmount)

		// 近30天收益
		cell = row.AddCell()
		var monthCommission float32 = float32(d.MonthSelfBuyCommission) / 100
		cell.Value = fmt.Sprintf("%.2f", monthCommission)

		// 备注
		cell = row.AddCell()
		cell.Value = d.Remark

		// 支付宝姓名
		cell = row.AddCell()
		cell.Value = d.AlipayName

		// 客服微信号
		cell = row.AddCell()
		cell.Value = ""

		// 微信号
		cell = row.AddCell()
		cell.Value = d.WeiXin
	}

	fmt.Println("数据抓取完毕，正在导入Excel表格...")
	if err := file.Save(filename); err != nil {
		fmt.Printf("保存Excel表格失败: %s\n", err)
		return err
	}
	return nil
}

func main() {
	fmt.Printf("zutuanweb 为: %s\n", zutuanweb)

	if err := run(); err != nil {
		fmt.Printf("抓取数据失败：%s\n", err)
		time.Sleep(10 * time.Second)
	}
}
