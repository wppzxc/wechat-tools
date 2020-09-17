package front

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/jinzhu/gorm"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/wppzxc/wechat-tools/pkg/config"
	"github.com/wppzxc/wechat-tools/pkg/database"
	"github.com/wppzxc/wechat-tools/pkg/utils"
	"github.com/wppzxc/wechat-tools/pkg/wechat"
	"k8s.io/klog"
)

type InviteManager struct {
	ParentWindow         *walk.MainWindow
	MainPage             *TabPage
	ManageGroupsTextEdit *walk.TextEdit
	ManageOwnersTextEdit *walk.TextEdit
	ManageGroupsStr      string
	ManageOwnersStr      string
	Start                bool
	OwnerFilePathEdit    *walk.LineEdit
}

func GetInviteManager(mw *walk.MainWindow) *InviteManager {
	im := &InviteManager{
		ParentWindow:         mw,
		ManageGroupsTextEdit: new(walk.TextEdit),
	}

	if config.GlobalConfig.InviteMangerConf == nil {
		config.GlobalConfig.InviteMangerConf = new(config.InviteManageConf)
	}

	im.ManageGroupsStr = utils.GetUsersNameStr(config.GlobalConfig.InviteMangerConf.ManageGroups)
	im.ManageOwnersStr = utils.GetUsersNameStr(config.GlobalConfig.InviteMangerConf.ManageOwners)

	im.MainPage = &TabPage{
		Title:  "群裂变设置",
		Layout: VBox{},
		DataBinder: DataBinder{
			AutoSubmit: true,
			DataSource: config.GlobalConfig.InviteMangerConf,
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
						Text: "指定管理的群：",
					},
					Composite{
						DataBinder: DataBinder{
							AutoSubmit: true,
							DataSource: im,
						},
						Layout: HBox{},
						Children: []Widget{
							TextEdit{
								AssignTo: &im.ManageGroupsTextEdit,
								Text:     Bind("ManageGroupsStr"),
								ReadOnly: true,
							},
							Composite{
								Layout: VBox{},
								Children: []Widget{
									PushButton{
										Text: "选择",
										OnClicked: func() {
											if cmd, err := im.showManageGoupDig(); err != nil {
												walk.MsgBox(im.ParentWindow, "错误", err.Error(), walk.MsgBoxIconError)
											} else if cmd == walk.DlgCmdOK {
												klog.Infof("dlg choose ok! manageGroups is %+v", config.GlobalConfig.InviteMangerConf.ManageGroups)
												groupsStr := utils.GetCommonUsersNameStr(config.GlobalConfig.InviteMangerConf.ManageGroups)
												im.ManageGroupsTextEdit.SetText(groupsStr)
											}
										},
									},
									PushButton{
										Text:      "选择所有群",
										OnClicked: im.chooseAllGroups,
									},
									PushButton{
										Text: "清空",
										OnClicked: func() {
											config.GlobalConfig.InviteMangerConf.ManageGroups = make([]config.CommonUserInfo, 0)
											im.ManageGroupsTextEdit.SetText("")
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
						Text: "指定全局管理员：",
					},
					Composite{
						DataBinder: DataBinder{
							AutoSubmit: true,
							DataSource: im,
						},
						Layout: HBox{},
						Children: []Widget{
							TextEdit{
								AssignTo: &im.ManageOwnersTextEdit,
								Text:     Bind("ManageOwnersStr"),
								ReadOnly: true,
							},
							Composite{
								Layout: VBox{},
								Children: []Widget{
									PushButton{
										Text: "选择",
										OnClicked: func() {
											if cmd, err := im.showManageOwnerDig(); err != nil {
												walk.MsgBox(im.ParentWindow, "错误", err.Error(), walk.MsgBoxIconError)
											} else if cmd == walk.DlgCmdOK {
												klog.Infof("dlg choose ok! manageOwners is %+v", config.GlobalConfig.InviteMangerConf.ManageOwners)
												groupsStr := utils.GetCommonUsersNameStr(config.GlobalConfig.InviteMangerConf.ManageOwners)
												im.ManageOwnersTextEdit.SetText(groupsStr)
											}
										},
									},
									PushButton{
										Text: "清空",
										OnClicked: func() {
											config.GlobalConfig.InviteMangerConf.ManageOwners = make([]config.CommonUserInfo, 0)
											im.ManageOwnersTextEdit.SetText("")
										},
									},
								},
							},
							Composite{
								Layout: VBox{},
								Children: []Widget{
									LineEdit{
										AssignTo: &im.OwnerFilePathEdit,
										ReadOnly: true,
									},
									PushButton{
										Text:      "选择json文件",
										OnClicked: im.showChooseWhiteListFile,
									},
									PushButton{
										Text:      "导入全局管理员",
										OnClicked: im.importOwnerList,
									},
								},
							},
						},
					},
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					Label{
						Text: "设置提醒时间: 起",
					},
					NumberEdit{
						Value:    Bind("AlertTimeBegin", Range{0, 24}),
						Suffix:   " /(24)",
						Decimals: 0,
					},
					Label{
						Text: "设置提醒时间: 止",
					},
					NumberEdit{
						Value:    Bind("AlertTimeEnd", Range{0, 24}),
						Suffix:   " /(24)",
						Decimals: 0,
					},
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					Label{
						Text: "邀请提醒间隔：",
					},
					NumberEdit{
						Value:    Bind("AlertHours", Range{1, 999}),
						Suffix:   " 小时",
						Decimals: 0,
					},
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					Label{
						Text: "踢除群聊间隔：",
					},
					NumberEdit{
						Value:    Bind("RemoveHours", Range{1, 999}),
						Suffix:   " 小时",
						Decimals: 0,
					},
				},
			},
		},
	}
	return im
}

// InitInviteData 初始化管理员和已有群用户数据到 sqlite
func InitInviteData() error {
	if err := initManager(); err != nil {
		klog.Error(err)
		return err
	}
	if err := initExistUsers(); err != nil {
		klog.Error(err)
		return err
	}
	return nil
}

func initManager() error {
	klog.Info("开始初始化全局管理员...")
	managerGroups := config.GlobalConfig.InviteMangerConf.ManageGroups
	managerOwners := config.GlobalConfig.InviteMangerConf.ManageOwners
	mustExistUsers := make([]database.User, 0)
	for _, g := range managerGroups {
		for _, u := range managerOwners {
			mustExistUsers = append(mustExistUsers, database.User{
				GroupWxid: g.Wxid,
				NickName:  u.Name,
				Wxid:      u.Wxid,
				Role:      database.UserRoleOwner,
			})
		}
		// 机主也必须存在于管理的群中
		mustExistUsers = append(mustExistUsers, database.User{
			GroupWxid:        g.Wxid,
			NickName:         config.GlobalConfig.LocalUser.Nickname,
			Wxid:             config.GlobalConfig.LocalUser.Wxid,
			InviteUserNumber: 0,
			Alerted:          false,
			Role:             database.UserRoleOwner,
		})
	}
	for _, u := range mustExistUsers {
		user, err := database.GetGroupUserByWxid(u.GroupWxid, u.Wxid)
		// 如果用户不存在，则创建
		if err == gorm.ErrRecordNotFound {
			if err := database.CreateUser(&u); err != nil {
				return fmt.Errorf("初始化数据, 创建用户%s(%s)失败: %s", u.NickName, u.Wxid, err)
			}
			continue
		} else if err != nil {
			return fmt.Errorf("初始化数据, 获取用户信息失败: %s", err)
		}

		// 更新权限为全局管理员
		if user.Role != database.UserRoleOwner {
			user.Role = database.UserRoleOwner
			if err := database.UpdateGroupUserByWxid(*user, user.GroupWxid, user.Wxid); err != nil {
				return fmt.Errorf("初始化数据, 更新用户%s(%s)权限失败: %s", user.NickName, user.Wxid, err)
			}
		}

		// 删除黑名单
		_, err = database.GetBlackListByWxid(user.Wxid)
		if gorm.IsRecordNotFoundError(err) {
			continue
		} else if err != nil {
			return fmt.Errorf("初始化数据, 获取黑名单用户%s(%s)信息失败: %s", user.NickName, user.Wxid, err)
		}
		if err := database.DeleteBlackLists([]*database.User{user}); err != nil {
			return fmt.Errorf("初始化数据, 移除黑名单用户%s(%s)信息失败: %s", user.NickName, user.Wxid, err)
		}
	}
	klog.Info("初始化全局管理员完毕")
	return nil
}

func initExistUsers() error {
	existUsers := make([]*database.User, 0)
	for _, group := range config.GlobalConfig.InviteMangerConf.ManageGroups {
		users, err := wechat.GetGroupUserList(config.GlobalConfig.LocalUser.RobotWxid, group.Wxid)
		if err != nil {
			klog.Error(err)
			return err
		}
		// remove black list users and already in database users
		for _, u := range users {
			if _, err := database.GetBlackListByWxid(u.Wxid); err == nil {
				klog.Infof("用户 %s(%s) 已存在黑名单中，跳过", u.Name, u.Wxid)
				continue
			} else if !gorm.IsRecordNotFoundError(err) {
				klog.Error(err)
				return err
			}
			_, err := database.GetGroupUserByWxid(group.Wxid, u.Wxid)
			if err != nil {
				if gorm.IsRecordNotFoundError(err) {
					existUsers = append(existUsers, &database.User{
						NickName:         u.Name,
						Wxid:             u.Wxid,
						GroupWxid:        group.Wxid,
						InviteUserNumber: 0,
						Alerted:          false,
						Role:             database.UserRoleNormal,
					})
				} else {
					klog.Error(err)
					return err
				}
			}
		}
	}

	if err := database.CreateUsers(existUsers); err != nil {
		klog.Errorf("初始化群聊数据失败：%s", err)
		return err
	}
	klog.Infof("初始化群用户成功，添加%d个新用户", len(existUsers))
	return nil
}

func (im *InviteManager) showChooseWhiteListFile() {
	dlg := new(walk.FileDialog)
	dlg.Filter = "JSON Files (*.json)|*.json"
	dlg.Title = "选择json格式文件"

	if ok, err := dlg.ShowOpen(im.ParentWindow); err != nil {
		klog.Error(err)
	} else if !ok {
		return
	}

	im.OwnerFilePathEdit.SetText(dlg.FilePath)
}

func (im *InviteManager) importOwnerList() {
	if len(im.OwnerFilePathEdit.Text()) == 0 {
		walk.MsgBox(im.ParentWindow, "错误", fmt.Sprintf("请选择json文件！"), walk.MsgBoxIconError)
		return
	}

	fileData, err := ioutil.ReadFile(im.OwnerFilePathEdit.Text())
	if err != nil {
		walk.MsgBox(im.ParentWindow, "错误", fmt.Sprintf("读取json文件失败: %s！", err.Error()), walk.MsgBoxIconError)
		return
	}
	data := make([]byte, 0)
	for _, b := range fileData {
		if b >= 32 && b <= 126 {
			data = append(data, b)
		}
	}

	users := make([]*database.User, 0)
	if err := json.Unmarshal(data, &users); err != nil {
		walk.MsgBox(im.ParentWindow, "错误", fmt.Sprintf("解析json文件失败: %s！", err.Error()), walk.MsgBoxIconError)
		return
	}

	if len(users) == 0 {
		walk.MsgBox(im.ParentWindow, "错误", fmt.Sprintf("解析json文件失败: %s！", "文件包含用户数为0"), walk.MsgBoxIconError)
		return
	}

	owners := make([]config.CommonUserInfo, 0)
	for _, u := range users {
		owners = append(owners, config.CommonUserInfo{
			Wxid: u.Wxid,
			Name: u.NickName,
		})
	}

	config.GlobalConfig.InviteMangerConf.ManageOwners = owners
	groupsStr := utils.GetCommonUsersNameStr(config.GlobalConfig.InviteMangerConf.ManageOwners)
	im.ManageOwnersTextEdit.SetText(groupsStr)
	im.ManageOwnersStr = groupsStr
}

func (im *InviteManager) chooseAllGroups() {
	if len(config.GlobalConfig.LocalUser.RobotWxid) == 0 {
		if err := utils.SetLocalUserInfo(); err != nil {
			walk.MsgBox(im.ParentWindow, "错误", fmt.Sprintf("获取登录信息失败！'%s'", err), walk.MsgBoxIconError)
			return
		}
	}
	groups, err := wechat.GetGroupList(config.GlobalConfig.LocalUser.RobotWxid)
	if err != nil {
		walk.MsgBox(im.ParentWindow, "错误", fmt.Sprintf("获取所有群失败！'%s'", err), walk.MsgBoxIconError)
	}
	config.GlobalConfig.InviteMangerConf.ManageOwners = groups

	groupsStr := utils.GetCommonUsersNameStr(config.GlobalConfig.InviteMangerConf.ManageOwners)
	im.ManageOwnersTextEdit.SetText(groupsStr)
}
