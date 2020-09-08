package front

import (
	"fmt"
	. "github.com/lxn/walk/declarative"
	"github.com/wppzxc/wechat-tools/pkg/config"
	"github.com/wppzxc/wechat-tools/pkg/wechat"
	"k8s.io/klog"
)

func (im *InviteManager) showManageOwnerDig() (int, error) {
	if config.GlobalConfig.LocalUser == nil {
		user, err := wechat.GetLocalUserInfo(0)
		if err != nil {
			return -1, err
		}
		config.GlobalConfig.LocalUser = user
	}
	if config.GlobalConfig.InviteMangerConf.ManageOwners == nil {
		config.GlobalConfig.InviteMangerConf.ManageOwners = make([]config.CommonUserInfo, 0)
	}

	users, err := wechat.GetFriendsList(config.GlobalConfig.LocalUser.Wxid)
	if err != nil {
		return -1, fmt.Errorf("获取好友列表失败: %s", err)
	}
	if len(users) == 0 {
		return -1, fmt.Errorf("获取好友列表失败: 群聊列表数目为 0")
	}
	usersModel := getGroupsListModel(users)
	return Dialog{
		AssignTo: &usersModel.mainDlg,
		Title:    "请选择管理群",
		MinSize:  Size{400, 600},
		Layout:   VBox{},
		Children: []Widget{
			ListBox{
				AssignTo:       &usersModel.groupsListBox,
				Model:          usersModel,
				OnCurrentIndexChanged: func() {
					klog.Infof("%+v", usersModel.groupsListBox.SelectedIndexes())
				},
				MultiSelection: true,
			},
			PushButton{
				Text:      "确定",
				OnClicked: usersModel.CommitOwners,
			},
		},
	}.Run(im.ParentWindow)
}

func (c *commonUserInfoModel) CommitOwners() {
	defer c.mainDlg.Accept()
	groupIndexes := c.groupsListBox.SelectedIndexes()
	for _, index := range groupIndexes {
		config.GlobalConfig.InviteMangerConf.ManageOwners = append(config.GlobalConfig.InviteMangerConf.ManageOwners, c.items[index])
	}
}

