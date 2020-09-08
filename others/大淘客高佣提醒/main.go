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
	"sort"
	"strconv"
	"strings"
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
	CommissionPriceBegin float64
	CommissionPriceEnd   float64
	CommissionRateBegin  int
	CommissionRateEnd    int
	DingdingWebhookToken string
	TaobaoAppKey         string
	TaobaoAppSecret      string
	TaobaoAdzongID       string
	CheckTaoLiJin        bool
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
	taobaoClient  *TaoBaoClient
	errorMsg      *walk.Label
	lastWorkTime  string
)

func main() {
	klog.InitFlags(nil)
	flag.Set("log_file", "./dataoke_gaoyong.log")
	flag.Set("log_file_max_size", "100")
	flag.Set("logtostderr", "false")
	flag.Set("alsologtostderr", "false")
	flag.Parse()

	defer klog.Flush()
	config = new(appConfig)
	config.AppKey = "5e9d2dbadc286"
	config.AppSecret = "8f3c81484fdf7bd2695ddbbc6a128201"
	config.DingdingWebhookToken = "9d7888ccd788c2aed4fcc79377c02400a816cd18db6598f09ecd152c7daf8466"
	config.TaobaoAppKey = "30029015"
	config.TaobaoAppSecret = "b9f71d954e6a331a160c6c33956f7c44"
	config.TaobaoAdzongID = "110409800054"
	dataokeClient = NewDataokeClient()
	taobaoClient = NewTaoBaoClient()
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
						Text: "淘宝AppKey",
					},
					LineEdit{
						Text: Bind("TaobaoAppKey"),
					},
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					Label{
						Text: "淘宝AppSecret",
					},
					LineEdit{
						Text:         Bind("TaobaoAppSecret"),
						PasswordMode: true,
					},
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					Label{
						Text: "淘宝推广位id",
					},
					LineEdit{
						Text: Bind("TaobaoAdzongID"),
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
						Text: "券后价范围",
					},
					NumberEdit{
						Value:    Bind("CommissionPriceBegin", Range{0, 99}),
						Decimals: 2,
					},
					Label{
						Text: "~",
					},
					NumberEdit{
						Value:    Bind("CommissionPriceEnd", Range{0, 99}),
						Decimals: 2,
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
						Decimals: 0,
						Suffix:   " /%",
					},
					Label{
						Text: "~",
					},
					NumberEdit{
						Value:    Bind("CommissionRateEnd", Range{0, 99}),
						Decimals: 0,
						Suffix:   " /%",
					},
				},
			},
			// Composite{
			// 	Layout: HBox{},
			// 	Children: []Widget{
			// 		Label{
			// 			Text: "校验淘礼金",
			// 		},
			// 		CheckBox{
			// 			Checked: Bind("CheckTaoLiJin"),
			// 		},
			// 	},
			// },
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
		len(config.TaobaoAppKey) == 0 ||
		len(config.TaobaoAppSecret) == 0 ||
		len(config.TaobaoAdzongID) == 0 {
		errorMsg.SetText("错误：请输入完整参数!")
		return
	}
	running = true
	stopCh = make(chan struct{})
	dataokeClient = UpdateDataokeClientWithConfig(dataokeClient, config)
	taobaoClient = UpdateTaobaoClientWithConfig(taobaoClient, config)
	go wait.Until(work, 60*time.Second, stopCh)
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

