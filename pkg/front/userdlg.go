package front

import (
	"fmt"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/wppzxc/wechat-tools/pkg/config"
	"github.com/wppzxc/wechat-tools/pkg/wechat"
)

type userModel struct {
	walk.ListModelBase
	items        []config.CommonUserInfo
	mainDlg      *walk.Dialog
	usersListBox *walk.ListBox
}

type ListBoxSelectUsers struct {
	usersString string
}

func (u *userModel) ItemCount() int {
	return len(u.items)
}

func (u *userModel) Value(index int) interface{} {
	return u.items[index].Name
}

func (sr *SendReceiver) usersListCommit(um *userModel) {
	defer um.mainDlg.Accept()
	selectedUserIndexes := um.usersListBox.SelectedIndexes()
	if len(selectedUserIndexes) == 0 {
		config.GlobalConfig.SendReceiveConf.SendToUsers = make([]config.CommonUserInfo, 0)
		return
	}
	for _, index := range selectedUserIndexes {
		config.GlobalConfig.SendReceiveConf.SendToUsers = append(config.GlobalConfig.SendReceiveConf.SendToUsers, um.items[index])
	}
	return
}

func (sr *SendReceiver) showUsersDlg() (int, error) {
	if config.GlobalConfig.LocalUser == nil {
		user, err := wechat.GetLocalUserInfo(0)
		if err != nil {
			return -1, err
		}
		config.GlobalConfig.LocalUser = user
	}
	groups, err := wechat.GetGroupList(config.GlobalConfig.LocalUser.Wxid)
	if err != nil {
		return -1, fmt.Errorf("获取群聊列表失败: %s", err)
	}
	if len(groups) == 0 {
		return -1, fmt.Errorf("获取群聊列表失败: 群聊列表数目为 0")
	}
	friends, err := wechat.GetFriendsList(config.GlobalConfig.LocalUser.Wxid)
	if err != nil {
		return -1, fmt.Errorf("获取好友列表失败: %s", err)
	}
	allUsers := append(groups, friends...)
	usersModel := getUsersListModel(allUsers)
	return Dialog{
		AssignTo: &usersModel.mainDlg,
		Title:    "请选择要转发的用户",
		MinSize:  Size{400, 600},
		Layout:   VBox{},
		Children: []Widget{
			ListBox{
				AssignTo:       &usersModel.usersListBox,
				Model:          usersModel,
				MultiSelection: true,
			},
			PushButton{
				Text: "确认",
				OnClicked: func() {
					sr.usersListCommit(usersModel)
				},
			},
		},
	}.Run(sr.ParentWindow)

}

func getUsersListModel(users []config.CommonUserInfo) *userModel {
	model := new(userModel)
	model.items = users
	model.mainDlg = new(walk.Dialog)
	return model
}
