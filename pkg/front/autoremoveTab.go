package front

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/tealeg/xlsx"
	"github.com/wppzxc/wechat-tools/pkg/config"
	"github.com/wppzxc/wechat-tools/pkg/database"
	"github.com/wppzxc/wechat-tools/pkg/wechat"
	"k8s.io/klog"
)

type AutoRemover struct {
	ParentWindow *walk.MainWindow
	MainPage     *TabPage
	WhiteList    *walk.LineEdit
	BlackList    *walk.LineEdit
}

func GetAutoRemoverPage(mw *walk.MainWindow) *AutoRemover {
	ar := &AutoRemover{
		ParentWindow: mw,
	}
	if config.GlobalConfig.AutoRemoveConf == nil {
		config.GlobalConfig.AutoRemoveConf = new(config.AutoRemoveConf)
	}

	ar.MainPage = &TabPage{
		Title:  "群管设置",
		Layout: VBox{},
		DataBinder: DataBinder{
			AutoSubmit: true,
			DataSource: config.GlobalConfig.AutoRemoveConf,
		},
		Children: []Widget{
			Composite{
				Layout: HBox{},
				Children: []Widget{
					Label{
						Text: "是否启用",
					},
					CheckBox{
						Checked: Bind("Start"),
					},
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					Label{
						Text: "发送链接",
					},
					CheckBox{
						Checked: Bind("SendLink"),
					},
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					Label{
						Text: "发送二维码",
					},
					CheckBox{
						Checked: Bind("SendQRCode"),
					},
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					Label{
						Text: "发送视频",
					},
					CheckBox{
						Checked: Bind("SendVideo"),
					},
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					Label{
						Text: "发送语音",
					},
					CheckBox{
						Checked: Bind("SendVoice"),
					},
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					Label{
						Text: "发送名片",
					},
					CheckBox{
						Checked: Bind("SendCard"),
					},
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					Label{
						Text: "发送分享链接",
					},
					CheckBox{
						Checked: Bind("ShareLink"),
					},
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					Label{
						Text: "发送小程序",
					},
					CheckBox{
						Checked: Bind("Applets"),
					},
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					Label{
						Text: "发送内容包含过滤词",
					},
					CheckBox{
						Checked: Bind("FilterWords"),
					},
					Label{
						Text: "指定过滤词：",
					},
					LineEdit{
						Text: Bind("FilterWordsString"),
					},
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					Label{
						Text: "用户名称包含过滤词",
					},
					CheckBox{
						Checked: Bind("FilterNames"),
					},
					Label{
						Text: "指定过滤词",
					},
					LineEdit{
						Text: Bind("FilterNamesString"),
					},
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					Label{
						Text: "防止炸群",
					},
					CheckBox{
						Checked: Bind("MsgLength"),
					},
					Label{
						Text: "消息最大长度",
					},
					NumberEdit{
						Value:    Bind("MaxMsgLength", Range{0, 99}),
						Suffix:   " /(max 99)",
						Decimals: 0,
					},
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					LineEdit{
						AssignTo: &ar.WhiteList,
					},
					PushButton{
						Text:      "选择",
						OnClicked: ar.showChooseWhiteListFile,
					},
					PushButton{
						Text:      "导入白名单",
						OnClicked: ar.importWhiteList,
					},
					PushButton{
						Text:      "导出",
						OnClicked: ar.exportWhiteList,
					},
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					LineEdit{
						AssignTo: &ar.BlackList,
					},
					PushButton{
						Text:      "选择",
						OnClicked: ar.showChooseBlackListFile,
					},
					PushButton{
						Text:      "导入黑名单",
						OnClicked: ar.importBlackList,
					},
					PushButton{
						Text:      "导出",
						OnClicked: ar.exportBlackList,
					},
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					PushButton{
						Text: "查看管理员",
						OnClicked: func() {
							cmd := exec.Command(`cmd`, `/c`, `start`, `http://127.0.0.1:8074/roles/manager/users`)
							cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
							cmd.Start()
						},
					},
					PushButton{
						Text: "查看白名单",
						OnClicked: func() {
							cmd := exec.Command(`cmd`, `/c`, `start`, `http://127.0.0.1:8074/roles/white/users`)
							cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
							cmd.Start()
						},
					},
					PushButton{
						Text: "查看黑名单",
						OnClicked: func() {
							cmd := exec.Command(`cmd`, `/c`, `start`, `http://127.0.0.1:8074/roles/black/users`)
							cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
							cmd.Start()
						},
					},
					PushButton{
						Text:      "导出群成员数量",
						OnClicked: ar.exportGroupsMemberInfo,
					},
				},
			},
		},
	}
	return ar
}

