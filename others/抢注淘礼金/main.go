package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	cron "github.com/robfig/cron/v3"
	"k8s.io/klog"
)

const (
	defaultTaoLiJinMethodVersion = "2.0"
	defaultSignMethod            = "md5"
	createTaoLiJinMethod         = "taobao.tbk.dg.vegas.tlj.create"
	defaultTaoLiJinName          = "淘礼金来了"
	taobaoRouteUrlProd           = "http://gw.api.taobao.com/router/rest"
	createTaoKouLingMethod       = "taobao.tbk.tpwd.create"
)

type config struct {
	TaobaoAppKey    string
	TaobaoAppSecret string
	TaobaoAdzongID  string
	ItemID          string
	Name            string
	PerFace         float32
	StartTime       string
	EndTime         string
	UseStartTime    string
	UseEndTime      string
	TotalNum        int
}

type TaoBaoClient struct {
	CommonParams url.Values
	InputParams  url.Values
	Client       *http.Client
	Sign         string
}

var (
	mainView         *walk.MainWindow
	options          *config
	taobaoClient     *TaoBaoClient
	startButton      *walk.PushButton
	successChan      chan string
	errorChan        chan string
	startTimeEdit    *walk.LineEdit
	endTimeEdit      *walk.LineEdit
	useStartTimeEdit *walk.LineEdit
	useEndTimeEdit   *walk.LineEdit
	resultEdit       *walk.LineEdit
	running          bool
	copyButton       *walk.PushButton
)

func main() {
	klog.InitFlags(nil)
	flag.Set("log_file", "./dataoke_qiangzhu.log")
	flag.Set("log_file_max_size", "100")
	flag.Set("logtostderr", "false")
	flag.Set("alsologtostderr", "false")
	flag.Parse()
	defer klog.Flush()

	errorChan = make(chan string)
	successChan = make(chan string)

	options = new(config)
	options.TaobaoAppKey = "30029015"
	options.TaobaoAppSecret = "b9f71d954e6a331a160c6c33956f7c44"
	options.TaobaoAdzongID = "110409800054"
	options.Name = "淘礼金来了"
	options.TotalNum = 1000

	taobaoClient = NewTaoBaoClient()

	if _, err := (MainWindow{
		Title:    "抢注淘礼金",
		AssignTo: &mainView,
		Size:     Size{Width: 400, Height: 300},
		Layout:   VBox{},
		DataBinder: DataBinder{
			AutoSubmit: true,
			DataSource: options,
		},
		Children: []Widget{
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
						Text: "商品ID",
					},
					LineEdit{
						Text: Bind("ItemID"),
					},
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					Label{
						Text: "淘礼金名称",
					},
					LineEdit{
						Text: Bind("Name"),
					},
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					PushButton{
						Text:      "今天",
						OnClicked: setToday,
					},
					PushButton{
						Text:      "明天",
						OnClicked: setTomorrow,
					},
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					Label{
						Text: "发放开始时间",
					},
					LineEdit{
						Text:     Bind("StartTime"),
						AssignTo: &startTimeEdit,
					},
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					Label{
						Text: "发放结束时间",
					},
					LineEdit{
						Text:     Bind("EndTime"),
						AssignTo: &endTimeEdit,
					},
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					Label{
						Text: "使用开始时间",
					},
					LineEdit{
						Text:     Bind("UseStartTime"),
						AssignTo: &useStartTimeEdit,
					},
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					Label{
						Text: "使用结束时间",
					},
					LineEdit{
						Text:     Bind("UseEndTime"),
						AssignTo: &useEndTimeEdit,
					},
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					Label{
						Text: "淘礼金面额",
					},
					NumberEdit{
						Value:    Bind("PerFace", Range{0, 99}),
						Decimals: 2,
					},
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					Label{
						Text: "淘礼金数量",
					},
					NumberEdit{
						Value: Bind("TotalNum", Range{0, 9999}),
					},
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					PushButton{
						Text:      "开始",
						OnClicked: startCronTask,
						AssignTo:  &startButton,
					},
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					Label{
						Text: "抢注结果：",
					},
					LineEdit{
						ReadOnly: true,
						AssignTo: &resultEdit,
					},
					PushButton{
						Text:      "复制",
						AssignTo:  &copyButton,
						OnClicked: copyResult,
					},
				},
			},
		},
	}).Run(); err != nil {
		panic(err)
	}
}

