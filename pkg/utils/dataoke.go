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
)

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
	Client       http.Client
	ReqUrl       string
	Sign         string
}

func NewDataokeClient() *DaTaoKeClient {
	commonParams := url.Values{
		"appKey":  []string{config.GlobalConfig.TaoLiJinConf.DTKAppKey},
		"version": []string{defaultDataokeApiVersion},
	}
	dtkClient := DaTaoKeClient{
		CommonParams: commonParams,
		InputParams:  nil,
		Client:       http.Client{},
		Sign:         "",
	}
	return &dtkClient
}

func (dtk *DaTaoKeClient) GetRealTimeListItem() ([]types.DaTaoKeItem, error) {
	dtk.ReqUrl = defaultGetRankingListUrl
	dtk.InputParams = url.Values{
		"rankType": []string{RankTypeRealTimeList},
	}
	dataokeResp, err := dtk.sign(config.GlobalConfig.TaoLiJinConf.DTKAppSecret).do()
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	return dataokeResp.Data, nil
}

func (dtk *DaTaoKeClient) sign(secret string) *DaTaoKeClient {
	params := url.Values{}
	for k, v := range dtk.CommonParams {
		params.Set(k, v[0])
	}
	for k, v := range dtk.InputParams {
		params.Set(k, v[0])
	}

	str := ""
	str = fmt.Sprintf("%s&key=%s", params.Encode(), secret)

	h := md5.New()
	h.Write([]byte(str))
	sign := hex.EncodeToString(h.Sum(nil))
	dtk.Sign = sign
	return dtk
}

func (dtk *DaTaoKeClient) do() (*types.DaTaoKeResponse, error) {
	params := url.Values{}
	for k, v := range dtk.CommonParams {
		params.Set(k, v[0])
	}
	for k, v := range dtk.InputParams {
		params.Set(k, v[0])
	}
	params.Set("sign", dtk.Sign)
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

	respData := new(types.DaTaoKeResponse)
	if err := json.Unmarshal(data, respData); err != nil {
		klog.Error(err)
		return nil, err
	}

	return respData, nil
}
