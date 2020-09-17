package config

import (
	"io/ioutil"
	"os"
	"path"
	"sync"

	"k8s.io/klog"
	"sigs.k8s.io/yaml"
)

const (
	defaultConfigFile = "./config.yaml"
)

var GlobalConfig *Config

const (
	DefaultInviteNum = 3
)

var Running bool
var runningLock *sync.Mutex

type Config struct {
	InviteMangerConf          *InviteManageConf
	LocalUser                 *LocalUserInfo
	SendReceiveConf           SendReceiveConf
	AutoRemoveConf            *AutoRemoveConf
	TaoLiJinConf              *TaoLiJinConf
	AutoAgreeFriendVerifyConf *AutoAgreeFriendVerifyConf
}

type LocalUserInfo struct {
	RobotWxid        string `json:"robot_wxid"`
	Wxid             string `json:"wxid"`
	WxNum            string `json:"wx_num"`
	Nickname         string `json:"nickname"`
	HeadUrl          string `json:"head_url"`
	HeadImgUrl       string `json:"headimgurl"`
	Signature        string `json:"signature"`
	BackgroundImgUrl string `json:"backgroundimgurl"`
	UpdateDesc       string `json:"update_desc"`
	Status           int    `json:"status"`
	WxHand           int64  `json:"wx_hand"`
	WxWindHandle     int64  `json:"wx_wind_handle"`
	Pid              int64  `json:"pid"`
	LoginTime        int64  `json:"login_time"`
}

type CommonUserInfo struct {
	Wxid      string `json:"Wxid"`
	Name      string `json:"Name"`
	RobotWxid string `json:"Robot_wxid"`
}

// 消息接收转发相关配置
// - keywords 包含指定关键词才转发
// - filterKeywords 包含指定关键词不转发
type SendReceiveConf struct {
	ReceiveFromGroup     ReceiveFromGroup
	SendToUsers          []CommonUserInfo
	Keywords             []string
	FilterKeywords       []string
	TranMoneySep         bool
	StartSendReceiver    bool
	SendInterval         int
	ActionInterval       int
	AutoAgreeGroupInvite bool
}

// AutoRemoveConf 自动踢人配置
type AutoRemoveConf struct {
	Start             bool
	SendLink          bool
	SendQRCode        bool
	SendVideo         bool
	SendVoice         bool
	SendCard          bool
	ShareLink         bool
	Applets           bool
	FilterWords       bool
	FilterWordsString string
	FilterNames       bool
	FilterNamesString string
	MsgLength         bool
	MaxMsgLength      int
}

// ReceiveFromGroup 定义监听群组map
type ReceiveFromGroup map[string]GroupUserInfo

type GroupUserInfo struct {
	GroupName string
	Users     []CommonUserInfo
}

type InviteManageConf struct {
	Start          bool
	WelcomeMsg     string
	AlertHours     int
	RemoveHours    int
	AlertTimeBegin int
	AlertTimeEnd   int
	ManageGroups   []CommonUserInfo
	ManageOwners   []CommonUserInfo
}

// TaoLiJinConf 淘礼金配置项
type TaoLiJinConf struct {
	Start    bool
	Interval int

	TBAppKey      string
	TBAppSecret   string
	TBAdzoneID    string
	TBTotalNum    string
	TBPerFaceRate int

	DTKAppKey    string
	DTKAppSecret string
	DTKRankType  string
	DTKCid       string
}

// AutoAgreeFriendVerifyConf 自动通过好友请求配置项
type AutoAgreeFriendVerifyConf struct {
	Start            bool
	WelcomeMsg       string
	AutoInviteGroups []CommonUserInfo
}

// LoadConfig 加载配置文件
func LoadConfig(filePath string) (*Config, error) {
	if len(filePath) == 0 {
		filePath = defaultConfigFile
	}
	conf := new(Config)
	_, err := os.Stat(filePath)
	if os.IsPermission(err) {
		klog.Error(err)
		return nil, err
	}
	if os.IsNotExist(err) {
		if err := os.MkdirAll(path.Dir(filePath), 0755); err != nil {
			klog.Error(err)
		}
		if _, err := os.Create(filePath); err != nil {
			klog.Error(err)
			return nil, err
		}
		return nil, nil
	}
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	if len(data) == 0 {
		return nil, nil
	}
	if err := yaml.Unmarshal(data, conf); err != nil {
		klog.Error(err)
		return nil, err
	}
	return conf, nil
}

// SaveConfig 保存配置文件
func SaveConfig(filePath string) error {
	if len(filePath) == 0 {
		filePath = defaultConfigFile
	}

	data, err := yaml.Marshal(GlobalConfig)
	if err != nil {
		return err
	}

	if len(data) == 0 {
		klog.Infof("no config data to save %s", filePath)
		return nil
	}

	if err := ioutil.WriteFile(filePath, data, 0755); err != nil {
		return err
	}
	return nil
}

func InitConfig() {
	conf, err := LoadConfig("")
	if err != nil {
		klog.Info("加载配置文件失败！")
		conf = new(Config)
	}
	if conf == nil {
		conf = new(Config)
		conf.InviteMangerConf = &InviteManageConf{
			ManageGroups: make([]CommonUserInfo, 0),
			ManageOwners: make([]CommonUserInfo, 0),
		}
		conf.LocalUser = new(LocalUserInfo)
		conf.SendReceiveConf = SendReceiveConf{
			ReceiveFromGroup: make(map[string]GroupUserInfo, 0),
			SendToUsers:      make([]CommonUserInfo, 0),
			Keywords:         make([]string, 0),
			FilterKeywords:   make([]string, 0),
		}
		conf.AutoRemoveConf = new(AutoRemoveConf)
		conf.TaoLiJinConf = new(TaoLiJinConf)
		conf.AutoAgreeFriendVerifyConf = &AutoAgreeFriendVerifyConf{
			AutoInviteGroups: make([]CommonUserInfo, 0),
		}
	}
	GlobalConfig = conf
}

// InitRunning 初始化 Running 和 runningLock
func InitRunning() {
	Running = false
	runningLock = new(sync.Mutex)
}

// Start 修改全局 Running 的值
func Start() {
	runningLock.Lock()
	defer runningLock.Unlock()
	Running = true
}

// Stop 修改全局 Running 的值
func Stop() {
	runningLock.Lock()
	defer runningLock.Unlock()
	Running = false
}

// GetRunning 获取 Running 状态
func GetRunning() bool {
	runningLock.Lock()
	defer runningLock.Unlock()
	return Running
}
