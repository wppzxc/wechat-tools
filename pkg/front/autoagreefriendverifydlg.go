package front

import (
	"fmt"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/wppzxc/wechat-tools/pkg/config"
	"github.com/wppzxc/wechat-tools/pkg/wechat"
	"k8s.io/klog"
)

type autoInviteGroupInfoModel struct {
	walk.ListModelBase
	items         []config.CommonUserInfo
	mainDlg       *walk.Dialog
	groupsListBox *walk.ListBox
}

func (a *autoInviteGroupInfoModel) ItemCount() int {
	return len(a.items)
}

func (a *autoInviteGroupInfoModel) Value(index int) interface{} {
	return a.items[index].Name
}

func (am *AutoAgreeFriendVerifyManager) showAutoInviteGoupDig() (int, error) {
	if config.GlobalConfig.LocalUser == nil {
		user, err := wechat.GetLocalUserInfo(0)
		if err != nil {
			return -1, err
		}
		config.GlobalConfig.LocalUser = user
	}
	if config.GlobalConfig.AutoAgreeFriendVerifyConf.AutoInviteGroups == nil {
		config.GlobalConfig.AutoAgreeFriendVerifyConf.AutoInviteGroups = make([]config.CommonUserInfo, 0)
	}

	groups, err := wechat.GetGroupList(config.GlobalConfig.LocalUser.Wxid)
	if err != nil {
		return -1, fmt.Errorf("获取群聊列表失败: %s", err)
	}

	if len(groups) == 0 {
		return -1, fmt.Errorf("获取群聊列表失败: 群聊列表数目为 0")
	}

	groupsModel := getAutoInviteGroupsListModel(groups)
	return Dialog{
		AssignTo: &groupsModel.mainDlg,
		Title:    "请选择群",
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
	}.Run(am.ParentWindow)
}

func getAutoInviteGroupsListModel(groups []config.CommonUserInfo) *autoInviteGroupInfoModel {
	model := new(autoInviteGroupInfoModel)
	for _, g := range groups {
		model.items = append(model.items, g)
	}
	model.mainDlg = new(walk.Dialog)
	model.groupsListBox = new(walk.ListBox)
	return model
}

func (a *autoInviteGroupInfoModel) CommitGroups() {
	defer a.mainDlg.Accept()
	groupIndexes := a.groupsListBox.SelectedIndexes()
	for _, index := range groupIndexes {
		config.GlobalConfig.AutoAgreeFriendVerifyConf.AutoInviteGroups = append(config.GlobalConfig.AutoAgreeFriendVerifyConf.AutoInviteGroups, a.items[index])
	}
}
