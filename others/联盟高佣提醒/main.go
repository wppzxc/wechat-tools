package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	// "github.com/lxn/walk"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog"
)

type appConfig struct {
	AppKey               string
	AppSecret            string
	CommissionPriceBegin float64 // 折扣起始价格
	CommissionPriceEnd   float64 // 折扣结束价格
	CommissionRateBegin  float64 // 佣金率起始
	CommissionRateEnd    float64 // 佣金率结束
	DingdingWebhookToken string
	KeyWordsString       string
	KeyWords             []string
	UserType             string
	Interval             int
	ZkPriceBegin         int
	ZkPriceEnd           int
	HasCoupon            bool
	// TaobaoAppKey         string
	// TaobaoAppSecret      string
	// TaobaoAdzongID       string
	// CheckTaoLiJin        bool
}

const (
	dingdingWebhookURL = "https://oapi.dingtalk.com/robot/send?access_token="
)

var (
	config        *appConfig
	stopCh        chan struct{}
	running       bool
	startButton   *walk.PushButton
	dataokeClient *DaTaoKeClient
	errorMsg      *walk.Label
	lastWorkTime  string
	sendedItems   *sync.Map
)

func main() {
	klog.InitFlags(nil)
	flag.Set("log_file", "./lianmeng_gaoyong.log")
	flag.Set("log_file_max_size", "100")
	flag.Set("logtostderr", "false")
	flag.Set("alsologtostderr", "false")
	flag.Parse()

	defer klog.Flush()
	config = new(appConfig)
	config.AppKey = "5e9d2dbadc286"
	config.AppSecret = "8f3c81484fdf7bd2695ddbbc6a128201"
	config.DingdingWebhookToken = "9d7888ccd788c2aed4fcc79377c02400a816cd18db6598f09ecd152c7daf8466"
	config.Interval = 1
	config.KeyWordsString = "学生笔记本/洗碗刷/暖宝宝/中性笔/铅笔/折纸/颜料/小白鞋清洗/晨光/活页本/练字帖/油画棒/文具/跳绳/气球/数据线/防水袋/得力/记号笔/修正带/耳机/灯泡/玩具/筷子/胶带/粘毛器/笔芯/毛巾/圆珠笔/挂钩/牙刷/头绳"
	if _, err := (MainWindow{
		Title:  "大淘客高佣提醒",
		Size:   Size{Width: 400, Height: 300},
		Layout: VBox{},
		DataBinder: DataBinder{
			AutoSubmit: true,
			DataSource: config,
		},
		Children: []Widget{
			Composite{
				Layout: HBox{},
				Children: []Widget{
					Label{
						Text: "大淘客AppKey",
					},
					LineEdit{
						Text: Bind("AppKey"),
					},
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					Label{
						Text: "大淘客AppSecret",
					},
					LineEdit{
						Text:         Bind("AppSecret"),
						PasswordMode: true,
					},
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					Label{
						Text: "钉钉接口token",
					},
					LineEdit{
						Text:         Bind("DingdingWebhookToken"),
						PasswordMode: true,
					},
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					Label{
						Text: "店铺类型",
					},
					RadioButtonGroup{
						DataMember: "UserType",
						Buttons: []RadioButton{
							RadioButton{
								Name:  "Taobao",
								Text:  "淘宝",
								Value: "0",
							},
							RadioButton{
								Name:  "Tianmao",
								Text:  "天猫",
								Value: "1",
							},
							RadioButton{
								Name:  "All",
								Text:  "全部",
								Value: "2",
							},
						},
					},
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					Label{
						Text: "是否有优惠券",
					},
					CheckBox{
						Checked: Bind("HasCoupon"),
					},
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					Label{
						Text: "刷新间隔(默认1分钟)",
					},
					NumberEdit{
						Value:    Bind("Interval", Range{0, 60}),
						Decimals: 0,
						Suffix:   " /min",
					},
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					Label{
						Text: "关键词(用 / 分隔)",
					},
					LineEdit{
						Text: Bind("KeyWordsString"),
					},
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					Label{
						Text: "券后价范围",
					},
					NumberEdit{
						Value:    Bind("CommissionPriceBegin", Range{0, 9999}),
						Decimals: 2,
					},
					Label{
						Text: "~",
					},
					NumberEdit{
						Value:    Bind("CommissionPriceEnd", Range{0, 9999}),
						Decimals: 2,
					},
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					Label{
						Text: "折扣价范围",
					},
					NumberEdit{
						Value:    Bind("ZkPriceBegin", Range{0, 9999}),
						Decimals: 0,
					},
					Label{
						Text: "~",
					},
					NumberEdit{
						Value:    Bind("ZkPriceEnd", Range{0, 9999}),
						Decimals: 0,
					},
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					Label{
						Text: "佣金率范围",
					},
					NumberEdit{
						Value:    Bind("CommissionRateBegin", Range{0, 99}),
						Decimals: 2,
						Suffix:   " /%",
					},
					Label{
						Text: "~",
					},
					NumberEdit{
						Value:    Bind("CommissionRateEnd", Range{0, 99}),
						Decimals: 2,
						Suffix:   " /%",
					},
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					PushButton{
						Text:      "开始",
						OnClicked: startLoopCheck,
						AssignTo:  &startButton,
					},
					PushButton{
						Text:      "停止",
						OnClicked: stopLoopCheck,
					},
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					Label{
						TextColor: walk.RGB(0xff, 0, 0),
						AssignTo:  &errorMsg,
					},
				},
			},
		},
	}).Run(); err != nil {
		panic(err)
	}
}

