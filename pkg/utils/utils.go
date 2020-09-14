package utils

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/tuotoo/qrcode"
	"github.com/wppzxc/wechat-tools/pkg/config"
	"github.com/wppzxc/wechat-tools/pkg/database"
	"github.com/wppzxc/wechat-tools/pkg/types"
	"github.com/wppzxc/wechat-tools/pkg/wechat"
	"k8s.io/klog"
)

// IsAtMsg 判断是否是at消息
// func IsAtMsg(msg string) (atNickname string, atWxid string, sendMsg string, isAtMsg bool) {
// 	index := strings.Index(msg, "@at,nickname=")
// 	if index >= 0 {
// 		strs := strings.SplitN(msg, "  ", 2)
// 		kvs := strs[0][1 : len(strs[0])-1]
// 		params := strings.Split(kvs, ",")
// 		atNickname = strings.Split(params[1], "=")[1]
// 		atWxid = strings.Split(params[2], "=")[1]
// 		sendMsg = strs[1]
// 		return atNickname, atWxid, sendMsg, true
// 	}
// 	return "", "", "", false
// }

// IsAtMsg 判断是否是at消息
func IsAtMsg(msg string) (string, string, string, bool) {
	reg := regexp.MustCompile(`\[@at,nickname=.*\]`)
	matchStr := reg.FindString(msg)
	if len(matchStr) == 0 {
		return "", "", "", false
	}
	sendMsg := strings.TrimSpace(strings.Replace(msg, matchStr, "", -1))
	kvs := matchStr[1 : len(matchStr)-1]
	params := strings.Split(kvs, ",")
	atNickname := strings.Split(params[1], "=")[1]
	atWxid := strings.Split(params[2], "=")[1]
	return atNickname, atWxid, sendMsg, true
}

// TranMoneySep 转换 ￥/$ 为 ()
func TranMoneySep(str string) string {
	reg := regexp.MustCompile(`[a-zA-Z0-9]{11}`)
	matchStr := reg.FindString(str)
	if len(matchStr) == 0 {
		klog.Infof(`没有匹配到淘口令: "%s"`, str)
		return str
	}

	matchRune := []rune(matchStr)
	beginSep := matchRune[0]
	endSep := matchRune[len(matchRune)-1]

	resultRune := []rune(str)
	for idx, r := range resultRune {
		if r == beginSep && resultRune[idx+10] == endSep {
			// 替换开始字符
			if idx == 0 {
				tmpRune := make([]rune, 0)
				tmpRune = append(tmpRune, 40)
				resultRune = append(tmpRune, resultRune...)
			} else {
				resultRune[idx-1] = 40
			}

			// 替换结尾字符
			if idx+11 == len(resultRune) {
				resultRune = append(resultRune, 41)
			} else {
				resultRune[idx+11] = 40
			}
			break
		}
	}
	return string(resultRune)
}

// GetUsersNameStr 获取用户名字字符串
// return name1/name2/name3/name4
func GetUsersNameStr(users []config.CommonUserInfo) string {
	str := ""
	for _, u := range users {
		str = path.Join(str, u.Name)
	}
	return str
}

// GetGroupUsersNameStr 获取群组和用户字符串
// return groupName1 ==>> username1/username2 || groupName2 ==>> username1/username2
func GetGroupUsersNameStr(groupUsers config.ReceiveFromGroup) string {
	result := ""
	for _, groupUserInfo := range groupUsers {
		usersStr := ""
		for _, u := range groupUserInfo.Users {
			usersStr = path.Join(usersStr, u.Name)
		}
		result = result + fmt.Sprintf("%s ==>> %s || ", groupUserInfo.GroupName, usersStr)
	}
	return result
}

// GetCommonUsersNameStr 根据群组列表获取群组名字字符串
func GetCommonUsersNameStr(groups []config.CommonUserInfo) string {
	result := ""
	for _, g := range groups {
		result = path.Join(result, fmt.Sprintf("%s(%s)", g.Name, g.Wxid))
	}
	return result
}

// ImageMsgSendParam 发送图片消息
func ImageMsgSendParam(path string, toWxid string) types.SendParam {
	sendParam := types.SendParam{
		Api:       types.SendImageMsgApi,
		RobotWxid: config.GlobalConfig.LocalUser.Wxid,
		ToWxid:    toWxid,
		Path:      path,
	}
	return sendParam
}

// TextMsgSendParam 发送文字消息
func TextMsgSendParam(errMsg string, toWxid string) types.SendParam {
	sendParam := types.SendParam{
		Api:       types.SendTextMsgApi,
		Msg:       errMsg,
		RobotWxid: config.GlobalConfig.LocalUser.Wxid,
		ToWxid:    toWxid,
	}
	return sendParam
}

// AlertAtMsgSendParam 发送at消息
func AlertAtMsgSendParam(groupWxid string, atWxid string, atName string, msg string) types.SendParam {
	sendParam := types.SendParam{
		Api:        types.SendGroupMsgAndAtApi,
		RobotWxid:  config.GlobalConfig.LocalUser.Wxid,
		GroupWxid:  groupWxid,
		MemberWxid: atWxid,
		MemberName: atName,
		Msg:        msg,
	}
	return sendParam
}

