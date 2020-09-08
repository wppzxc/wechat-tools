package front

import (
	"fmt"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/wppzxc/wechat-tools/pkg/config"
	"github.com/wppzxc/wechat-tools/pkg/wechat"
	"k8s.io/klog"
)

type commonUserInfoModel struct {
	walk.ListModelBase
	items         []config.CommonUserInfo
	mainDlg       *walk.Dialog
	groupsListBox *walk.ListBox
}

func (c *commonUserInfoModel) ItemCount() int {
	return len(c.items)
}

func (c *commonUserInfoModel) Value(index int) interface{} {
	return c.items[index].Name
}

func (im *InviteManager) showManageGoupDig() (int, error) {
	if config.GlobalConfig.LocalUser == nil {
		user, err := wechat.GetLocalUserInfo(0)
		if err != nil {
			return -1, err
		}
		config.GlobalConfig.LocalUser = user
	}
	if config.GlobalConfig.InviteMangerConf.ManageGroups == nil {
		config.GlobalConfig.InviteMangerConf.ManageGroups = make([]config.CommonUserInfo, 0)
	}

	groups, err := wechat.GetGroupList(config.GlobalConfig.LocalUser.Wxid)
	if err != nil {
		return -1, fmt.Errorf("获取群聊列表失败: %s", err)
	}
	if len(groups) == 0 {
		return -1, fmt.Errorf("获取群聊列表失败: 群聊列表数目为 0")
	}
	groupsModel := getGroupsListModel(groups)
	return Dialog{
		AssignTo: &groupsModel.mainDlg,
		Title:    "请选择管理群",
		MinSize:  Size{Width: 400, Height: 600},
		Layout:   VBox{},
		Children: []Widget{
			ListBox{
				AssignTo:       &groupsModel.groupsListBox,
				Model:          groupsModel,
				OnCurrentIndexChanged: func() {
					klog.Infof("%+v", groupsModel.groupsListBox.SelectedIndexes())
				},
				MultiSelection: true,
			},
			PushButton{
				Text:      "确定",
				OnClicked: groupsModel.CommitGroups,
			},
		},
	}.Run(im.ParentWindow)
}

func getGroupsListModel(groups []config.CommonUserInfo) *commonUserInfoModel {
	model := new(commonUserInfoModel)
	for _, g := range groups {
		model.items = append(model.items, g)
	}
	model.mainDlg = new(walk.Dialog)
	model.groupsListBox = new(walk.ListBox)
	return model
}

func (c *commonUserInfoModel) CommitGroups() {
	defer c.mainDlg.Accept()
	groupIndexes := c.groupsListBox.SelectedIndexes()
	for _, index := range groupIndexes {
		config.GlobalConfig.InviteMangerConf.ManageGroups = append(config.GlobalConfig.InviteMangerConf.ManageGroups, c.items[index])
	}
}