func startLoopCheck() {
	if len(config.AppKey) == 0 ||
		len(config.AppSecret) == 0 ||
		len(config.DingdingWebhookToken) == 0 ||
		len(config.KeyWordsString) == 0 {
		errorMsg.SetText("错误：请输入完整参数!")
		return
	}
	config.KeyWords = strings.Split(config.KeyWordsString, "/")
	if len(config.KeyWords) == 0 {
		errorMsg.SetText("错误：请输入关键词!")
		return
	}
	if len(config.UserType) == 0 {
		klog.Info("未指定店铺类型，默认全部")
		config.UserType = "2"
	}
	if config.Interval == 0 {
		klog.Info("未指定刷新间隔，默认1分钟")
		config.Interval = 1
	}
	running = true
	stopCh = make(chan struct{})
	dataokeClients := make([]*DaTaoKeClient, 0)
	for _, key := range config.KeyWords {
		dataokeClients = append(dataokeClients, NewDataokeClientWithKeyWord(key))
	}
	for _, client := range dataokeClients {
		go wait.Until(client.worker, time.Duration(config.Interval)*time.Minute, stopCh)
	}
	// 24小时清空一次已发送信息
	go wait.Until(func() { sendedItems = new(sync.Map); klog.Info("初始化 sendedItems") }, 24*time.Hour, stopCh)
	errorMsg.SetText("")
	startButton.SetEnabled(false)
}

func stopLoopCheck() {
	if running {
		running = false
		close(stopCh)
		startButton.SetEnabled(true)
	}
}

