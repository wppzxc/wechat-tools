package front

import (
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/wppzxc/wechat-tools/pkg/config"
)

type AutoRemover struct {
	ParentWindow *walk.MainWindow
	MainPage     *TabPage
}

func GetAutoRemoverPage(mw *walk.MainWindow) *AutoRemover {
	ar := &AutoRemover{
		ParentWindow: mw,
	}
	if config.GlobalConfig.AutoRemoveConf == nil {
		config.GlobalConfig.AutoRemoveConf = new(config.AutoRemoveConf)
	}

	ar.MainPage = &TabPage{
		Title:  "自动踢人",
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
		},
	}
	return ar
}
