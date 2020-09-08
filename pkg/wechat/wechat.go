package wechat

import (
	"encoding/json"
	"fmt"
	"github.com/wppzxc/wechat-tools/pkg/config"
	"github.com/wppzxc/wechat-tools/pkg/types"
	"io/ioutil"
	"k8s.io/klog"
	"net/http"
	"strings"
)

const (
	sendTextMsg    = 100
	sendGroupAtMsg = 102
	sendImageMsg   = 103
	sendVideoMsg   = 104
	sendFileMsg    = 105
)

// Send 发送消息
func Send(sendP types.SendParam) error {
	jsonStr, err := json.Marshal(sendP)
	if err != nil {
		klog.Error(err)
		return err
	}
	resp, err := http.Post(types.DefaultRemoteEndPoint, "application/x-www-form-urlencoded", strings.NewReader(string(jsonStr)))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	klog.Infof("send msg %s ok : %s", jsonStr, string(data))
	return nil
}

// GetGroupList 获取群组列表
func GetGroupList(robotWxid string) ([]config.CommonUserInfo, error) {
	resp, err := http.Post(fmt.Sprintf("http://%s:%s/httpAPI", types.DefaultRemoteHost, types.DefaultRemotePort), "",
		strings.NewReader(fmt.Sprintf(`{"api": "%s", "robot_wxid": "%s", "is_refresh": true}`, types.GetGroupListApi, robotWxid)))
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
	respData := new(types.ResponseUserList)
	if err := json.Unmarshal(data, respData); err != nil {
		klog.Errorf("Error in get local user info : %s ", err)
		return nil, err
	}
	if len(respData.Data) == 0 {
		return nil, fmt.Errorf("获取群聊列表失败！")
	}
	return respData.Data, nil
}

// GetFriendsList 获取好友列表
func GetFriendsList(robotWxid string) ([]config.CommonUserInfo, error) {
	resp, err := http.Post(fmt.Sprintf("http://%s:%s/httpAPI", types.DefaultRemoteHost, types.DefaultRemotePort), "",
		strings.NewReader(fmt.Sprintf(`{"api": "%s", "robot_wxid": "%s", "is_refresh": true}`, types.GetFriendListApi, robotWxid)))
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
	respData := new(types.ResponseUserList)
	if err := json.Unmarshal(data, respData); err != nil {
		klog.Errorf("Error in get local user info : %s ", err)
		return nil, err
	}
	if len(respData.Data) == 0 {
		return nil, fmt.Errorf("获取登录用户失败，请重新登录！")
	}
	return respData.Data, nil
}

// GetGroupUserList 获取群成员列表
func GetGroupUserList(robotWxid string, groupWxid string) ([]config.CommonUserInfo, error) {
	resp, err := http.Post(fmt.Sprintf("http://%s:%s/httpAPI", types.DefaultRemoteHost, types.DefaultRemotePort), "",
		strings.NewReader(fmt.Sprintf(`{"api": "%s", "robot_wxid": "%s", "group_wxid": "%s", "is_refresh":true}`, types.GetGroupMemberList, robotWxid, groupWxid)))
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
	respData := new(types.ResponseUserList)
	if err := json.Unmarshal(data, respData); err != nil {
		klog.Errorf("Error in get local user info : %s ", err)
		return nil, err
	}
	if len(respData.Data) == 0 {
		return nil, fmt.Errorf("获取群成员列表失败！")
	}
	return respData.Data, nil
}

// GetLocalUserInfo 获取已登录信息
func GetLocalUserInfo(index int) (*config.LocalUserInfo, error) {
	resp, err := http.Post(fmt.Sprintf("http://%s:%s/httpAPI", types.DefaultRemoteHost, types.DefaultRemotePort), "",
		strings.NewReader(fmt.Sprintf(`{"api": "%s"}`, types.GetLoggedAccountListApi)))
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
	respData := new(types.ResponseLocalUser)
	if err := json.Unmarshal(data, respData); err != nil {
		klog.Errorf("Error in get local user info : %s ", err)
		return nil, err
	}
	user := make([]config.LocalUserInfo, 0)
	if err := json.Unmarshal([]byte(respData.Data), &user); err != nil {
		klog.Error(err)
		return nil, err
	}
	if len(user) == 0 {
		return nil, fmt.Errorf("获取登录用户失败，请重新登录！")
	}
	return &user[index], nil
}

type verifyMsg struct {
	ToWxid       string `json:"to_wxid"`
	ToName       string `json:"to_name"`
	MsgID        int64  `json:"msgid"`
	FromWxid     string `json:"from_wxid"`
	FromNickname string `json:"from_nickname"`
	V1           string `json:"v1"`
	V2           string `json:"v2"`
	Sex          int    `json:"sex"`
	FromContent  string `json:"from_content"`
	Headimgurl   string `json:"headimgurl"`
	Type         int    `json:"type"`
}

// AgreeFriendVerify 同意好友请求
func AgreeFriendVerify(robotWxid string, jsonMsg string) error {
	msgStr := strings.Replace(jsonMsg, "\"", "\\\"", -1)
	msgStr = fmt.Sprintf(`"%s"`, msgStr)
	jsonStr := fmt.Sprintf(`{"api": "%s", "robot_wxid": "%s", "json_msg": %s}`, types.AgreeFriendVerify, robotWxid, msgStr)
	klog.Infof("input is: %s", jsonStr)
	resp, err := http.Post(fmt.Sprintf("http://%s:%s/httpAPI", types.DefaultRemoteHost, types.DefaultRemotePort), "",
		strings.NewReader(jsonStr))
	if err != nil {
		klog.Error(err)
		return err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		klog.Error(err)
		return err
	}
	if resp.StatusCode != http.StatusOK {
		errMsg := fmt.Errorf("接收好友请求失败: '%s'", string(data))
		return errMsg
	}
	klog.Infof("同意好友请求: %s", jsonMsg)
	return nil
}

// InviteInGroup 邀请加入群聊
func InviteInGroup(robotWxid string, groupWxid string, friendWxid string) error {
	resp, err := http.Post(fmt.Sprintf("http://%s:%s/httpAPI", types.DefaultRemoteHost, types.DefaultRemotePort), "",
		strings.NewReader(fmt.Sprintf(`{"api": "%s", "robot_wxid": "%s", "group_wxid": "%s", "friend_wxid": "%s"}`, types.InviteInGroup, robotWxid, groupWxid, friendWxid)))
	if err != nil {
		klog.Error(err)
		return err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		klog.Error(err)
		return err
	}
	if resp.StatusCode != http.StatusOK {
		errMsg := fmt.Errorf("邀请加入群聊失败: '%s'", string(data))
		return errMsg
	}
	klog.Infof("邀请用户%s加入群聊%s", friendWxid, groupWxid)
	return nil
}