func work() {
	if len(lastWorkTime) == 0 {
		lastWorkTime = time.Now().Add(time.Minute * -1).Format("2006-01-02 15:04:05")
	}
	defer func() {
		lastWorkTime = time.Now().Format("2006-01-02 15:04:05")
	}()
	items, err := dataokeClient.GetPullGoodsByTime(lastWorkTime)
	if err != nil {
		klog.Error(err)
		errorMsg.SetText(fmt.Sprintf("错误(%s): 拉取大淘客商品信息失败: %s", time.Now().Format("2006-01-02 15:04:05"), err))
		return
	}
	if len(items) == 0 {
		klog.Info("未找到新上架商品信息")
		return
	}

	klog.Infof("拉取新上架商品%d件", len(items))
	for _, item := range items {
		// 佣金率符合
		if item.CommissionRate >= float32(config.CommissionRateBegin) && item.CommissionRate <= float32(config.CommissionRateEnd) {
			// 佣金符合
			if float64(item.ActualPrice) >= config.CommissionPriceBegin && float64(item.ActualPrice) <= config.CommissionPriceEnd {

				if config.CheckTaoLiJin {
					// 是否成功创建淘礼金
					klog.Infof("查询商品 %s:%s 是否可以创建淘礼金", item.DTitle, item.GoodsId)
					if err := canCreateTaoLiJin(item); err != nil {
						klog.Error(err)
						continue
					}
					klog.Infof("商品 %s:%s 可以创建淘礼金!", item.DTitle, item.GoodsId)
				}
				klog.Infof("发送消息到钉钉")
				if err := alertDingding(item); err != nil {
					klog.Error(err)
					errorMsg.SetText(fmt.Sprintf("错误(%s): 拉取大淘客商品信息失败: %s", time.Now().Format("2006-01-02 15:04:05"), err))
					continue
				}
				klog.Info("发送成功")
			}
		}
	}
}

func alertDingding(item DaTaoKeItem) error {
	msg := `{
		"msgtype": "text", 
		"text": {
			"content": "%s"
		}
	}`

	content := `大淘客发现最新的高佣商品
	1，商品链接: %s
	2，短标题: %s
	3，券后价: %.2f
	4，佣金比例: %.2f
	5，优惠券开始时间: %s
	6，店铺类型: %s
	7，到手佣金: %.2f
	8，实际成本: %.2f
`

	shopName := "天猫"
	if item.ShopType == 0 {
		shopName = "淘宝"
	}
	//  到手佣金 = 券后价*佣金比例*90%
	finalCommisson := item.ActualPrice * (item.CommissionRate / 100) * 0.9
	//  到手成本 = 券后价-到手佣金
	finalCost := item.ActualPrice - finalCommisson
	contentText := fmt.Sprintf(content,
		item.ItemLink,
		item.DTitle,
		item.ActualPrice,
		item.CommissionRate,
		item.CouponStartTime,
		shopName,
		finalCommisson,
		finalCost)

	sendText := fmt.Sprintf(msg, contentText)
	klog.Info(sendText)
	if _, err := http.Post(dingdingWebhookURL+config.DingdingWebhookToken, "application/json", bytes.NewBufferString(sendText)); err != nil {
		klog.Error(err)
		return err
	}
	klog.Infof("发送商品信息: '%+v' 到钉钉成功", item)

	goodsIDMsg := fmt.Sprintf(msg, item.GoodsId+".")
	if _, err := http.Post(dingdingWebhookURL+config.DingdingWebhookToken, "application/json", bytes.NewBufferString(goodsIDMsg)); err != nil {
		klog.Error(err)
		return err
	}
	klog.Infof("发送商品id: '%s' 到钉钉成功", item.GoodsId)
	return nil
}

