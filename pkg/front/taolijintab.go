package front

import (
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/wppzxc/wechat-tools/pkg/config"
)

type TaoLiJin struct {
	ParentWindow *walk.MainWindow
	MainPage     *TabPage
}

func GetTaoLiJinPage(mw *walk.MainWindow) *TaoLiJin {
	tlj := &TaoLiJin{
		ParentWindow: mw,
	}

	if config.GlobalConfig.TaoLiJinConf == nil {
		config.GlobalConfig.TaoLiJinConf = new(config.TaoLiJinConf)
	}

	tlj.MainPage = &TabPage{
		Title:  "淘礼金",
		Layout: VBox{},
		DataBinder: DataBinder{
			AutoSubmit: true,
			DataSource: config.GlobalConfig.TaoLiJinConf,
		},
		Children: []Widget{
			Composite{
				Layout: VBox{},
				Children: []Widget{
					Composite{
						Layout: HBox{},
						Children: []Widget{
							CheckBox{
								Text:    "开启淘礼金发送",
								Checked: Bind("Start"),
							},
							Label{
								Text: "刷新间隔",
							},
							NumberEdit{
								Value:    Bind("Interval", Range{0, 999}),
								Suffix:   " / 秒",
								Decimals: 0,
							},
						},
					},
					Label{
						Text: "淘宝API设置",
					},
					Composite{
						Layout: HBox{},
						Children: []Widget{
							Label{
								Text: "淘宝API appKey: ",
							},
							LineEdit{
								Text: Bind("TBAppKey"),
							},
						},
					},
					Composite{
						Layout: HBox{},
						Children: []Widget{
							Label{
								Text: "淘宝API appSecret: ",
							},
							LineEdit{
								Text: Bind("TBAppSecret"),
							},
						},
					},
					Composite{
						Layout: HBox{},
						Children: []Widget{
							Label{
								Text: "淘宝API 推广位ID: ",
							},
							LineEdit{
								Text: Bind("TBAdzoneID"),
							},
						},
					},
				},
			},
			Composite{
				Layout: VBox{},
				Children: []Widget{
					Label{
						Text: "大淘客API设置",
					},
					Composite{
						Layout: HBox{},
						Children: []Widget{
							Label{
								Text: "大淘客API APP_KEY: ",
							},
							LineEdit{
								Text: Bind("DTKAppKey"),
							},
						},
					},
					Composite{
						Layout: HBox{},
						Children: []Widget{
							Label{
								Text: "淘宝API APP_SECRET: ",
							},
							LineEdit{
								Text: Bind("DTKAppSecret"),
							},
						},
					},
				},
			},
			Composite{
				Layout: VBox{},
				Children: []Widget{
					Composite{
						Layout: HBox{},
						Children: []Widget{
							Label{
								Text: "淘礼金占佣金比例",
							},
							NumberEdit{
								Value:    Bind("TBPerFaceRate", Range{0, 100}),
								Suffix:   " /%",
								Decimals: 0,
							},
						},
					},
					Composite{
						Layout: HBox{},
						Children: []Widget{
							Label{
								Text: "淘礼金数量",
							},
							LineEdit{
								Text: Bind("TBTotalNum"),
							},
						},
					},
				},
			},
		},
	}
	return tlj
}
