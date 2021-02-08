package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"k8s.io/klog"
)

type FormData struct {
	WxInfos []WxInfo `json:"wxInfos"`
	Titles  []string `json:"titles"`
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	RespError
}

type WxInfo struct {
	WxAppID     string `json:"wxAppID"`
	WxAppSecret string `json:"wxAppSecret"`
}

type RespError struct {
	ErrCode int64  `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

type BatchMaterial struct {
	Item []MetarialItem `json:"item"`
	RespError
}

type MetarialItem struct {
	MediaID string   `json:"media_id"`
	Content NewsItem `json:"content"`
}

type NewsItem struct {
	NewsItem []Article `json:"news_item"`
	RespError
}

type WxMaterial struct {
	Articles []Article `json:"articles"`
}

type WxImageResp struct {
	MediaID string `json:"media_id"`
	URL     string `json:"url"`
	RespError
}

type Article struct {
	Title            string `json:"title"`
	ThumbMediaID     string `json:"thumb_media_id"`
	Author           string `json:"author"`
	Digest           string `json:"digest"`
	ShowCoverPic     int    `json:"show_cover_pic"`
	Content          string `json:"content"`
	ContentSourceURL string `json:"content_source_url"`
}

type UpdateWxNews struct {
	MediaID  string  `json:"media_id"`
	Index    int     `json:"index"`
	Articles Article `json:"articles"`
}

func main() {
	klog.InitFlags(nil)
	flag.Set("log_file", "./main.log")
	flag.Set("log_file_max_size", "100")
	flag.Set("logtostderr", "false")
	flag.Set("alsologtostderr", "false")
	flag.Parse()

	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.PUT, echo.POST, echo.DELETE},
	}))

	e.POST("/upload", upload)
	e.POST("file", uploadFile)
	e.GET("/version", version)
	klog.Fatal(e.Start(":8080"))
}

func version(c echo.Context) error {
	return c.JSON(http.StatusOK, "v1.0.0")
}

func uploadFile(c echo.Context) error {
	file, err := c.FormFile("image.jpg")
	if err != nil {
		klog.Error(err)
		return c.JSON(http.StatusBadRequest, err)
	}
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create("./" + "image.jpg")
	if err != nil {
		klog.Error(err)
		return c.JSON(http.StatusInternalServerError, err)
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		klog.Error(err)
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, "文件上传成功！")
}

func upload(c echo.Context) error {
	formData := new(FormData)
	if err := c.Bind(formData); err != nil {
		klog.Errorf("Error in bind formdata", err)
		return c.JSON(http.StatusBadRequest, err)
	}

	body, _ := ioutil.ReadAll(c.Request().Body)
	defer c.Request().Body.Close()
	klog.Info(string(body))

	if len(formData.WxInfos) == 0 {
		return c.JSON(http.StatusBadRequest, "wxInfos can't be null")
	}

	if len(formData.Titles) == 0 {
		return c.JSON(http.StatusBadRequest, "titles can't be null")
	}

	var innerErrors string

	for _, wx := range formData.WxInfos {
		// 获取token
		token, err := getAccessToken(wx)
		if err != nil {
			klog.Error(err)
			innerErrors = fmt.Sprintf("%s\n %s get token error: %s", innerErrors, wx.WxAppID, err)
			continue
		}
		// 获取最近的图文素材
		news, err := getWxMaterial(token)
		if err != nil {
			klog.Error(err)
			innerErrors = fmt.Sprintf("%s\n %s get material error: %s", innerErrors, wx.WxAppID, err)
			continue
		}
		// 上传图片，并获取id
		imageResp, err := uploadWxImage(token)
		if err != nil {
			klog.Error(err)
			innerErrors = fmt.Sprintf("%s\n %s upload image error: %s", innerErrors, wx.WxAppID, err)
			continue
		}
		// 更新图文素材
		for i, article := range news.Content.NewsItem {
			updateItem := UpdateWxNews{
				MediaID: news.MediaID,
				Index:   i,
				Articles: Article{
					// TODO: 标题是8个循环用，还是所有的一起循环用
					Title:        formData.Titles[i],
					ThumbMediaID: imageResp.MediaID,
					Author:       article.Author,
					// Digest: article.Digest,
					Digest:       "摘要",
					ShowCoverPic: article.ShowCoverPic,
					Content:      article.Content,
					// ContentSourceURL: article.ContentSourceURL,
					ContentSourceURL: "http://hm90391r5790.jshdnb.com/jump?activity_id=702f6b88707173d7429693150a479a36b07d4",
				},
			}
			data, err := json.Marshal(&updateItem)
			if err != nil {
				klog.Error(err)
				innerErrors = fmt.Sprintf("%s\n %s inner error: %s", innerErrors, wx.WxAppID, err)
				break
			}
			resp, err := http.Post(fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/material/update_news?access_token=%s", token.AccessToken),
				"application/json",
				strings.NewReader(string(data)))
			if err != nil {
				klog.Error(err)
				innerErrors = fmt.Sprintf("%s\n %s update news error: %s", innerErrors, wx.WxAppID, err)
				break
			}

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				klog.Error(err)
				innerErrors = fmt.Sprintf("%s\n %s inner error: %s", innerErrors, wx.WxAppID, err)
				break
			}
			defer resp.Body.Close()

			re := new(RespError)
			if err := json.Unmarshal(body, re); err != nil {
				klog.Error(err)
				innerErrors = fmt.Sprintf("%s\n %s inner error: %s", innerErrors, wx.WxAppID, err)
				break
			}
			if re.ErrCode != 0 {
				klog.Error(re.ErrMsg)
				innerErrors = fmt.Sprintf("%s\n %s update news error: %s", innerErrors, wx.WxAppID, re.ErrMsg)
			}
			klog.Infof("更新%s 第%d标题成功", wx.WxAppID, i+1)
		}
	}
	if len(innerErrors) == 0 {
		return c.JSON(http.StatusOK, "ok")
	}

	return c.JSON(http.StatusOK, innerErrors)
}

func getAccessToken(wx WxInfo) (*TokenResponse, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s", wx.WxAppID, wx.WxAppSecret))
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	defer resp.Body.Close()

	token := new(TokenResponse)
	if err := json.Unmarshal(data, token); err != nil {
		klog.Error(err)
		return nil, err
	}

	return token, nil
}

func getWxMaterial(token *TokenResponse) (*MetarialItem, error) {
	resp, err := http.Post(fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/material/batchget_material?access_token=%s", token.AccessToken), "application/json", strings.NewReader(`{"type":"news","offset":0,"count":1}`))
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	batchMaterial := new(BatchMaterial)
	if err := json.Unmarshal(data, batchMaterial); err != nil {
		klog.Error(err)
		return nil, err
	}

	return &batchMaterial.Item[0], nil
}

func getMaterial(mediaID string, token *TokenResponse) (*NewsItem, error) {
	resp, err := http.Post(fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/material/get_material?access_token=%s", token.AccessToken), "application/json", strings.NewReader(fmt.Sprintf(`{"media_id":"%s"}`, mediaID)))
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	klog.Info("temp++++++++++++++++++++++++++++++++++++")
	klog.Info(string(data))

	item := new(NewsItem)
	if err := json.Unmarshal(data, item); err != nil {
		klog.Error(err)
		return nil, err
	}

	if item.ErrCode != 0 {
		klog.Errorf("err_code: %d, errmsg: %s", item.ErrCode, item.ErrMsg)
		return nil, errors.New(item.ErrMsg)
	}
	return item, nil
}

func uploadWxMaterial(wxm *WxMaterial, token *TokenResponse) error {
	data, err := json.Marshal(wxm)
	if err != nil {
		klog.Error(err)
		return err
	}
	resp, err := http.Post(fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/material/add_news?access_token=%s", token.AccessToken),
		"application/json", strings.NewReader(string(data)))
	if err != nil {
		klog.Error(err)
		return err
	}

	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		klog.Error(err)
		return err
	}
	klog.Info("上传成功: %s", string(respData))
	return nil
}

func uploadWxImage(token *TokenResponse) (*WxImageResp, error) {
	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/material/add_material?access_token=%s&type=image", token.AccessToken)
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	file, errFile1 := os.Open("./image.jpg")
	defer file.Close()
	part1, errFile1 := writer.CreateFormFile("media", filepath.Base("./image.jpg"))
	_, errFile1 = io.Copy(part1, file)
	if errFile1 != nil {
		klog.Error(errFile1)
		return nil, errFile1
	}
	err := writer.Close()
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPost, url, payload)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := client.Do(req)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	fmt.Println(string(body))
	wximresp := new(WxImageResp)
	if err := json.Unmarshal(body, wximresp); err != nil {
		klog.Error(err)
		return nil, err
	}
	return wximresp, nil
}
