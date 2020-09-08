package test

import (
	"fmt"
	"github.com/lxn/walk"
	"image"
	"image/jpeg"
	"os"
	"strings"
	"testing"
)

type MyMainWindow struct {
	*walk.MainWindow
	model *EnvModel
	lb    *walk.ListBox
	te    *walk.TextEdit
}

type EnvItem struct {
	name  string
	value string
}

type EnvModel struct {
	walk.ListModelBase
	items []EnvItem
}

func NewEnvModel() *EnvModel {
	env := os.Environ()

	m := &EnvModel{items: make([]EnvItem, len(env))}

	for i, e := range env {
		j := strings.Index(e, "=")
		if j == 0 {
			continue
		}

		name := e[0:j]
		value := strings.Replace(e[j+1:], ";", "\r\n", -1)

		m.items[i] = EnvItem{name, value}
	}

	return m
}

func (m *EnvModel) ItemCount() int {
	return len(m.items)
}

func (m *EnvModel) Value(index int) interface{} {
	return m.items[index].name
}

func TestImageRGB(t *testing.T) {
	img2 := image.NewGray(image.Rect(0, 0, 16, 16))
	img2.Pix = []uint8{200, 129, 92, 99, 91, 84, 85, 89, 97, 92, 95, 93, 91, 94, 125, 204, 133, 78, 90, 86, 103, 131, 126, 108, 85, 91, 92, 91, 97, 91, 85, 125, 88, 95, 106, 181, 221, 216, 223, 218, 182, 107, 90, 99, 91, 90, 95, 89, 90, 107, 208, 212, 190, 215, 213, 191, 212, 209, 103, 86, 92, 97, 95, 91, 88, 163, 232, 184, 121, 217, 211, 125, 170, 235, 171, 86, 85, 93, 92, 90, 93, 193, 225, 211, 205, 212, 224, 210, 203, 204, 204, 152, 121, 91, 87, 98, 88, 188, 220, 220, 219, 216, 212, 203, 214, 220, 214, 235, 228, 173, 93, 90, 86, 141, 230, 211, 211, 218, 204, 216, 188, 169, 227, 161, 187, 235, 154, 83, 93, 87, 167, 221, 222, 221, 211, 215, 194, 191, 221, 178, 201, 229, 185, 88, 96, 89, 106, 203, 185, 186, 197, 217, 216, 219, 214, 223, 216, 216, 190, 85, 93, 88, 111, 109, 84, 101, 95, 197, 229, 213, 218, 213, 219, 221, 127, 83, 93, 92, 96, 88, 95, 95, 90, 102, 179, 211, 217, 224, 221, 154, 87, 97, 91, 93, 91, 91, 91, 95, 97, 91, 89, 121, 132, 121, 129, 121, 96, 88, 89, 95, 95, 94, 93, 90, 96, 97, 96, 88, 86, 87, 87, 93, 85, 96, 126, 82, 91, 94, 92, 94, 104, 88, 94, 93, 96, 99, 98, 96, 84, 119, 198, 124, 85, 88, 91, 86, 85, 90, 90, 89, 87, 88, 87, 85, 125, 199}
	img2.Stride = 16
	newFile, err := os.Create("D:\\project\\go\\src\\github.com\\wppzxc\\wechat-tools\\assets\\img\\new-icon.jpg")
	if err != nil {
		fmt.Printf("Error in save image2 : %s", err)
	}
	defer newFile.Close()
	if err := jpeg.Encode(newFile, img2, nil); err != nil {
		fmt.Printf("Error in encode image : %s", err)
	}
}
