package front

import (
	"fmt"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/wppzxc/wechat-tools/pkg/config"
	"github.com/wppzxc/wechat-tools/pkg/utils"
	"github.com/wppzxc/wechat-tools/pkg/wechat"
	"k8s.io/klog"
)

var ListenGroups []config.CommonUserInfo
var localUser *config.LocalUserInfo

func init() {
	ListenGroups = make([]config.CommonUserInfo, 0)
	localUser = new(config.LocalUserInfo)
}

func GetCreateTaoLiJinPage(mw *walk.MainWindow) *SendReceiver {
	ct := &SendReceiver{
		ParentWindow:        mw,
		GroupUserLineEdit:   new(walk.TextEdit),
		UsersTextEdit:       new(walk.TextEdit),
		ShowGroupUserDlgBtn: new(walk.PushButton),
	}

	ct.MainPage = &TabPage{
		Title:  "基础配置",
		Layout: VBox{},
		DataBinder: DataBinder{
			AutoSubmit: true,
			DataSource: ct,
		},
		Children: []Widget{
			Composite{
				Layout: VBox{},
				Children: []Widget{
					HSpacer{},
					Composite{
						Layout: VBox{},
						Children: []Widget{
							Label{
								Text: "设置文案格式(淘口令用 %s 代替)",
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
										AssignTo: &ct.GroupUserLineEdit,
										Text:     Bind("GroupUsers"),
										ReadOnly: true,
									},
									Composite{
										Layout: VBox{},
										Children: []Widget{
											PushButton{
												Text:     "添加",
												AssignTo: &ct.ShowGroupUserDlgBtn,
												OnClicked: func() {
													if cmd, err := ct.showManageGoupDig(); err != nil {
														walk.MsgBox(ct.ParentWindow, "错误", err.Error(), walk.MsgBoxIconError)
													} else if cmd == walk.DlgCmdOK {
														klog.Infof("dlg choose ok! groupUserInfo is %+v", ListenGroups)
														groupUsersStr := utils.GetCommonUsersNameStr(ListenGroups)
														ct.GroupUserLineEdit.SetText(groupUsersStr)
													}
												},
											},
											PushButton{
												Text: "清除",
												OnClicked: func() {
													ListenGroups = make([]config.CommonUserInfo, 0)
													ct.GroupUserLineEdit.SetText("")
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

	return ct
}

func (sr *SendReceiver) showManageGoupDig() (int, error) {
	if localUser == nil {
		user, err := wechat.GetLocalUserInfo(0)
		if err != nil {
			return -1, err
		}
		localUser = user
	}

	groups, err := wechat.GetGroupList(localUser.Wxid)
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
				// 暂时禁止多选
				// MultiSelection: true,
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
