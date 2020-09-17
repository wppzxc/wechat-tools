package main

import (
	"flag"
	"image"
	"os"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/wppzxc/wechat-tools/pkg/config"
	"github.com/wppzxc/wechat-tools/pkg/controller"
	_ "github.com/wppzxc/wechat-tools/pkg/database"
	"github.com/wppzxc/wechat-tools/pkg/front"
	"github.com/wppzxc/wechat-tools/pkg/prometheus"
	"github.com/wppzxc/wechat-tools/pkg/utils"
	"github.com/wppzxc/wechat-tools/version"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
)

// 主界面图标
var icon = &image.Gray{
	Pix: []uint8{
		200, 129, 92, 99, 91, 84, 85, 89, 97, 92, 95, 93, 91, 94, 125, 204, 133, 78, 90, 86, 103, 131, 126, 108,
		85, 91, 92, 91, 97, 91, 85, 125, 88, 95, 106, 181, 221, 216, 223, 218, 182, 107, 90, 99, 91, 90, 95, 89,
		90, 107, 208, 212, 190, 215, 213, 191, 212, 209, 103, 86, 92, 97, 95, 91, 88, 163, 232, 184, 121, 217, 211,
		125, 170, 235, 171, 86, 85, 93, 92, 90, 93, 193, 225, 211, 205, 212, 224, 210, 203, 204, 204, 152, 121,
		91, 87, 98, 88, 188, 220, 220, 219, 216, 212, 203, 214, 220, 214, 235, 228, 173, 93, 90, 86, 141, 230,
		211, 211, 218, 204, 216, 188, 169, 227, 161, 187, 235, 154, 83, 93, 87, 167, 221, 222, 221, 211, 215, 194,
		191, 221, 178, 201, 229, 185, 88, 96, 89, 106, 203, 185, 186, 197, 217, 216, 219, 214, 223, 216, 216, 190,
		85, 93, 88, 111, 109, 84, 101, 95, 197, 229, 213, 218, 213, 219, 221, 127, 83, 93, 92, 96, 88, 95, 95,
		90, 102, 179, 211, 217, 224, 221, 154, 87, 97, 91, 93, 91, 91, 91, 95, 97, 91, 89, 121, 132, 121, 129, 121,
		96, 88, 89, 95, 95, 94, 93, 90, 96, 97, 96, 88, 86, 87, 87, 93, 85, 96, 126, 82, 91, 94, 92, 94, 104,
		88, 94, 93, 96, 99, 98, 96, 84, 119, 198, 124, 85, 88, 91, 86, 85, 90, 90, 89, 87, 88, 87, 85, 125, 199},
	Stride: 16,
	Rect:   image.Rect(0, 0, 16, 16),
}

type mainView struct {
	mainView           *walk.MainWindow
	sendReceiveTab     *front.SendReceiver
	inviteManagerTab   *front.InviteManager
	taoLiJinTab        *front.TaoLiJin
	autoRemoverTab     *front.AutoRemover
	autoAgreeFriendTab *front.AutoAgreeFriendVerifyManager
	startBtn           *walk.PushButton
	stopBtn            *walk.PushButton
	stopCh             chan struct{}
	httpController     *controller.Controller
	createTaolijinTab  *front.SendReceiver
}

func main() {
	klog.InitFlags(nil)
	flag.Set("log_file", "./wechat-tools.log")
	flag.Set("log_file_max_size", "100")
	flag.Set("logtostderr", "false")
	flag.Set("alsologtostderr", "false")
	flag.Parse()
	defer func() {
		if r := recover(); r != nil {
			klog.Error("Recovered in f", r)
		}
		klog.Flush()
		os.Exit(1)
	}()

	config.InitRunning()
	mw := &mainView{
		mainView: new(walk.MainWindow),
		startBtn: new(walk.PushButton),
		stopBtn:  new(walk.PushButton),
	}

	mw.httpController = controller.NewController()
	// 启动http服务
	klog.Info("启动http服务")
	go func() {
		klog.Info(mw.httpController.Echo.Start(":8074"))
	}()

	// 启动 metrics 状态检查
	// go wait.Until(healthCheck, 60*time.Second, context.Background().Done())

	config.InitConfig()

	// mw.sendReceiveTab = front.GetSendReceiverPage(mw.mainView)
	// mw.inviteManagerTab = front.GetInviteManager(mw.mainView)
	// mw.taoLiJinTab = front.GetTaoLiJinPage(mw.mainView)
	// mw.autoRemoverTab = front.GetAutoRemoverPage(mw.mainView)
	// mw.autoAgreeFriendTab = front.GetAutoAgreeFriendVerifyManager(mw.mainView)
	mw.createTaolijinTab = front.GetCreateTaoLiJinPage(mw.mainView)

	icon, _ := walk.NewIconFromImageForDPI(icon, 96)
	if _, err := (MainWindow{
		Icon:     icon,
		AssignTo: &mw.mainView,
		Title:    getMainTitle(),
		Size:     Size{Width: 700, Height: 700},
		Layout:   VBox{},
		Children: []Widget{
			TabWidget{
				Pages: []TabPage{
					// *mw.sendReceiveTab.MainPage,
					// *mw.inviteManagerTab.MainPage,
					// *mw.autoAgreeFriendTab.MainPage,
					// *mw.taoLiJinTab.MainPage,
					// *mw.autoRemoverTab.MainPage,
					*mw.createTaolijinTab.MainPage,
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					PushButton{
						Text: "打开微信",
					},
					PushButton{
						Text:      "开始",
						OnClicked: mw.start,
						AssignTo:  &mw.startBtn,
					},
					PushButton{
						Text:      "停止",
						OnClicked: mw.stop,
						AssignTo:  &mw.stopBtn,
					},
				},
			},
		},
	}).Run(); err != nil {
		klog.Error(err)
	}
}