func (dtk *DaTaoKeClient) worker() {
	// 获取商品信息
	items := make([]DaTaoKeItem, 0)
	for i := 1; i <= 1; i++ {
		tmpItems, err := dtk.getTbService(i)
		if err != nil {
			klog.Errorf("'%s' 拉取第%d页商品失败: %s", dtk.KeyWord, i, err)
			continue
		}
		items = append(items, tmpItems...)
	}
	if len(items) == 0 {
		klog.Infof("关键词'%s': 拉取商品数量为0", dtk.KeyWord)
		errorMsg.SetText(fmt.Sprintf("错误(%s): 拉取联盟商品信息'%s'失败: 拉取商品数量为0", time.Now().Format("2006-01-02 15:04:05"), dtk.KeyWord))
		return
	}
	klog.Infof("拉取到%d件联盟商品", len(items))

	// 过滤已发送信息
	// 过滤店铺类型
	tmpSendItems := make([]DaTaoKeItem, 0)
	for _, item := range items {
		userType := strconv.Itoa(item.UserType)
		switch config.UserType {
		case "2":
			if _, ok := sendedItems.Load(item.ItemID); !ok {
				tmpSendItems = append(tmpSendItems, item)
			}
		default:
			if userType == config.UserType {
				if _, ok := sendedItems.Load(item.ItemID); !ok {
					tmpSendItems = append(tmpSendItems, item)
				}
			}
		}
	}
	klog.Infof("过滤到%d件未发送商品", len(tmpSendItems))

	// 过滤券后价
	needSendItems := make([]DaTaoKeItem, 0)
	for _, item := range tmpSendItems {
		// 优惠券金额
		couponAmount, err := strconv.ParseFloat(item.CouponAmount, 32)
		if err != nil {
			couponAmount = 0
			klog.Errorf("转换优惠券金额 '%s' 错误: %s", item.CouponAmount, err)
		}
		// 折扣价格
		finalPrice, err := strconv.ParseFloat(item.ZkFinalPrice, 32)
		if err != nil {
			klog.Errorf("转换折扣价格 '%s' 错误: %s", item.ZkFinalPrice, err)
			finalPrice = 0
		}

		// 券后价 = 折扣价格 - 优惠券面额
		actualPrice := finalPrice - couponAmount
		if actualPrice >= config.CommissionPriceBegin && actualPrice <= config.CommissionPriceEnd {
			needSendItems = append(needSendItems, item)
			continue
		}
		klog.Infof("商品%d，优惠券金额'%s', 折扣价格'%s', 券后价(折扣价格 - 优惠券)'%.2f' 不在指定范围 %.2f~%.2f", item.ItemID, item.CouponAmount, item.ZkFinalPrice, actualPrice, config.CommissionPriceBegin, config.CommissionPriceEnd)
	}
	klog.Infof("'%s' 过滤到%d件符合券后价条件的商品", dtk.KeyWord, len(needSendItems))
	if len(needSendItems) == 0 {
		klog.Infof("关键词%s: 没有商品可发送", dtk.KeyWord)
		return
	}
	// 开始发送到钉钉
	for _, sendItem := range needSendItems {
		if err := alertDingding(sendItem); err != nil {
			klog.Errorf("关键词%s: %d发送失败: ", dtk.KeyWord, sendItem.ItemID)
			continue
		}
		klog.Infof("关键词%s: %d发送成功", dtk.KeyWord, sendItem.ItemID)
		sendedItems.Store(sendItem.ItemID, sendItem)
	}
}

func alertDingding(item DaTaoKeItem) error {
	msg := `{
		"msgtype": "markdown", 
		"markdown": {
			"title": "联盟高佣提醒",
			"text": "%s"
		}
	}`

	imageMsg := `{
		"msgtype": "markdown", 
		"markdown": {
			"title": "商品图片",
			"text": "%s"
		}
	}`
	imageFormat := `![screenshot](%s)`

	// 1，商品id
	// 2，商品短标题
	// 3，券后价
	// 4，佣金比例
	// 5，优惠券开始时间
	// 6，店铺类型：天猫或淘宝
	// 7，到手佣金
	// 8，实际成本
	// 9，商品白底图
	content := `#### 联盟商品高佣提醒
	1. 商品id: %d
	2. 商品短标题: %s
	3. 券后价: %s
	4. 佣金比例: %s
	5. 优惠券开始时间: %s
	6. 店铺类型：%s
	7. 到手佣金: %s
	8. 实际成本: %s
	`

	shopType := "天猫"
	if item.UserType == 0 {
		shopType = "集市"
	}

	// 优惠券金额
	couponAmount, err := strconv.ParseFloat(item.CouponAmount, 32)
	if err != nil {
		couponAmount = 0
		klog.Errorf("转换优惠券金额 '%s' 错误: %s", item.CouponAmount, err)
	}
	// 折扣价格
	finalPrice, err := strconv.ParseFloat(item.ZkFinalPrice, 32)
	if err != nil {
		klog.Errorf("转换折扣价格 '%s' 错误: %s", item.ZkFinalPrice, err)
		finalPrice = 0
	}
	// 佣金比率
	commissionRate, err := strconv.ParseFloat(item.CommissionRate, 32)
	if err != nil {
		klog.Errorf("转换佣金比率 '%s' 错误: %s", item.CommissionRate, err)
		commissionRate = 0
	}
	commissionRate = commissionRate / 100

	// 券后价 = 折扣价格 - 优惠券面额
	actualPrice := finalPrice - couponAmount
	//  到手佣金 = 券后价 * 佣金比例 * 90%
	finalCommisson := actualPrice * commissionRate * 0.9
	//  到手成本 = 券后价-到手佣金
	finalCost := actualPrice - finalCommisson
	contentText := fmt.Sprintf(content,
		item.ItemID,
		item.ShortTitle,
		fmt.Sprintf("%.2f", actualPrice),
		item.CommissionRate,
		item.CouponStartTime,
		shopType,
		fmt.Sprintf("%.2f", finalCommisson),
		fmt.Sprintf("%.2f", finalCost),
	)

	// 发送图片
	imageContent := fmt.Sprintf(imageFormat, item.WhiteImage)
	sendImageMsg := fmt.Sprintf(imageMsg, imageContent)
	if len(item.WhiteImage) == 0 {
		sendImageMsg = fmt.Sprintf(imageMsg, ".没有图片")
	}
	if _, err := http.Post(dingdingWebhookURL+config.DingdingWebhookToken, "application/json", bytes.NewBufferString(sendImageMsg)); err != nil {
		klog.Error(err)
		return err
	}
	klog.Info("发送图片成功")
	// 发送文案
	sendText := fmt.Sprintf(msg, contentText)
	klog.Info(sendText)
	if _, err := http.Post(dingdingWebhookURL+config.DingdingWebhookToken, "application/json", bytes.NewBufferString(sendText)); err != nil {
		klog.Error(err)
		return err
	}
	klog.Info("发送文案成功")
	return nil
}