func (ar *AutoRemover) showChooseWhiteListFile() {
	dlg := new(walk.FileDialog)
	dlg.Filter = "JSON Files (*.json)|*.json"
	dlg.Title = "选择json格式白名单文件"

	if ok, err := dlg.ShowOpen(ar.ParentWindow); err != nil {
		klog.Error(err)

	} else if !ok {
		return
	}

	ar.WhiteList.SetText(dlg.FilePath)
}

func (ar *AutoRemover) showChooseBlackListFile() {
	dlg := new(walk.FileDialog)
	dlg.Filter = "JSON Files (*.json)|*.json"
	dlg.Title = "选择json格式黑名单文件"

	if ok, err := dlg.ShowOpen(ar.ParentWindow); err != nil {
		klog.Error(err)

	} else if !ok {
		return
	}

	ar.BlackList.SetText(dlg.FilePath)
}

func (ar *AutoRemover) importWhiteList() {
	if len(ar.WhiteList.Text()) == 0 {
		walk.MsgBox(ar.ParentWindow, "错误", fmt.Sprintf("请选择白名单文件！"), walk.MsgBoxIconError)
		return
	}

	fileData, err := ioutil.ReadFile(ar.WhiteList.Text())
	if err != nil {
		walk.MsgBox(ar.ParentWindow, "错误", fmt.Sprintf("读取白名单文件失败: %s！", err.Error()), walk.MsgBoxIconError)
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
		klog.Error(err)
		walk.MsgBox(ar.ParentWindow, "错误", fmt.Sprintf("解析白名单文件失败: %s！", err.Error()), walk.MsgBoxIconError)
		return
	}

	if err := database.DeleteBlackLists(users); err != nil {
		klog.Error(err)
		walk.MsgBox(ar.ParentWindow, "错误", fmt.Sprintf("删除黑名单信息失败: %s！", err.Error()), walk.MsgBoxIconError)
		return
	}

	num := 0
	for _, u := range users {
		if err := database.CreateWhiteList(u); err != nil {
			klog.Error(err)
			continue
		}
		num++
	}

	walk.MsgBox(ar.ParentWindow, "成功", fmt.Sprintf("成功导入白名单: %d个！", num), walk.MsgBoxIconInformation)
}

func (ar *AutoRemover) importBlackList() {
	if len(ar.BlackList.Text()) == 0 {
		walk.MsgBox(ar.ParentWindow, "错误", fmt.Sprintf("请选择黑名单文件！"), walk.MsgBoxIconError)
		return
	}
	fileData, err := ioutil.ReadFile(ar.BlackList.Text())
	if err != nil {
		walk.MsgBox(ar.ParentWindow, "错误", fmt.Sprintf("读取黑名单文件失败: %s！", err.Error()), walk.MsgBoxIconError)
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
		walk.MsgBox(ar.ParentWindow, "错误", fmt.Sprintf("解析黑名单文件失败: %s！", err.Error()), walk.MsgBoxIconError)
		return
	}
	if err := database.DeleteWhiteLists(users); err != nil {
		klog.Error(err)
		walk.MsgBox(ar.ParentWindow, "错误", fmt.Sprintf("删除黑名单信息失败: %s！", err.Error()), walk.MsgBoxIconError)
		return
	}

	num := 0
	for _, u := range users {
		if err := database.CreateBlackList(u); err != nil {
			klog.Error(err)
			continue
		}
		num++
	}

	walk.MsgBox(ar.ParentWindow, "成功", fmt.Sprintf("成功导入白名单: %d个！", num), walk.MsgBoxIconInformation)
}

type exportUser struct {
	HeadImg  string `json:"head_img"`
	NickName string `json:"nick_name"`
	Username string `json:"user_name"`
	Wxid     string `json:"wxid"`
}

