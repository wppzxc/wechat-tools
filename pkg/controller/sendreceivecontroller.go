package controller

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/wppzxc/wechat-tools/pkg/config"
	"github.com/wppzxc/wechat-tools/pkg/database"
	"github.com/wppzxc/wechat-tools/pkg/log"
	wechatToolsPrometheus "github.com/wppzxc/wechat-tools/pkg/prometheus"
	"github.com/wppzxc/wechat-tools/pkg/types"
	"github.com/wppzxc/wechat-tools/pkg/utils"
	"github.com/wppzxc/wechat-tools/pkg/wechat"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
)

const (
	defaultServerAddr = ":8074"
)

const (
	defaultHTML = `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<title>wechat-tools</title>
</head>
<body>
 
<table border="1">
  <tr>
    <th>群id</th>
    <th>微信id</th>
    <th>微信昵称</th>
  </tr>
%s
</table>
 
</body>
</html>`

	defaultTableRaw = `<tr>
    <td>%s</td>
    <td>%s</td>
    <td>%s</td>
  </tr>`
)

// Controller 主控制器
type Controller struct {
	Echo              *echo.Echo
	DB                *gorm.DB
	workqueue         workqueue.RateLimitingInterface
	actionqueue       workqueue.RateLimitingInterface
	tbClient          *utils.TaoBaoClient
	dtkClient         *utils.DaTaoKeClient
	taolijinSendItems map[string]types.DaTaoKeItem
	taolijinRunHours  int
	taolijinNowHour   int
	agreeLock         *sync.Mutex
}

// NewController 创建controller
func NewController() *Controller {
	e := echo.New()
	e.HideBanner = true
	// add metrics
	p := prometheus.NewPrometheus("base_http_server", nil)
	p.Use(e)

	wechatToolsPrometheus.InitMetrics()

	e.Use(middleware.LoggerWithConfig(log.DefaultLogConfig))
	e.Use(middleware.Recover())

	// db := database.InitDB()

	controller := &Controller{
		Echo:              e,
		// DB:                db,
		taolijinSendItems: make(map[string]types.DaTaoKeItem, 0),
		taolijinRunHours:  0,
		taolijinNowHour:   time.Now().Hour(),
		agreeLock:         new(sync.Mutex),
	}
	controller.Echo = e
	// controller.DB = db

	controller.Echo.POST("/receive", controller.server)
	// controller.Echo.GET("/roles/:role/users", controller.getUsers)
	return controller
}

// StartServer 开启http服务接收可爱猫回调
func (ctl *Controller) StartServer() {
	klog.Fatal(ctl.Echo.Start(defaultServerAddr))
}

// StartMsgSendWork 开始消息转发
func (ctl *Controller) StartMsgSendWork(stopCh chan struct{}) {
	go func() {
		klog.Info("开始消息发送...")
		for {
			working := ctl.processMsg()
			if !working {
				klog.Warning("Error in msg send work, the workqueue has been shutdown!")
				return
			}
		}
	}()
	<-stopCh
	ctl.workqueue.ShutDown()
	klog.Info("停止消息发送！")
	return
}

// StartActionSendWork 开始处理微信动作，如：加好友，踢群成员等
func (ctl *Controller) StartActionSendWork(stopCh chan struct{}) {
	go func() {
		klog.Info("开始动作发送...")
		for {
			working := ctl.processAction()
			if !working {
				klog.Warning("Error in msg send work, the workqueue has been shutdown!")
				return
			}
		}
	}()
	<-stopCh
	ctl.actionqueue.ShutDown()
	klog.Info("停止动作发送！")
	return
}

func (ctl *Controller) processMsg() bool {
	obj, shutdown := ctl.workqueue.Get()
	if shutdown {
		return false
	}
	defer ctl.workqueue.Done(obj)
	defer ctl.workqueue.Forget(obj)

	sendParam, ok := obj.(types.SendParam)
	if !ok {
		klog.Warningf("Get sendParam from workqueue error : %+v", obj)
		return true
	}
	if err := wechat.Send(sendParam); err != nil {
		klog.Errorf("Error in send msg to wechat : %s", err)
		wechatToolsPrometheus.TransFailedMsgs.Inc()
	} else {
		klog.Infof("成功调用可爱猫接口, 发送消息: %+v", sendParam)
	}
	// 发送间隔
	time.Sleep(time.Duration(config.GlobalConfig.SendReceiveConf.SendInterval) * time.Millisecond)
	return true
}