//
//
//
//
//
const (
	defaultDataokeApiVersion = "v2.1.0"
	defaultGetRankingListUrl = "https://openapi.dataoke.com/api/tb-service/get-tb-service"
)

// 榜单类型，1.实时榜 2.全天榜 3.热推榜 4.复购榜 5.热词飙升榜 6.热词排行榜 7.综合热搜榜
const (
	RankTypeRealTimeList         = "1"
	RankTypeAllDayList           = "2"
	RankTypeHotRecommendList     = "3"
	RankTypeRepurchaseList       = "4"
	RankTypeHotWordsRiseList     = "5"
	RankTypeAllHotWordsList      = "6"
	RankTypeComprehensiveHotList = "7"
)

type DaTaoKeClient struct {
	CommonParams url.Values
	InputParams  url.Values
	Client       *http.Client
	ReqUrl       string
	Sign         string
	KeyWord      string
}

func NewDataokeClient() *DaTaoKeClient {
	commonParams := url.Values{
		"version": []string{defaultDataokeApiVersion},
	}
	dtkClient := DaTaoKeClient{
		CommonParams: commonParams,
		InputParams:  nil,
		Client:       new(http.Client),
		Sign:         "",
	}
	return &dtkClient
}

func NewDataokeClientWithKeyWord(keyWord string) *DaTaoKeClient {
	commonParams := url.Values{
		"version": []string{defaultDataokeApiVersion},
	}
	dtkClient := DaTaoKeClient{
		CommonParams: commonParams,
		InputParams:  nil,
		Client:       new(http.Client),
		Sign:         "",
		KeyWord:      keyWord,
	}
	return &dtkClient
}

const (
	defaultTaoLiJinMethodVersion = "2.0"
	defaultSignMethod            = "md5"
	createTaoLiJinMethod         = "taobao.tbk.dg.vegas.tlj.create"
	defaultTaoLiJinName          = "淘礼金来了"
	taobaoRouteUrlProd           = "http://gw.api.taobao.com/router/rest"
)

func UpdateDataokeClientWithConfig(dataokeClient *DaTaoKeClient, config *appConfig) *DaTaoKeClient {
	commonParams := url.Values{
		"appKey":  []string{config.AppKey},
		"version": []string{defaultDataokeApiVersion},
	}
	dataokeClient.CommonParams = commonParams
	return dataokeClient
}