// RemoveMsgSendParam 删除用户
func RemoveMsgSendParam(groupWxid string, memberWxid string) types.SendParam {
	sendParam := types.SendParam{
		Api:        types.RemoveGroupMember,
		MemberWxid: memberWxid,
		GroupWxid:  groupWxid,
		RobotWxid:  config.GlobalConfig.LocalUser.Wxid,
	}
	return sendParam
}

// UsersContain 判断是否包含用户
func UsersContain(strs []config.CommonUserInfo, wxid string) bool {
	for _, s := range strs {
		if s.Wxid == wxid {
			return true
		}
	}
	return false
}

// DataBaseUsersContain 判断是否包含用户
func DataBaseUsersContain(users []database.User, wxid string) bool {
	if users == nil {
		return false
	}
	for _, u := range users {
		if u.Wxid == wxid {
			return true
		}
	}
	return false
}

// CheckQeCode 判断图片是否包含二维码
func CheckQeCode(url string) bool {
	file, err := os.Open(url)
	if err != nil {
		klog.Errorf("打开图片'%s'失败: %s", url, err)
		return false
	}
	defer file.Close()
	qr, err := qrcode.Decode(file)
	if err != nil {
		klog.Errorf("解析图片'%s'是否为二维码失败: %s", url, err)
		return false
	}
	klog.Infof("解析二维码图片'%s'成功: %s", url, qr.Content)
	return true
}

// StringsIndex 判断是否包含字符串
func StringsIndex(strs []string, str string) bool {
	for _, s := range strs {
		if index := strings.Index(str, s); index >= 0 {
			return true
		}
	}
	return false
}

// SetLocalUserInfo 获取可爱猫登录微信账户信息
func SetLocalUserInfo() error {
	localUser, err := wechat.GetLocalUserInfo(0)
	if err != nil {
		return err
	}
	config.GlobalConfig.LocalUser = localUser
	return nil
}

// CheckKeaimao 检车可爱猫是否运行
func CheckKeaimao() error {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(types.DefaultRemoteHost, types.DefaultRemotePort), types.DefaultTimeout)
	if err != nil {
		return fmt.Errorf("检查可爱猫启动失败, 请重启可爱猫: %s", err)
	}
	defer conn.Close()
	return nil
}

// CheckWechat 检查微信是否运行
// 通过获取群组列表
// 微信运转正常的话，请求时间超不过1秒
// 微信运转异常，或者获取列表失败，则返回错误
func CheckWechat() error {
	status, err := GetWeChatProcessStatus("WeChat.exe")
	if err != nil {
		klog.Error(err)
		return err
	}

	if status == "Running" {
		return nil
	}

	errMsg := fmt.Sprintf("微信运行状态错误: %s", status)
	klog.Error(errMsg)
	return fmt.Errorf(errMsg)

}

// GetWeChatProcessStatus 获取微信进程状态
func GetWeChatProcessStatus(appName string) (string, error) {
	cmd := exec.Command("tasklist", "/V")
	output, err := cmd.CombinedOutput()
	if err != nil {
		klog.Error(err)
		return "", err
	}
	n := strings.Index(string(output), "System")
	if n == -1 {
		klog.Error("获取所有进程失败: 进程列表为空")
		return "", fmt.Errorf("获取所有进程失败: 进程列表为空")
	}

	data := string(output)[n:]
	Processes := strings.Split(data, "\n")
	for _, process := range Processes {
		if strings.Index(process, appName) >= 0 {
			klog.Infof("获取微信进程: %s", process)
			if strings.Index(process, "Running") <= 0 {
				return "Unkown", nil
			}
			return "Running", nil
		}
	}

	return "", fmt.Errorf("获取进程 %s 失败，未找到该进程! ", appName)
}

// CheckWechatTools 检查 wechat-tools 是否正常
func CheckWechatTools() error {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort("127.0.0.1", "8074"), types.DefaultTimeout)
	if err != nil {
		return fmt.Errorf("检查wechat-tools启动失败, 请重启wechat-tools: %s", err)
	}
	defer conn.Close()
	return nil
}

// DownloadGoodsImage 下载大淘客商品图片
func DownloadGoodsImage(item *types.DaTaoKeItem) (string, error) {
	resp, err := http.Get(item.MainPic)
	if err != nil {
		klog.Error(err)
		return "", err
	}
	defer resp.Body.Close()
	now := strconv.FormatInt(time.Now().Unix(), 10)
	img, err := os.Create(fmt.Sprintf("./tmp/image/%s.jpg", now))
	if err != nil {
		klog.Error(err)
		return "", err
	}
	io.Copy(img, resp.Body)
	imgPath, err := filepath.Abs(img.Name())
	if err != nil {
		klog.Error(err)
		return "", err
	}
	return imgPath, nil
}