func setToday() {
	today := time.Now().Format("2006-01-02")
	todayStart := fmt.Sprintf("%s 00:00:00", today)
	todayEnd := fmt.Sprintf("%s 23:59:59", today)
	todayUseStart := today
	todayUseEnd := today

	startTimeEdit.SetText(todayStart)
	endTimeEdit.SetText(todayEnd)
	useStartTimeEdit.SetText(todayUseStart)
	useEndTimeEdit.SetText(todayUseEnd)

	options.StartTime = todayStart
	options.EndTime = todayEnd
	options.UseStartTime = todayUseStart
	options.UseEndTime = todayUseEnd
}

func setTomorrow() {
	tomorrow := time.Now().Add(24 * time.Hour).Format("2006-01-02")
	tomorrowStart := fmt.Sprintf("%s 00:00:00", tomorrow)
	tomorrowEnd := fmt.Sprintf("%s 23:59:59", tomorrow)
	tomorrowUseStart := tomorrow
	tomorrowUseEnd := tomorrow

	startTimeEdit.SetText(tomorrowStart)
	endTimeEdit.SetText(tomorrowEnd)
	useStartTimeEdit.SetText(tomorrowUseStart)
	useEndTimeEdit.SetText(tomorrowUseEnd)

	options.StartTime = tomorrowStart
	options.EndTime = tomorrowEnd
	options.UseStartTime = tomorrowUseStart
	options.UseEndTime = tomorrowUseEnd
}

func copyResult() {
	result := resultEdit.Text()
	clipboard.WriteAll(result)
}

func startCronTask() {
	if running {
		return
	}
	if len(options.ItemID) == 0 ||
		len(options.StartTime) == 0 ||
		len(options.EndTime) == 0 ||
		len(options.UseStartTime) == 0 ||
		len(options.UseEndTime) == 0 {
		walk.MsgBox(mainView, "错误", "请输入完整参数", walk.MsgBoxIconError)
		return
	}
	go func() {
		crontab := cron.New()
		task := func() {
			klog.Infof("任务开始时间: %s", time.Now().Format("2006-01-02 15:04:05.000000000"))
			num, url, err := createTaoLiJinWithSevenTimes()
			if err != nil {
				errorChan <- err.Error()
				return
			}
			klog.Infof("创建淘礼金成功: %s", url)
			tkl, err := taobaoClient.CreateTaoKouLing(options.Name, url)
			if err != nil {
				errorChan <- fmt.Errorf("创建淘口令失败! 淘礼金url: %s, Error: %s", url, err.Error()).Error()
				return
			}
			klog.Infof("创建淘口令成功: %s", tkl)
			result := fmt.Sprintf("数量：%d, 淘口令：%s", num, tkl)
			successChan <- result
			return
		}
		crontab.AddFunc("28 14 * * *", task)
		crontab.Start()
		select {
		case errMsg := <-errorChan:
			klog.Errorf("抢注淘礼金失败: %s", errMsg)
			walk.MsgBox(mainView, "错误", errMsg, walk.MsgBoxIconError)
			os.Exit(0)
		case result := <-successChan:
			klog.Infof("抢注淘礼金成功: %s", result)
			resultEdit.SetText(result)
		}
		crontab.Stop()
		// os.Exit(0)
	}()

	running = true
	startButton.SetEnabled(false)
}