func getMainTitle() string {
	v := version.Get().String()
	return "微信工具 " + v
}

func (m *mainView) start() {
	klog.Info("wechat-tools 开始工作...")
	// 获取用户信息
	// klog.Info("获取登录信息")
	// if config.GlobalConfig.LocalUser == nil {
	// 	if err := utils.SetLocalUserInfo(); err != nil {
	// 		walk.MsgBox(m.mainView, "错误", err.Error(), walk.MsgBoxIconError)
	// 		return
	// 	}
	// }

	// 检查参数是否正确
	klog.Info("校验群消息转发启动参数")
	if err := m.sendReceiveTab.ValidateConfig(); err != nil {
		walk.MsgBox(m.mainView, "错误", err.Error(), walk.MsgBoxIconError)
		return
	}

	// 检查微信是否启动
	klog.Info("检查可爱猫微信是否启动")
	if err := utils.CheckKeaimao(); err != nil {
		walk.MsgBox(m.mainView, "错误", err.Error(), walk.MsgBoxIconError)
		return
	}

	// 初始化消息队列
	m.httpController.SetNewWorkqueue(workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()))
	// 初始化动作队列
	// m.httpController.SetNewActionqueue(workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()))

	// 初始化强制邀请数据数据，如：全局管理员和群成员
	// if err := front.InitInviteData(); err != nil {
	// 	walk.MsgBox(m.mainView, "错误", fmt.Sprintf("初始化群邀请数据失败： %s", err.Error()), walk.MsgBoxIconError)
	// 	return
	// }

	// 初始化数群监控数据数据
	// if err := m.httpController.InitSendReceiveDatabase(); err != nil {
	// 	walk.MsgBox(m.mainView, "错误", fmt.Sprintf("初始化群监控数据失败： %s", err.Error()), walk.MsgBoxIconError)
	// 	return
	// }

	// 初始话淘宝client
	controller.TaobaoClient = utils.NewTaoBaoClient()
	controller.DataokeClient = utils.NewDataokeClient()

	m.setUIEnable(false)
	m.stopCh = make(chan struct{})
	config.Start()

	// 开启发微信消息任务线程
	go m.httpController.StartMsgSendWork(m.stopCh)
	// 开启发微信动作任务线程
	// go m.httpController.StartActionSendWork(m.stopCh)

	// 开启强制邀请任务线程
	// if config.GlobalConfig.InviteMangerConf.Start {
	// 	go m.httpController.StartInviteManger(m.stopCh)
	// }

	// 开启淘礼金任务线程
	// if config.GlobalConfig.TaoLiJinConf.Start {
	// 	go m.httpController.StartTaoLiJinWorker(m.stopCh)
	// }

}

func (m *mainView) stop() {
	klog.Info("wechat-tools 停止工作...")
	if config.GetRunning() {
		config.Stop()
		m.setUIEnable(true)
		close(m.stopCh)
	}
}

func (m *mainView) setUIEnable(enable bool) {
	m.startBtn.SetEnabled(enable)
}

func healthCheck() {
	if !config.GetRunning() {
		prometheus.SendReceiveStatus.Set(0)
		prometheus.InviteManagerStatus.Set(0)
		prometheus.TaoLiJinStatus.Set(0)
		prometheus.AutoAgreeStatus.Set(0)
		prometheus.AutoRemoveStatus.Set(0)
	} else {
		if config.GlobalConfig.SendReceiveConf.StartSendReceiver {
			prometheus.SendReceiveStatus.Set(1)
		} else {
			prometheus.SendReceiveStatus.Set(0)
		}

		if config.GlobalConfig.InviteMangerConf.Start {
			prometheus.InviteManagerStatus.Set(1)
		} else {
			prometheus.InviteManagerStatus.Set(0)
		}

		if config.GlobalConfig.TaoLiJinConf.Start {
			prometheus.TaoLiJinStatus.Set(1)
		} else {
			prometheus.TaoLiJinStatus.Set(0)
		}

		if config.GlobalConfig.AutoAgreeFriendVerifyConf.Start {
			prometheus.AutoAgreeStatus.Set(1)
		} else {
			prometheus.AutoAgreeStatus.Set(0)
		}

		if config.GlobalConfig.AutoRemoveConf.Start {
			prometheus.AutoRemoveStatus.Set(1)
		} else {
			prometheus.AutoRemoveStatus.Set(0)
		}
	}

	// 检查可爱猫是否正常
	if err := utils.CheckKeaimao(); err != nil {
		klog.Error(err)
		prometheus.KeaimaoStatus.Set(0)
	} else {
		prometheus.KeaimaoStatus.Set(1)
	}

	// 检查微信是否正常
	if err := utils.CheckWechat(); err != nil {
		klog.Error(err)
		prometheus.WechatStatus.Set(0)
	} else {
		prometheus.WechatStatus.Set(1)
	}

	// 检查wechat-tools是否正常
	if err := utils.CheckWechatTools(); err != nil {
		klog.Error(err)
		prometheus.WechatToolsStatus.Set(0)
	} else {
		prometheus.WechatToolsStatus.Set(1)
	}
}