func (ctl *Controller) processAction() bool {
	obj, shutdown := ctl.actionqueue.Get()
	if shutdown {
		return false
	}
	defer ctl.actionqueue.Done(obj)
	defer ctl.actionqueue.Forget(obj)

	sendParam, ok := obj.(types.SendParam)
	if !ok {
		klog.Warningf("Get sendParam from workqueue error : %+v", obj)
		return true
	}
	if err := wechat.Send(sendParam); err != nil {
		klog.Errorf("Error in send msg to wechat : %s", err)
		wechatToolsPrometheus.TransFailedMsgs.Inc()
	} else {
		klog.Infof("成功调用可爱猫接口, 发送消息: %+v", sendParam)
	}
	// 发送间隔
	time.Sleep(time.Duration(config.GlobalConfig.SendReceiveConf.ActionInterval) * time.Millisecond)
	return true
}

func (ctl *Controller) enqueueSendMsg(sendP types.SendParam) {
	ctl.workqueue.AddRateLimited(sendP)
}

func (ctl *Controller) enqueueSendAction(sendP types.SendParam) {
	ctl.actionqueue.AddRateLimited(sendP)
}

// SetNewWorkqueue 更新workqueue
func (ctl *Controller) SetNewWorkqueue(workqueue workqueue.RateLimitingInterface) {
	ctl.workqueue = workqueue
}

// SetNewActionqueue 更新actionqueue
func (ctl *Controller) SetNewActionqueue(workqueue workqueue.RateLimitingInterface) {
	ctl.actionqueue = workqueue
}

// InitSendReceiveDatabase 初始化消息转发模块的的数据
func (ctl *Controller) InitSendReceiveDatabase() error {
	for groupWxid := range config.GlobalConfig.SendReceiveConf.ReceiveFromGroup {
		_, err := database.GetGroupUserByWxid(groupWxid, config.GlobalConfig.LocalUser.Wxid)
		if err == nil {
			continue
		}
		if err := database.CreateUser(&database.User{
			GroupWxid:        groupWxid,
			NickName:         config.GlobalConfig.LocalUser.Nickname,
			Wxid:             config.GlobalConfig.LocalUser.Wxid,
			InviteUserNumber: 0,
			Role:             database.UserRoleOwner,
		}); err != nil {
			return err
		}
	}
	klog.Info("初始化数据成功")
	return nil
}

func (ctl *Controller) getUsers(c echo.Context) error {
	role := c.Param("role")
	allUsers := make([]database.User, 0)

	switch role {
	case database.UserRoleOwner:
		owners := config.GlobalConfig.InviteMangerConf.ManageOwners
		raws := ""
		for _, u := range owners {
			raw := fmt.Sprintf(defaultTableRaw, "-", u.Wxid, u.Name)
			raws = fmt.Sprintf("%s\n%s", raws, raw)
		}
		html := fmt.Sprintf(defaultHTML, raws)
		return c.HTML(http.StatusOK, html)
	case database.UserRoleManager:
		for _, g := range config.GlobalConfig.InviteMangerConf.ManageGroups {
			users, err := database.GetGroupUsersByRole(g.Wxid, database.UserRoleManager)
			if err != nil {
				klog.Error(err)
				return err
			}
			allUsers = append(allUsers, users...)
		}
		raws := ""
		for _, u := range allUsers {
			raw := fmt.Sprintf(defaultTableRaw, u.GroupWxid, u.Wxid, u.NickName)
			raws = fmt.Sprintf("%s\n%s", raws, raw)
		}
		html := fmt.Sprintf(defaultHTML, raws)
		return c.HTML(http.StatusOK, html)
	case database.UserRoleNormal:
		for _, g := range config.GlobalConfig.InviteMangerConf.ManageGroups {
			users, err := database.GetGroupUsersByRole(g.Wxid, database.UserRoleNormal)
			if err != nil {
				klog.Error(err)
				return err
			}
			allUsers = append(allUsers, users...)
		}
		raws := ""
		for _, u := range allUsers {
			raw := fmt.Sprintf(defaultTableRaw, u.GroupWxid, u.Wxid, u.NickName)
			raws = fmt.Sprintf("%s\n%s", raws, raw)
		}
		html := fmt.Sprintf(defaultHTML, raws)
		return c.HTML(http.StatusOK, html)
	case database.UserRoleBlack:
		bls, err := database.GetAllBlackLists()
		if err != nil {
			return err
		}
		raws := ""
		for _, u := range bls {
			raw := fmt.Sprintf(defaultTableRaw, "-", u.Wxid, "-")
			raws = fmt.Sprintf("%s\n%s", raws, raw)
		}
		html := fmt.Sprintf(defaultHTML, raws)
		return c.HTML(http.StatusOK, html)
	default:
		return fmt.Errorf("不支持的角色类型：%s", role)
	}
}
