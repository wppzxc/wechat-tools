package front

import (
	"fmt"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/wppzxc/wechat-tools/pkg/config"
	"github.com/wppzxc/wechat-tools/pkg/wechat"
	"k8s.io/klog"
)

type groupUserModel struct {
	walk.ListModelBase
	items               []config.CommonUserInfo
	mainDlg             *walk.Dialog
	groupsListBox       *walk.ListBox
	usersListBox        *walk.ListBox
	userItems           []config.CommonUserInfo
	localUser           *config.LocalUserInfo
	listenGroupUserInfo config.ReceiveFromGroup
}

func (m *groupUserModel) ItemCount() int {
	return len(m.items)
}

func (m *groupUserModel) Value(index int) interface{} {
	return m.items[index].Name
}

func (m *groupUserModel) groupIndexChanged() {
	index := m.groupsListBox.CurrentIndex()
	group := m.items[index]
	users, err := wechat.GetGroupUserList(m.localUser.Wxid, group.Wxid)
	if err != nil {
		klog.Error(err)
		walk.MsgBox(m.mainDlg, "错误", err.Error(), walk.MsgBoxIconError)
	}
	m.userItems = users
	userModel := new(groupUserModel)
	for _, u := range users {
		userModel.items = append(userModel.items, u)
	}
	if err := m.usersListBox.SetModel(userModel); err != nil {
		klog.Error(err)
		walk.MsgBox(m.mainDlg, "错误", err.Error(), walk.MsgBoxIconError)
	}
}

func (sr *SendReceiver) groupUsersCommit(gum *groupUserModel) {
	defer gum.mainDlg.Accept()
	groupIndex := gum.groupsListBox.CurrentIndex()
	group := gum.items[groupIndex]
	selectdGroupUserIndexes := gum.usersListBox.SelectedIndexes()
	if len(selectdGroupUserIndexes) == 0 {
		gum.listenGroupUserInfo = nil
		return
	}
	selectedGroupUsers := make([]config.CommonUserInfo, 0)
	for _, index := range selectdGroupUserIndexes {
		selectedGroupUsers = append(selectedGroupUsers, gum.userItems[index])
	}

	config.GlobalConfig.SendReceiveConf.ReceiveFromGroup[group.Wxid] = config.GroupUserInfo{
		Users:     selectedGroupUsers,
		GroupName: group.Name,
	}
}

func (sr *SendReceiver) showGroupsUserDlg() (int, error) {
	if config.GlobalConfig.LocalUser == nil {
		user, err := wechat.GetLocalUserInfo(0)
		if err != nil {
			return -1, err
		}
		config.GlobalConfig.LocalUser = user
	}
	if config.GlobalConfig.SendReceiveConf.ReceiveFromGroup == nil {
		config.GlobalConfig.SendReceiveConf.ReceiveFromGroup = make(config.ReceiveFromGroup)
	}

	groups, err := wechat.GetGroupList(config.GlobalConfig.LocalUser.Wxid)
	if err != nil {
		return -1, fmt.Errorf("获取群聊列表失败: %s", err)
	}
	if len(groups) == 0 {
		return -1, fmt.Errorf("获取群聊列表失败: 群聊列表数目为 0")
	}
	groupsModel := getGroupUserModel(groups, config.GlobalConfig.LocalUser)
	groupsModel.listenGroupUserInfo = config.GlobalConfig.SendReceiveConf.ReceiveFromGroup
	return Dialog{
		AssignTo: &groupsModel.mainDlg,
		Title:    "请选择群组和群成员",
		MinSize:  Size{Width: 400, Height: 600},
		Layout:   VBox{},
		Children: []Widget{
			HSplitter{
				Children: []Widget{
					ListBox{
						AssignTo:              &groupsModel.groupsListBox,
						Model:                 groupsModel,
						OnCurrentIndexChanged: groupsModel.groupIndexChanged,
					},
					ListBox{
						AssignTo:       &groupsModel.usersListBox,
						MultiSelection: true,
					},
				},
			},
			PushButton{
				Text: "确认",
				OnClicked: func() {
					sr.groupUsersCommit(groupsModel)
				},
			},
		},
	}.Run(sr.ParentWindow)
}

func getGroupUserModel(groups []config.CommonUserInfo, localUser *config.LocalUserInfo) *groupUserModel {
	model := new(groupUserModel)
	for _, g := range groups {
		model.items = append(model.items, g)
	}
	model.mainDlg = new(walk.Dialog)
	model.groupsListBox = new(walk.ListBox)
	model.usersListBox = new(walk.ListBox)
	model.localUser = localUser
	return model
}
