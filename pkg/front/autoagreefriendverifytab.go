package front

import (
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/wppzxc/wechat-tools/pkg/config"
	// "github.com/wppzxc/wechat-tools/pkg/utils"
	"k8s.io/klog"
)

// AutoAgreeFriendVerifyManager 是自动通过好友请求模块控制器
type AutoAgreeFriendVerifyManager struct {
	ParentWindow             *walk.MainWindow
	MainPage                 *TabPage
	AutoInviteGroupsTextEdit *walk.TextEdit
	AutoInviteGroupsStr      string
}

// GetAutoAgreeFriendVerifyManager 初始化tab界面
func GetAutoAgreeFriendVerifyManager(mw *walk.MainWindow) *AutoAgreeFriendVerifyManager {
	am := &AutoAgreeFriendVerifyManager{
		ParentWindow: mw,
		AutoInviteGroupsTextEdit: new(walk.TextEdit),
	}
	if config.GlobalConfig.AutoAgreeFriendVerifyConf == nil {
		config.GlobalConfig.AutoAgreeFriendVerifyConf = new(config.AutoAgreeFriendVerifyConf)
	}

	// am.AutoInviteGroupsStr = utils.GetUsersNameStr(config.GlobalConfig.AutoAgreeFriendVerifyConf.AutoInviteGroups)

	am.MainPage = &TabPage{
		Title: "自动通过好友请求",
		Layout: VBox{},
		DataBinder: DataBinder{
			AutoSubmit: true,
			DataSource: config.GlobalConfig.AutoAgreeFriendVerifyConf,
		},
		Children: []Widget{
			Composite{
				Layout: HBox{},
				Children: []Widget{
					Label{
						Text: "是否开启",
					},
					CheckBox{
						Checked: Bind("Start"),
					},
				},
			},
			Composite{
				Layout: VBox{},
				Children: []Widget{
					Label{
						Text: "设置欢迎语",
					},
					TextEdit{
						Text: Bind("WelcomeMsg"),
					},
					Label{
						Text: "指定自动邀请群组：",
					},
					Composite{
						DataBinder: DataBinder{
							AutoSubmit: true,
							DataSource: am,
						},
						Layout: HBox{},
						Children: []Widget{
							TextEdit{
								AssignTo: &am.AutoInviteGroupsTextEdit,
								Text: Bind("AutoInviteGroupsStr"),
								ReadOnly: true,
							},
							Composite{
								Layout: VBox{},
								Children: []Widget{
									PushButton{
										Text: "选择",
										OnClicked: func() {
											if cmd, err := am.showAutoInviteGoupDig(); err != nil {
												walk.MsgBox(am.ParentWindow, "错误", err.Error(), walk.MsgBoxIconError)
											} else if cmd == walk.DlgCmdOK {
												klog.Infof("dlg choose ok! manageGroups is %+v", config.GlobalConfig.AutoAgreeFriendVerifyConf.AutoInviteGroups)
												// groupsStr := utils.GetCommonUsersNameStr(config.GlobalConfig.AutoAgreeFriendVerifyConf.AutoInviteGroups)
												am.AutoInviteGroupsTextEdit.SetText("groupsStr")
											}
										},
									},
									PushButton{
										Text: "清空",
										OnClicked: func() {
											config.GlobalConfig.AutoAgreeFriendVerifyConf.AutoInviteGroups = make([]config.CommonUserInfo, 0)
											am.AutoInviteGroupsTextEdit.SetText("")
										},
									},
								},
							},
						},
					},
				},
			},
			Composite{

			},
		},
	}
	return am
}