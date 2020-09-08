package front

import (
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/wppzxc/wechat-tools/pkg/config"
	"github.com/wppzxc/wechat-tools/pkg/utils"
	"k8s.io/klog"
	"path"
)

type SendReceiver struct {
	StartSendReceiver   bool
	ParentWindow        *walk.MainWindow
	MainPage            *TabPage
	GroupUserLineEdit   *walk.TextEdit
	GroupUsers          string
	UsersTextEdit       *walk.TextEdit
	Users               string
	ShowGroupUserDlgBtn *walk.PushButton
	ShowSendUserDlgBtn  *walk.PushButton
	Keywords            string
	FilterKeywords      string
	TranMoneySep        bool
	SendInterval        int
	ActionInterval      int
}

func GetSendReceiverPage(mw *walk.MainWindow) *SendReceiver {
	sr := &SendReceiver{
		ParentWindow:        mw,
		GroupUserLineEdit:   new(walk.TextEdit),
		UsersTextEdit:       new(walk.TextEdit),
		ShowGroupUserDlgBtn: new(walk.PushButton),
		ShowSendUserDlgBtn:  new(walk.PushButton),
		Keywords:            path.Join(config.GlobalConfig.SendReceiveConf.Keywords...),
		FilterKeywords:      path.Join(config.GlobalConfig.SendReceiveConf.FilterKeywords...),
		TranMoneySep:        config.GlobalConfig.SendReceiveConf.TranMoneySep,
		StartSendReceiver:   config.GlobalConfig.SendReceiveConf.StartSendReceiver,
		SendInterval:        config.GlobalConfig.SendReceiveConf.SendInterval,
		ActionInterval:      config.GlobalConfig.SendReceiveConf.ActionInterval,
	}

	sr.GroupUsers = utils.GetGroupUsersNameStr(config.GlobalConfig.SendReceiveConf.ReceiveFromGroup)
	sr.Users = utils.GetUsersNameStr(config.GlobalConfig.SendReceiveConf.SendToUsers)
	if config.GlobalConfig.SendReceiveConf.ReceiveFromGroup == nil {
		config.GlobalConfig.SendReceiveConf.ReceiveFromGroup = make(config.ReceiveFromGroup)
	}
	if config.GlobalConfig.SendReceiveConf.SendToUsers == nil {
		config.GlobalConfig.SendReceiveConf.SendToUsers = make([]config.CommonUserInfo, 0)
	}
	if config.GlobalConfig.SendReceiveConf.Keywords == nil {
		config.GlobalConfig.SendReceiveConf.Keywords = make([]string, 0)
	}
	if config.GlobalConfig.SendReceiveConf.FilterKeywords == nil {
		config.GlobalConfig.SendReceiveConf.FilterKeywords = make([]string, 0)
	}

	sr.MainPage = &TabPage{
		Title:  "基础配置",
		Layout: VBox{},
		DataBinder: DataBinder{
			AutoSubmit: true,
			DataSource: sr,
		},
		Children: []Widget{
			Composite{
				Layout: VBox{},
				Children: []Widget{
					HSpacer{},
					Composite{
						Layout: HBox{},
						Children: []Widget{
							CheckBox{
								Text:    "启动监听转发",
								Checked: Bind("StartSendReceiver"),
							},
							CheckBox{
								Text:    "开启 ￥、$ 转 ( （",
								Checked: Bind("TranMoneySep"),
							},
						},
					},
					Composite{
						Layout: HBox{},
						Children: []Widget{
							Label{
								Text: "设置消息发送间隔：毫秒",
							},
							NumberEdit{
								Value:    Bind("SendInterval", Range{0, 10000}),
								Suffix:   " /(max 10000ms)",
								Decimals: 0,
							},
							Label{
								Text: "设置动作处理间隔：毫秒",
							},
							NumberEdit{
								Value:    Bind("ActionInterval", Range{0, 10000}),
								Suffix:   " /(max 10000ms)",
								Decimals: 0,
							},
						},
					},
					Composite{
						Layout: VBox{},
						Children: []Widget{
							Label{
								Text: "指定监听群组",
							},
							Composite{
								Layout: HBox{},
								Children: []Widget{
									TextEdit{
										AssignTo: &sr.GroupUserLineEdit,
										Text:     Bind("GroupUsers"),
										ReadOnly: true,
									},
									Composite{
										Layout: VBox{},
										Children: []Widget{
											PushButton{
												Text:     "添加",
												AssignTo: &sr.ShowGroupUserDlgBtn,
												OnClicked: func() {
													if cmd, err := sr.showGroupsUserDlg(); err != nil {
														walk.MsgBox(sr.ParentWindow, "错误", err.Error(), walk.MsgBoxIconError)
													} else if cmd == walk.DlgCmdOK {
														klog.Infof("dlg choose ok! groupUserInfo is %+v", config.GlobalConfig.SendReceiveConf.ReceiveFromGroup)
														groupUsersStr := utils.GetGroupUsersNameStr(config.GlobalConfig.SendReceiveConf.ReceiveFromGroup)
														sr.GroupUserLineEdit.SetText(groupUsersStr)
													}
												},
											},
											PushButton{
												Text: "清除",
												OnClicked: func() {
													config.GlobalConfig.SendReceiveConf.ReceiveFromGroup = make(config.ReceiveFromGroup)
													sr.GroupUserLineEdit.SetText("")
												},
											},
										},
									},
								},
							},
						},
					},
					Composite{
						Layout: VBox{},
						Children: []Widget{
							Label{
								Text: "指定转发用户",
							},
							Composite{
								Layout: HBox{},
								Children: []Widget{
									TextEdit{
										AssignTo: &sr.UsersTextEdit,
										Text:     Bind("Users"),
										ReadOnly: true,
									},
									Composite{
										Layout: VBox{},
										Children: []Widget{
											PushButton{
												Text:     "选择转发用户",
												AssignTo: &sr.ShowSendUserDlgBtn,
												MaxSize:  Size{Width: 200, Height: 100},
												OnClicked: func() {
													if cmd, err := sr.showUsersDlg(); err != nil {
														walk.MsgBox(sr.ParentWindow, "错误", err.Error(), walk.MsgBoxIconError)
													} else if cmd == walk.DlgCmdOK {
														klog.Infof("dlg choose ok! users is %+v", config.GlobalConfig.SendReceiveConf.SendToUsers)
														userStr := utils.GetUsersNameStr(config.GlobalConfig.SendReceiveConf.SendToUsers)
														sr.UsersTextEdit.SetText(userStr)
													}
												},
											},
											PushButton{
												Text: "清除",
												OnClicked: func() {
													config.GlobalConfig.SendReceiveConf.SendToUsers = make([]config.CommonUserInfo, 0)
													sr.UsersTextEdit.SetText("")
												},
											},
										},
									},
								},
							},
						},
					},
					Composite{
						Layout: VBox{},
						Children: []Widget{
							TextLabel{
								Text: "指定关键词，用 / 分隔",
							},
							TextEdit{
								Text:    Bind("Keywords"),
								VScroll: true,
							},
						},
					},
					Composite{
						Layout: VBox{},
						Children: []Widget{
							TextLabel{
								Text: "屏蔽关键词，用 / 分隔",
							},
							TextEdit{
								Text:    Bind("FilterKeywords"),
								VScroll: true,
							},
						},
					},
				},
			},
		},
	}

	return sr
}
