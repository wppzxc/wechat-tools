package utils

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/wppzxc/wechat-tools/pkg/config"
	"github.com/wppzxc/wechat-tools/pkg/types"
	"io/ioutil"
	"k8s.io/klog"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

const (
	taobaoRouteUrlProd    = "http://gw.api.taobao.com/router/rest"
	taobaoRouteUrlSandBox = "http://gw.api.tbsandbox.com/router/rest"

	defaultTaoLiJinName          = "淘礼金来了"
	createTaoLiJinMethod         = "taobao.tbk.dg.vegas.tlj.create"
	defaultTaoLiJinMethodVersion = "2.0"
	defaultSignMethod            = "md5"

	defaultTaoKouLingText  = "淘礼金来了"
	createTaoKouLingMethod = "taobao.tbk.tpwd.create"

	defaultTaoLiJinTotalNum = "1"
)

type TaoBaoClient struct {
	CommonParams url.Values
	InputParams  url.Values
	Client       http.Client
	Sign         string
}

type ParamSlice [][]string

func (p ParamSlice) Len() int           { return len(p) }
func (p ParamSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p ParamSlice) Less(i, j int) bool { return p[i][0] < p[j][0] }

func NewTaoBaoClient() *TaoBaoClient {
	now := time.Now()
	startTime := now.Format("2006-01-02 15:04:05")
	//endTime := now.Add(24 * time.Hour).Format("2006-01-02 15:04:05")
	commonParams := url.Values{
		"app_key":     []string{config.GlobalConfig.TaoLiJinConf.TBAppKey},
		"sign_method": []string{defaultSignMethod},
		"timestamp":   []string{startTime},
		"v":           []string{defaultTaoLiJinMethodVersion},
		"format":      []string{"json"},
	}

	tbClient := TaoBaoClient{
		CommonParams: commonParams,
		InputParams:  nil,
		Client:       http.Client{},
	}
	return &tbClient
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

	taobaoResp, err := tbc.sign(config.GlobalConfig.TaoLiJinConf.TBAppSecret).Do()
	if err != nil {
		klog.Error(err)
		return "", err
	}

	return taobaoResp.TbkTpwdCreateResponse.Data.Model, nil
}

func (tbc *TaoBaoClient) CreateTaoLiJinUrl(itemId string, perFace string, name string) (string, error) {
	tbc.CommonParams.Set("method", createTaoLiJinMethod)
	if len(name) == 0 {
		name = defaultTaoLiJinName
	}
	now := time.Now()
	startTime := now.Format("2006-01-02 15:04:05")
	endTime := now.Add(24 * time.Hour).Format("2006-01-02 15:04:05")
	tbc.InputParams = url.Values{
		"adzone_id":                []string{config.GlobalConfig.TaoLiJinConf.TBAdzoneID},
		"item_id":                  []string{itemId},
		"total_num":                []string{config.GlobalConfig.TaoLiJinConf.TBTotalNum},
		"name":                     []string{name},
		"user_total_win_num_limit": []string{"1"},
		"security_switch":          []string{"false"},
		"per_face":                 []string{perFace},
		"send_start_time":          []string{startTime},
		"send_end_time":            []string{endTime},
	}

	taobaoResp, err := tbc.sign(config.GlobalConfig.TaoLiJinConf.TBAppSecret).Do()
	if err != nil {
		klog.Error(err)
		return "", err
	}

	if taobaoResp.TbkDgVegasTljCreateResponse.Result.Success {
		return taobaoResp.TbkDgVegasTljCreateResponse.Result.Model.SendUrl, nil
	}

	return "", fmt.Errorf("Error in create tao li jin '%+v'", taobaoResp)
}

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

// Do do http request
func (tbc *TaoBaoClient) Do() (*types.TaoBaoApiResponse, error) {
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

	respData := new(types.TaoBaoApiResponse)
	if err := json.Unmarshal(data, respData); err != nil {
		klog.Error(err)
		return nil, err
	}

	return respData, nil
}