func NewTaoBaoClient() *TaoBaoClient {
	now := time.Now()
	startTime := now.Format("2006-01-02 15:04:05")
	commonParams := url.Values{
		"app_key":     []string{options.TaobaoAppKey},
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

func createTaoLiJinWithSevenTimes() (int, string, error) {
	var url string
	var err error
	var isRetry bool
	var double bool
	for i := options.TotalNum; i >= 10; i = i / 2 {
		klog.Infof("创建淘礼金数量: %d，开始时间：%s", i, time.Now().Format("2006-01-02 15:04:05.000000000"))
		isRetry, double, url, err = createTaoLiJinUrlWithRetry()
		klog.Infof("创建淘礼金数量: %d，结束时间：%s", i, time.Now().Format("2006-01-02 15:04:05.000000000"))
		if err != nil {
			klog.Infof("创建结果: err: %s, url: %s", err.Error(), url)
			err = fmt.Errorf("总数量：%d创建淘礼金失败：%s\n ", i, err)
			if isRetry {
				if double {
					i = i * 2
				}
				continue
			}
		}
		return i, url, nil
	}
	return 0, "", err
}

func createTaoLiJinUrlWithRetry() (bool, bool, string, error) {
	url, err := taobaoClient.CreateTaoLiJinUrl(options)
	if err != nil {

		if strings.Index(err.Error(), "该商品不支持创建淘礼金红包") >= 0 {
			return false, false, "", err
		}

		if strings.Index(err.Error(), "ISP流控") >= 0 {
			return true, true, "", err
		}

		if strings.Index(err.Error(), "今日该商品淘礼金创建数已超上限，请您明日再试") >= 0 {
			return true, false, "", err
		}

		return true, false, "", err
	}
	return false, false, url, nil
}

func (tbc *TaoBaoClient) CreateTaoLiJinUrl(opts *config) (string, error) {
	tbc.CommonParams.Set("method", createTaoLiJinMethod)
	if len(opts.Name) == 0 {
		opts.Name = defaultTaoLiJinName
	}
	tbc.InputParams = url.Values{
		"adzone_id":                []string{"110409800054"},
		"item_id":                  []string{opts.ItemID},
		"total_num":                []string{fmt.Sprintf("%d", opts.TotalNum)},
		"name":                     []string{opts.Name},
		"user_total_win_num_limit": []string{"1"},
		"security_switch":          []string{"false"},
		"per_face":                 []string{fmt.Sprintf("%.2f", opts.PerFace)},
		"send_start_time":          []string{opts.StartTime},
		"send_end_time":            []string{opts.EndTime},
		"use_end_time_mode":        []string{"2"},
		"use_start_time":           []string{opts.UseStartTime},
		"use_end_time":             []string{opts.UseEndTime},
	}

	data, taobaoResp, err := tbc.sign(options.TaobaoAppSecret).Do()
	if err != nil {
		klog.Error(err)
		return "", err
	}

	if taobaoResp.TbkDgVegasTljCreateResponse.Result.Success {
		return taobaoResp.TbkDgVegasTljCreateResponse.Result.Model.SendUrl, nil
	}

	return "", fmt.Errorf("Error in create tao li jin '%s'", string(data))
}

func (tbc *TaoBaoClient) CreateTaoKouLing(text string, tljUrl string) (string, error) {
	tbc.CommonParams.Set("method", createTaoKouLingMethod)
	if len(text) < 5 {
		text = defaultTaoLiJinName
	}
	tbc.InputParams = url.Values{
		"text": []string{text},
		"url":  []string{tljUrl},
	}

	data, taobaoResp, err := tbc.sign(options.TaobaoAppSecret).Do()
	if err != nil {
		klog.Error(err)
		return "", err
	}

	if len(taobaoResp.TbkTpwdCreateResponse.Data.Model) == 0 {
		return "", fmt.Errorf("创建淘口令失败: %s", string(data))
	}

	return taobaoResp.TbkTpwdCreateResponse.Data.Model, nil
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
func (tbc *TaoBaoClient) Do() ([]byte, *TaoBaoApiResponse, error) {
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
		return nil, nil, err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		klog.Error(err)
		return nil, nil, err
	}
	klog.Infof("Get Taobao Response data : %s", string(data))

	respData := new(TaoBaoApiResponse)
	if err := json.Unmarshal(data, respData); err != nil {
		klog.Error(err)
		return nil, nil, err
	}

	return data, respData, nil
}