//
//
//
//
//
const (
	defaultDataokeApiVersion = "v1.2.2"
	defaultGetRankingListUrl = "https://openapi.dataoke.com/api/goods/get-ranking-list"
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

const (
	defaultTaoLiJinMethodVersion = "2.0"
	defaultSignMethod            = "md5"
	createTaoLiJinMethod         = "taobao.tbk.dg.vegas.tlj.create"
	defaultTaoLiJinName          = "淘礼金来了"
	taobaoRouteUrlProd           = "http://gw.api.taobao.com/router/rest"
)

func NewTaoBaoClient() *TaoBaoClient {
	now := time.Now()
	startTime := now.Format("2006-01-02 15:04:05")
	//endTime := now.Add(24 * time.Hour).Format("2006-01-02 15:04:05")
	commonParams := url.Values{
		"app_key":     []string{config.TaobaoAppKey},
		"sign_method": []string{defaultSignMethod},
		"timestamp":   []string{startTime},
		"v":           []string{defaultTaoLiJinMethodVersion},
		"format":      []string{"json"},
	}

	tbClient := TaoBaoClient{
		CommonParams: commonParams,
		InputParams:  nil,
		Client:       new(http.Client),
	}
	return &tbClient
}

func UpdateDataokeClientWithConfig(dataokeClient *DaTaoKeClient, config *appConfig) *DaTaoKeClient {
	commonParams := url.Values{
		"appKey":  []string{config.AppKey},
		"version": []string{defaultDataokeApiVersion},
	}
	dataokeClient.CommonParams = commonParams
	return dataokeClient
}

func UpdateTaobaoClientWithConfig(taobaoClient *TaoBaoClient, config *appConfig) *TaoBaoClient {

	taobaoClient.CommonParams.Set("app_key", config.TaobaoAppKey)
	return taobaoClient
}

func (dtk *DaTaoKeClient) GetRealTimeListItem() ([]DaTaoKeItem, error) {
	dtk.ReqUrl = defaultGetRankingListUrl
	dtk.InputParams = url.Values{
		"rankType": []string{RankTypeRealTimeList},
	}
	dataokeResp, err := dtk.sign(config.AppSecret).do()
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	if dataokeResp.Code != 0 {
		return nil, fmt.Errorf(dataokeResp.Msg)
	}
	return dataokeResp.Data.List, nil
}

func (dtk *DaTaoKeClient) GetPullGoodsByTime(lastWorkTime string) ([]DaTaoKeItem, error) {
	klog.Infof("拉取商品起始时间为: %s", lastWorkTime)
	dtk.ReqUrl = "https://openapi.dataoke.com/api/goods/pull-goods-by-time"
	dtk.InputParams = url.Values{
		"pageSize":  []string{"200"},
		"pageId":    []string{"1"},
		"startTime": []string{lastWorkTime},
		"sort":      []string{"4"},
		// "pre":       []string{"1"},
	}
	dtk.CommonParams = url.Values{
		"appKey":  []string{config.AppKey},
		"version": []string{"v1.2.3"},
	}
	dataokeResp, err := dtk.sign(config.AppSecret).do()
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	if dataokeResp.Code != 0 {
		return nil, fmt.Errorf(dataokeResp.Msg)
	}
	return dataokeResp.Data.List, nil
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
	klog.Info(string(data))

	respData := new(DaTaoKeResponse)
	if err := json.Unmarshal(data, respData); err != nil {
		klog.Error(err)
		return nil, err
	}

	return respData, nil
}

//
type DaTaoKeResponse struct {
	Time     int64           `json:"time"`
	Code     int             `json:"code"`
	Msg      string          `json:"msg"`
	Data     DaTaoKeDataList `json:"data"`
	TotalNum int64           `json:"totalNum"`
	PageId   string          `json:"pageId"`
}

type DaTaoKeDataList struct {
	List []DaTaoKeItem `json:"list"`
}

//id 				Number 	19259135 						商品id，在大淘客的商品id
//goodsId 			Number 	590858626868 					淘宝商品id
//ranking 			Number 	1 								榜单名次
//newRankingGoods 	Number 	1 								是否新上榜商品（12小时内入榜的商品） 0.否1.是
//dtitle 			String 	【李佳琦推荐】奢华芯肌素颜爆水霜 	短标题
//actualPrice 		Number 	39.9 							券后价
//commissionRate 	Number 	30 								佣金比例
//couponPrice 		Number 	300 							优惠券金额
//couponReceiveNum 	Number 	4000 							领券量
//couponTotalNum 	Number 	10000 							券总量
//monthSales 		Number 	8824 							月销量
//twoHoursSales 	Number 	1542 							2小时销量
//dailySales 		Number 	4545 							当天销量
//hotPush 			Number 	42 								热推值
//mainPic 			String 	“https://img.alicdn.com/i4/1687451966/O1CN01rTeKnv1QOTBnyOXDe\_!!1687451966.jpg“ 商品图
//title 			String 	“2019新款运动短裤女宽松防走光韩版外穿ins潮休闲学生bf夏季阔腿” 商品长标题
//desc 				String 	“多款可选！显瘦高腰韩版阔腿裤五分裤，不起球，不掉色。舒适面料，不挑身材，高腰设计” 商品描述
//originalPrice	 	Number 	29.9 							商品原价
//couponLink 		String 	“https://uland.taobao.com/quan/detail?sellerId=1687451966&activityId=ffef827d9a5747efbbe02a93c6d7ec13“ 优惠券链接
//couponStartTime 	String 	“2019-06-04 00:00:00” 			优惠券开始时间
//couponEndTime 	String 	“2019-06-06 23:59:59” 			优惠券结束时间
//commissionType 	Number 	3 								佣金类型
//createTime 		String 	“2019-06-03 17:55:18” 			创建时间
//activityType 		Number 	1 								活动类型
//picList 			Array 	“https://img.alicdn.com/imgextra/i2/1687451966/O1CN01WNuZcl1QOTCM9NsrO_!!1687451966.jpg,https://img.alicdn.com/imgextra/i4/1687451966/O1CN01h2ih4v1QOTCOxlZDj_!!1687451966.jpg“ 营销图
//guideName 		String 	易折网 							放单人名称
//shopType 			Number 	1 								店铺类型
//couponConditions 	Number 	29 								优惠券使用条件
//avgSales 			Number 	586 							日均销量（仅复购榜返回该字段）
//entryTime 		String 	“2019-06-06 10:59:59” 			入榜时间（仅复购榜返回该字段）
//sellerId 			String 	4014489195 						淘宝卖家id
//quanMLink 		Number 	10 								定金，若无定金，则显示0
//hzQuanOver 		Number 	100 							立减，若无立减金额，则显示0
//yunfeixian 		Number 	1 								0.不包运费险 1.包运费险
//estimateAmount 	Number 	25.2 							预估淘礼金
//freeshipRemoteDistrict Number 1 							偏远地区包邮，0.不包邮，1.包邮
//top 				Number 	1 								热词榜排名（适用于5.热词飙升榜6.热词排行榜）
//keyWord 			String 	螺蛳粉 							热搜词（适用于5.热词飙升榜6.热词排行榜）
//upVal 			Number 	1 								排名提升值（适用于5.热词飙升榜）
//hotVal 			Number 	123454 							排名热度值

type DaTaoKeItem struct {
	Id                     int64   `json:"id"`
	GoodsId                string  `json:"goodsId"`
	Ranking                int     `json:"ranking"`
	DTitle                 string  `json:"dtitle"`
	ActualPrice            float32 `json:"actualPrice"`
	CommissionRate         float32 `json:"commissionRate"`
	CouponPrice            float32 `json:"couponPrice"`
	CouponReceiveNum       int     `json:"couponReceiveNum"`
	CouponTotalNum         int     `json:"couponTotalNum"`
	MonthSales             int     `json:"monthSales"`
	TwoHoursSales          int     `json:"twoHoursSales"`
	DailySales             int     `json:"dailySales"`
	HotPush                int     `json:"hotPush"`
	MainPic                string  `json:"mainPic"`
	Title                  string  `json:"title"`
	Desc                   string  `json:"desc"`
	OriginalPrice          float32 `json:"originalPrice"`
	CouponLink             string  `json:"couponLink"`
	CouponStartTime        string  `json:"couponStartTime"`
	CouponEndTime          string  `json:"couponEndTime"`
	CommissionType         int     `json:"commissionType"`
	CreateTime             string  `json:"createTime"`
	ActivityType           int     `json:"activityType"`
	Imgs                   string  `json:"imgs"`
	GuideName              string  `json:"guideName"`
	ShopType               int     `json:"shopType"`
	CouponConditions       string  `json:"couponConditions"`
	NewRankingGoods        int     `json:"newRankingGoods"`
	SellerId               string  `json:"sellerId"`
	QuanMLink              int     `json:"quanMLink"`
	HzQuanOver             int     `json:"hzQuanOver"`
	Yunfeixian             int     `json:"yunfeixian"`
	EstimateAmount         float32 `json:"estimateAmount"`
	FreeshipRemoteDistrict int     `json:"freeshipRemoteDistrict"`
	ItemLink               string  `json:"itemLink"`
}

type TaoBaoClient struct {
	CommonParams url.Values
	InputParams  url.Values
	Client       *http.Client
	Sign         string
}

func canCreateTaoLiJin(item DaTaoKeItem) error {
	perFaceStr := "1.01"
	tljURL, err := taobaoClient.CreateTaoLiJinUrl(item.GoodsId, perFaceStr, "")
	if err != nil {
		klog.Error(err)
		return fmt.Errorf("创建淘礼金失败: %s", err)
	}
	if len(tljURL) == 0 {
		klog.Error("无法创建淘礼金")
		return fmt.Errorf("无法创建淘礼金")
	}
	return nil
}

func (tbc *TaoBaoClient) CreateTaoLiJinUrl(itemId string, perFace string, name string) (string, error) {
	tbc.CommonParams.Set("method", createTaoLiJinMethod)
	if len(name) == 0 {
		name = defaultTaoLiJinName
	}
	now := time.Now()
	startTime := now.Format("2006-01-02 15:04:05")
	endTime := now.Add(24 * time.Hour).Format("2006-01-02 15:04:05")
	// tbc.CommonParams.Set("timestamp", startTime)
	tbc.InputParams = url.Values{
		"adzone_id":                []string{"110409800054"},
		"item_id":                  []string{itemId},
		"total_num":                []string{"1"},
		"name":                     []string{name},
		"user_total_win_num_limit": []string{"1"},
		"security_switch":          []string{"false"},
		"per_face":                 []string{perFace},
		"send_start_time":          []string{startTime},
		"send_end_time":            []string{endTime},
	}

	taobaoResp, err := tbc.sign(config.TaobaoAppSecret).Do()
	if err != nil {
		klog.Error(err)
		return "", err
	}

	if taobaoResp.TbkDgVegasTljCreateResponse.Result.Success {
		return taobaoResp.TbkDgVegasTljCreateResponse.Result.Model.SendUrl, nil
	}

	return "", fmt.Errorf("Error in create tao li jin '%+v'", taobaoResp)
}

type ParamSlice [][]string

func (p ParamSlice) Len() int           { return len(p) }
func (p ParamSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p ParamSlice) Less(i, j int) bool { return p[i][0] < p[j][0] }

func (tbc *TaoBaoClient) sign(secret string) *TaoBaoClient {

	params := make(ParamSlice, 0)
	for k, v := range tbc.CommonParams {
		params = append(params, []string{k, v[0]})
	}
	for k, v := range tbc.InputParams {
		params = append(params, []string{k, v[0]})
	}
	sort.Sort(params)

	str := ""
	for _, s := range params {
		str = str + s[0] + s[1]
	}

	str = secret + str + secret
	h := md5.New()
	h.Write([]byte(str))
	sign := strings.ToUpper(hex.EncodeToString(h.Sum(nil)))

	tbc.Sign = sign
	return tbc
}

type TaoBaoApiResponse struct {
	TbkDgVegasTljCreateResponse TaoBaoTaoLiJinResponseData   `json:"tbk_dg_vegas_tlj_create_response"`
	TbkTpwdCreateResponse       TaoBaoTaoKouLingResponseData `json:"tbk_tpwd_create_response"`
}

type TaoBaoTaoLiJinResponseData struct {
	Result    TaoBaoResponseResult `json:"result"`
	RequestId string               `json:"request_id"`
}

type TaoBaoResponseResult struct {
	Model   TaoBaoResponseResultModel `json:"model"`
	Success bool                      `json:"success"`
	MsgCode string                    `json:"msg_code"`
	MsgInfo string                    `json:"msg_info"`
}

type TaoBaoResponseResultModel struct {
	RightsId  string `json:"rights_id"`
	SendUrl   string `json:"send_url"`
	VegasCode string `json:"vegas_code"`
}

type TaoBaoTaoKouLingResponseData struct {
	Data TaoBaoTaoKouLingResponseModelData `json:"data"`
}

type TaoBaoTaoKouLingResponseModelData struct {
	Model string `json:"model"`
}

// Do do http request
func (tbc *TaoBaoClient) Do() (*TaoBaoApiResponse, error) {
	params := url.Values{}
	for k, v := range tbc.CommonParams {
		params.Set(k, v[0])
	}
	for k, v := range tbc.InputParams {
		params.Set(k, v[0])
	}
	params.Set("sign", tbc.Sign)

	reqURL := fmt.Sprintf("%s?%s", taobaoRouteUrlProd, params.Encode())
	resp, err := tbc.Client.Get(reqURL)
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
	klog.Infof("Get Taobao Response data : %s", string(data))

	respData := new(TaoBaoApiResponse)
	if err := json.Unmarshal(data, respData); err != nil {
		klog.Error(err)
		return nil, err
	}

	return respData, nil
}
