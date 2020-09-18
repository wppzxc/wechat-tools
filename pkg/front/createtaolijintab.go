package front

import (
	"fmt"
	"path"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/wppzxc/wechat-tools/pkg/config"
	"github.com/wppzxc/wechat-tools/pkg/wechat"
	"k8s.io/klog"
)

// apikey 30029015
// apisecret b9f71d954e6a331a160c6c33956f7c44
// adzoneid 110409800054

var ListenGroups []config.CommonUserInfo
var LocalUser *config.LocalUserInfo
var Ct *SendReceiver

func init() {
	ListenGroups = make([]config.CommonUserInfo, 0)
	LocalUser = new(config.LocalUserInfo)
}

func GetCreateTaoLiJinPage(mw *walk.MainWindow) *SendReceiver {
	Ct = &SendReceiver{
		ParentWindow:        mw,
		GroupUserLineEdit:   new(walk.TextEdit),
		UsersTextEdit:       new(walk.TextEdit),
		ShowGroupUserDlgBtn: new(walk.PushButton),
		TaoBaoApiKey:        "30029015",
		TaoBaoApiSecret:     "b9f71d954e6a331a160c6c33956f7c44",
		TaoBaoAdZoneID:      "110409800054",
		Keywords: `来了
		_D_
		淘口令来了
		_T_
		实付金额_P_
		拍后发已拍`,
		DataokeApiKey:    "5e9d2dbadc286",
		DataokeApiSecret: "8f3c81484fdf7bd2695ddbbc6a128201",
	}

	Ct.MainPage = &TabPage{
		Title:  "淘礼金配置",
		Layout: VBox{},
		DataBinder: DataBinder{
			AutoSubmit: true,
			DataSource: Ct,
		},
		Children: []Widget{
			Composite{
				Layout: VBox{},
				Children: []Widget{
					HSpacer{},
					Composite{
						Layout: VBox{},
						Children: []Widget{
							Composite{
								Layout: HBox{},
								Children: []Widget{
									Label{
										Text: "大淘客API KEY",
									},
									LineEdit{
										Text: Bind("DataokeApiKey"),
									},
								},
							},
							Composite{
								Layout: HBox{},
								Children: []Widget{
									Label{
										Text: "大淘客API SECRET",
									},
									LineEdit{
										Text: Bind("DataokeApiSecret"),
									},
								},
							},
							Composite{
								Layout: HBox{},
								Children: []Widget{
									Label{
										Text: "淘宝API KEY",
									},
									LineEdit{
										Text: Bind("TaoBaoApiKey"),
									},
								},
							},
							Composite{
								Layout: HBox{},
								Children: []Widget{
									Label{
										Text: "淘宝API SECRET",
									},
									LineEdit{
										Text: Bind("TaoBaoApiSecret"),
									},
								},
							},
							Composite{
								Layout: HBox{},
								Children: []Widget{
									Label{
										Text: "淘宝API AdzoneID",
									},
									LineEdit{
										Text: Bind("TaoBaoAdZoneID"),
									},
								},
							},
						},
					},
					Composite{
						Layout: VBox{},
						Children: []Widget{
							Label{
								Text: "设置文案格式(短标题用 _D_ 代替，淘口令用 _T_ 代替，购买价用_P_代替)",
							},
							TextEdit{
								// 借用 keywords 来保存文案格式
								Text: Bind("Keywords"),
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
										AssignTo: &Ct.GroupUserLineEdit,
										Text:     Bind("GroupUsers"),
										ReadOnly: true,
									},
									Composite{
										Layout: VBox{},
										Children: []Widget{
											PushButton{
												Text:     "添加",
												AssignTo: &Ct.ShowGroupUserDlgBtn,
												OnClicked: func() {
													if cmd, err := Ct.showManageGoupDig(); err != nil {
														walk.MsgBox(Ct.ParentWindow, "错误", err.Error(), walk.MsgBoxIconError)
													} else if cmd == walk.DlgCmdOK {
														klog.Infof("dlg choose ok! groupUserInfo is %+v", ListenGroups)
														groupUsersStr := GetCommonUsersNameStr(ListenGroups)
														Ct.GroupUserLineEdit.SetText(groupUsersStr)
													}
												},
											},
											PushButton{
												Text: "清除",
												OnClicked: func() {
													ListenGroups = make([]config.CommonUserInfo, 0)
													Ct.GroupUserLineEdit.SetText("")
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	return Ct
}

func (sr *SendReceiver) showManageGoupDig() (int, error) {
	if len(LocalUser.RobotWxid) == 0 {
		user, err := wechat.GetLocalUserInfo(0)
		if err != nil {
			return -1, err
		}
		LocalUser = user
	}

	groups, err := wechat.GetGroupList(LocalUser.Wxid)
	if err != nil {
		return -1, fmt.Errorf("获取群聊列表失败: %s", err)
	}
	if len(groups) == 0 {
		return -1, fmt.Errorf("获取群聊列表失败: 群聊列表数目为 0")
	}
	groupsModel := getGroupsListModel(groups)
	return Dialog{
		AssignTo: &groupsModel.mainDlg,
		Title:    "请选择群",
		MinSize:  Size{Width: 400, Height: 600},
		Layout:   VBox{},
		Children: []Widget{
			ListBox{
				AssignTo: &groupsModel.groupsListBox,
				Model:    groupsModel,
				OnCurrentIndexChanged: func() {
					klog.Infof("%+v", groupsModel.groupsListBox.SelectedIndexes())
				},
				MultiSelection: true,
			},
			PushButton{
				Text:      "确定",
				OnClicked: groupsModel.localCommitGroups,
			},
		},
	}.Run(sr.ParentWindow)
}

func (c *commonUserInfoModel) localCommitGroups() {
	defer c.mainDlg.Accept()
	groupIndexes := c.groupsListBox.SelectedIndexes()
	for _, index := range groupIndexes {
		ListenGroups = append(ListenGroups, c.items[index])
	}
}

// GetCommonUsersNameStr 根据群组列表获取群组名字字符串
func GetCommonUsersNameStr(groups []config.CommonUserInfo) string {
	result := ""
	for _, g := range groups {
		result = path.Join(result, fmt.Sprintf("%s(%s)", g.Name, g.Wxid))
	}
	return result
}
