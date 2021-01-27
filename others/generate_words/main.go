package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"strings"
	"time"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

const (
	defaultFIle = "words.txt"
)

// Config config for main
type Config struct {
	MainWord     string `json:"mainWord"`
	Length       int    `json:"length"`
	SplitWord    string `json:"splitWord"`
	ResultNumber int    `json:"resultNumber"`
}

var (
	mw     *walk.MainWindow = new(walk.MainWindow)
	config *Config          = new(Config)
)

func main() {
	rand.Seed(time.Now().Unix())
	if _, err := (MainWindow{
		Title:    "词语生成",
		AssignTo: &mw,
		Size:     Size{Width: 400, Height: 300},
		Layout:   VBox{},
		DataBinder: DataBinder{
			AutoSubmit: true,
			DataSource: config,
		},
		Children: []Widget{
			Composite{
				Layout: HBox{},
				Children: []Widget{
					Label{
						Text: "中心词",
					},
					LineEdit{
						Text: Bind("MainWord"),
					},
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					Label{
						Text: "长度",
					},
					NumberEdit{
						Value:    Bind("Length", Range{0, 10000}),
						Decimals: 0,
					},
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					Label{
						Text: "分隔符",
					},
					LineEdit{
						Text: Bind("SplitWord"),
					},
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					Label{
						Text: "生成数量",
					},
					NumberEdit{
						Value:    Bind("ResultNumber", Range{0, 10000}),
						Decimals: 0,
					},
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					PushButton{
						Text:      "生成并导出",
						OnClicked: startAndExport,
					},
				},
			},
		},
	}).Run(); err != nil {
		panic(err)
	}
}

func startLoopCheck() {
	filename := defaultFIle
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		walk.MsgBox(mw, "错误", fmt.Sprintf("读取词库错误：%s", err.Error()), walk.MsgBoxIconError)
		return
	}

	words := strings.Split(string(content), "\r\n")

	results, err := generateWords(words)
	if err != nil {
		walk.MsgBox(mw, "错误", fmt.Sprintf("生成词汇错误：%s", err.Error()), walk.MsgBoxIconError)
		return
	}

	var dlg *walk.Dialog
	var acceptPB, cancelPB *walk.PushButton

	widgets := make([]Widget, 0)
	for _, r := range results {
		widgets = append(widgets, TextEdit{Text: r, ReadOnly: true})
	}

	Dialog{
		AssignTo:      &dlg,
		Title:         "结果展示",
		DefaultButton: &acceptPB,
		CancelButton:  &cancelPB,
		MinSize:       Size{300, 300},
		Layout:        VBox{},
		Children: []Widget{
			Composite{
				Layout:   VBox{},
				Children: widgets,
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					HSpacer{},
					PushButton{
						AssignTo: &acceptPB,
						Text:     "OK",
						OnClicked: func() {
							dlg.Accept()
						},
					},
					PushButton{
						AssignTo:  &cancelPB,
						Text:      "Cancel",
						OnClicked: func() { dlg.Cancel() },
					},
				},
			},
		},
	}.Run(mw)
}

func startAndExport() {
	filename := defaultFIle
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		walk.MsgBox(mw, "错误", fmt.Sprintf("读取词库错误：%s", err.Error()), walk.MsgBoxIconError)
		return
	}

	words := strings.Split(string(content), "\r\n")

	results, err := generateWords(words)
	if err != nil {
		walk.MsgBox(mw, "错误", fmt.Sprintf("生成词汇错误：%s", err.Error()), walk.MsgBoxIconError)
		return
	}

	str := strings.Join(results, "\n")

	if err := ioutil.WriteFile(fmt.Sprintf("%d.txt", time.Now().Unix()), []byte(str), 0600); err != nil {
		walk.MsgBox(mw, "错误", fmt.Sprintf("导出错误：%s", err.Error()), walk.MsgBoxIconError)
	}

	walk.MsgBox(mw, "成功", "导出成功", walk.MsgBoxIconInformation)

}

func generateWords(words []string) ([]string, error) {
	results := make([]string, 0)
	wordsLen := len(words)
	for i := 0; i < config.ResultNumber; i++ {
		result := config.MainWord
		for {
			word := words[rand.Intn(wordsLen)]
			result = result + config.SplitWord + word
			resultLen := len([]rune(result))
			if resultLen >= config.Length {
				results = append(results, result)
				break
			}
		}
	}
	return results, nil
}