func (dtk *DaTaoKeClient) getTbService(pageNum int) ([]DaTaoKeItem, error) {
	dtk.ReqUrl = "https://openapi.dataoke.com/api/tb-service/get-tb-service"
	dtk.InputParams = url.Values{
		"pageSize":    []string{"100"},
		"pageNo":      []string{fmt.Sprintf("%d", pageNum)},
		"keyWords":    []string{dtk.KeyWord},
		"sort":        []string{"tk_rate_des"},
		"source":      []string{"1"},
		"startPrice":  []string{fmt.Sprintf("%d", config.ZkPriceBegin)},
		"endPrice":    []string{fmt.Sprintf("%d", config.ZkPriceEnd)},
		"startTkRate": []string{fmt.Sprintf("%.0f", config.CommissionRateBegin*100)},
		"endTkRate":   []string{fmt.Sprintf("%.0f", config.CommissionRateEnd*100)},
		"hasCoupon":   []string{fmt.Sprintf("%t", config.HasCoupon)},
	}
	dtk.CommonParams = url.Values{
		"appKey":  []string{config.AppKey},
		"version": []string{"v2.1.0"},
	}
	dataokeResp, err := dtk.sign(config.AppSecret).do()
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	if dataokeResp.Code != 0 {
		return nil, fmt.Errorf(dataokeResp.Msg)
	}
	return dataokeResp.Data, nil
}

func (dtk *DaTaoKeClient) sign(secret string) *DaTaoKeClient {
	timestamp := strconv.FormatInt(time.Now().UnixNano()/1e6, 10)
	dtk.CommonParams.Set("nonce", "123456")
	dtk.CommonParams.Set("timer", timestamp)

	commonStr := fmt.Sprintf("appKey=%s&timer=%s&nonce=%s&key=%s", config.AppKey, timestamp, "123456", config.AppSecret)
	h := md5.New()
	h.Write([]byte(commonStr))
	sign := hex.EncodeToString(h.Sum(nil))

	dtk.Sign = sign
	return dtk
}

func (dtk *DaTaoKeClient) do() (*DaTaoKeResponse, error) {
	params := url.Values{}
	for k, v := range dtk.CommonParams {
		params.Set(k, v[0])
	}
	for k, v := range dtk.InputParams {
		params.Set(k, v[0])
	}
	params.Set("signRan", dtk.Sign)
	reqUrl := fmt.Sprintf("%s?%s", dtk.ReqUrl, params.Encode())

	resp, err := dtk.Client.Get(reqUrl)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	// klog.Info(string(data))

	respData := new(DaTaoKeResponse)
	if err := json.Unmarshal(data, respData); err != nil {
		klog.Error(err)
		return nil, err
	}

	return respData, nil
}

//
type DaTaoKeResponse struct {
	Time int64         `json:"time"`
	Code int           `json:"code"`
	Msg  string        `json:"msg"`
	Data []DaTaoKeItem `json:"data"`
	// TotalNum int64           `json:"totalNum"`
	// PageId   string          `json:"pageId"`
}

type DaTaoKeItem struct {
	Title                  string `json:"title"`
	Volume                 int    `json:"volume"`
	Nick                   string `json:"nick"`
	CouponStartTime        string `json:"coupon_start_time"`
	CouponEndTime          string `json:"coupon_end_time"`
	TkTotalSales           string `json:"tk_total_sales"`
	CouponID               string `json:"coupon_id"`
	PictURL                string `json:"pict_url"`
	ReservePrice           string `json:"reserve_price"`
	ZkFinalPrice           string `json:"zk_final_price"`
	UserType               int    `json:"user_type"`
	CommissionRate         string `json:"commission_rate"`
	SellerID               int64  `json:"seller_id"`
	CouponTotalCount       int    `json:"coupon_total_count"`
	CouponRemainCount      int    `json:"coupon_remain_count"`
	CouponInfo             string `json:"coupon_info"`
	ShopTitle              string `json:"shop_title"`
	ShopDsr                int64  `json:"shop_dsr"`
	LevelOneCategoryName   string `json:"level_one_category_name"`
	LevelOneCategoryID     int64  `json:"level_one_category_id"`
	CategoryName           string `json:"category_name"`
	CategoryID             int64  `json:"category_id"`
	ShortTitle             string `json:"short_title"`
	WhiteImage             string `json:"white_image"`
	CouponStartFee         string `json:"coupon_start_fee"`
	CouponAmount           string `json:"coupon_amount"`
	ItemDescription        string `json:"item_description"`
	ItemID                 int64  `json:"item_id"`
	YsylTljFace            int    `json:"ysyl_tlj_face"`
	PresaleDeposit         int    `json:"presale_deposit"`
	PresaleDiscountFeeText string `json:"presale_discount_fee_text"`
}