func (ar *AutoRemover) exportWhiteList() {
	whiteUsers, err := database.GetAllWhiteLists()
	if err != nil {
		walk.MsgBox(ar.ParentWindow, "错误", fmt.Sprintf("获取白名单失败: %s！", err.Error()), walk.MsgBoxIconError)
		return
	}

	exportUsers := make([]exportUser, 0)
	for _, u := range whiteUsers {
		exportUsers = append(exportUsers, exportUser{Wxid: u.Wxid})
	}
	if len(exportUsers) == 0 {
		walk.MsgBox(ar.ParentWindow, "错误", fmt.Sprintf("没有白名单用户！"), walk.MsgBoxIconError)
		return
	}

	data, err := json.Marshal(exportUsers)
	if err != nil {
		walk.MsgBox(ar.ParentWindow, "错误", fmt.Sprintf("导出白名单失败: %s！", err.Error()), walk.MsgBoxIconError)
		return
	}

	filename := fmt.Sprintf("./%d-白名单.json", time.Now().Unix())
	if err := ioutil.WriteFile(filename, data, os.ModePerm); err != nil {
		walk.MsgBox(ar.ParentWindow, "错误", fmt.Sprintf("导出白名单文件失败: %s！", err.Error()), walk.MsgBoxIconError)
		return
	}
	walk.MsgBox(ar.ParentWindow, "成功", fmt.Sprintf("导出成功"), walk.MsgBoxIconInformation)
	return
}

func (ar *AutoRemover) exportBlackList() {
	blackUsers, err := database.GetAllBlackLists()
	if err != nil {
		walk.MsgBox(ar.ParentWindow, "错误", fmt.Sprintf("获取黑名单失败: %s！", err.Error()), walk.MsgBoxIconError)
		return
	}

	exportUsers := make([]exportUser, 0)
	for _, u := range blackUsers {
		exportUsers = append(exportUsers, exportUser{Wxid: u.Wxid})
	}
	if len(exportUsers) == 0 {
		walk.MsgBox(ar.ParentWindow, "错误", fmt.Sprintf("没有黑名单用户！"), walk.MsgBoxIconError)
		return
	}

	data, err := json.Marshal(exportUsers)
	if err != nil {
		walk.MsgBox(ar.ParentWindow, "错误", fmt.Sprintf("导出黑名单失败: %s！", err.Error()), walk.MsgBoxIconError)
		return
	}

	filename := fmt.Sprintf("./%d-黑名单.json", time.Now().Unix())
	if err := ioutil.WriteFile(filename, data, os.ModePerm); err != nil {
		walk.MsgBox(ar.ParentWindow, "错误", fmt.Sprintf("导出黑名单文件失败: %s！", err.Error()), walk.MsgBoxIconError)
		return
	}
	walk.MsgBox(ar.ParentWindow, "成功", fmt.Sprintf("导出成功"), walk.MsgBoxIconInformation)
	return
}

func (ar *AutoRemover) exportGroupsMemberInfo() {
	groups, err := wechat.GetGroupList(config.GlobalConfig.LocalUser.RobotWxid)
	if err != nil {
		walk.MsgBox(ar.ParentWindow, "错误", fmt.Sprintf("导出群信息失败: %s！", err.Error()), walk.MsgBoxIconError)
	}
	filename := fmt.Sprintf("./%d-群信息.xlsx", time.Now().Unix())
	var file *xlsx.File
	var sheet *xlsx.Sheet
	var row *xlsx.Row
	var cell *xlsx.Cell
	file = xlsx.NewFile()
	sheet, _ = file.AddSheet("美逛后台数据")
	row = sheet.AddRow()

	cell = row.AddCell()
	cell.Value = "群名称"
	cell = row.AddCell()
	cell.Value = "群id"
	cell = row.AddCell()
	cell.Value = "群成员数量"

	for _, g := range groups {
		row = sheet.AddRow()
		cell = row.AddCell()
		cell.Value = g.Name
		cell = row.AddCell()
		cell.Value = g.Wxid
		cell = row.AddCell()
		members, err := wechat.GetGroupUserList(config.GlobalConfig.LocalUser.RobotWxid, g.Wxid)
		if err != nil {
			cell.Value = "读取失败"
		} else {
			cell.Value = fmt.Sprintf("%d", len(members))
		}
	}

	if err := file.Save(filename); err != nil {
		walk.MsgBox(ar.ParentWindow, "错误", fmt.Sprintf("保存文件失败: %s！", err.Error()), walk.MsgBoxIconError)
	}
}
