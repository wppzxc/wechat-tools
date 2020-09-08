package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
	"unsafe"

	"github.com/wppzxc/wechat-tools/pkg/config"
	"github.com/wppzxc/wechat-tools/pkg/types"
	"github.com/wppzxc/wechat-tools/pkg/utils"
)

func TestGetAtNameAndWxid(t *testing.T) {
	//msg := "[@at,nickname=闹闹,wxid=wxid_9866998668912]  测试"
	msg := "[@at,nickname=小毛,wxid=wxid_lluoefrhpwlc22]  测试"
	nickname, wxid, newMsg, yes := utils.IsAtMsg(msg)
	fmt.Printf("is at msg=%t\n", yes)
	fmt.Printf("nickname=%s\n", nickname)
	fmt.Printf("wxid=%s\n", wxid)
	fmt.Printf("newMsg=%s\n", newMsg)
	fmt.Printf("msg=%s\n", msg)
}

func TestGetLocalUserInfo(t *testing.T) {
	resp, err := http.Post("http://192.168.28.55:8073/send", "",
		strings.NewReader(fmt.Sprintf(`{"type": 204, "robot_wxid": "wxid_lluoefrhpwlc22", "is_refresh":1}`)))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	respData := new(types.ResponseUserList)
	if err := json.Unmarshal(data, respData); err != nil {
		fmt.Printf("Error in get local user info : %s ", err)
		return
	}
	fmt.Printf("%+v", respData.Data)
}

func TestCheckWeChat(t *testing.T) {
	config.GlobalConfig = new(config.Config)
	config.GlobalConfig.LocalUser = new(config.LocalUserInfo)
	config.GlobalConfig.LocalUser.RobotWxid = "wxid_lluoefrhpwlc22"
	if err := utils.CheckWechat(); err != nil {
		fmt.Println("Error in check wechat : ", err)
		return
	}

	fmt.Println("check wechat ok!")
}

type ulong int32
type ulong_ptr uintptr

type PROCESSENTRY32 struct {
	dwSize              ulong
	cntUsage            ulong
	th32ProcessID       ulong
	th32DefaultHeapID   ulong_ptr
	th32ModuleID        ulong
	cntThreads          ulong
	th32ParentProcessID ulong
	pcPriClassBase      ulong
	dwFlags             ulong
	szExeFile           [260]byte
}

func TestGetWechatID(t *testing.T) {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	CreateToolhelp32Snapshot := kernel32.NewProc("CreateToolhelp32Snapshot")
	pHandle, _, _ := CreateToolhelp32Snapshot.Call(uintptr(0x2), uintptr(0x0))
	if int(pHandle) == -1 {
		return
	}
	Process32Next := kernel32.NewProc("Process32Next")
	for {
		var proc PROCESSENTRY32
		proc.dwSize = ulong(unsafe.Sizeof(proc))
		if rt, _, _ := Process32Next.Call(uintptr(pHandle), uintptr(unsafe.Pointer(&proc))); int(rt) == 1 {
			name := string(proc.szExeFile[0:])
			if strings.Index(name, "WeChat.") >= 0 {
				fmt.Println("微信正在运行")
				fmt.Println("ProcessName : " + string(proc.szExeFile[0:]))
				fmt.Println("th32ModuleID : " + strconv.Itoa(int(proc.th32ModuleID)))
				fmt.Println("ProcessID : " + strconv.Itoa(int(proc.th32ProcessID)))
			}
		} else {
			break
		}
	}
	fmt.Printf("\n30秒后会自动关闭")
	time.Sleep(30 * time.Second)
	CloseHandle := kernel32.NewProc("CloseHandle")
	_, _, _ = CloseHandle.Call(pHandle)
}

func TestWindowsGetTaskList(t *testing.T) {
	var appName string = "WeChat.exe"

	cmd := exec.Command("tasklist", "/V")
	output, _ := cmd.CombinedOutput()

	n := strings.Index(string(output), "System")
	if n == -1 {
		fmt.Println("no find")
		os.Exit(1)
	}
	data := string(output)[n:]
	Processes := strings.Split(data, "\n")
	fmt.Println(Processes)
	for _, process := range Processes {
		if strings.Index(process, appName) >= 0 {
			fmt.Printf("\n微信进程如下：\n")
			fmt.Println(process)
			vals := strings.Fields(process)
			fmt.Printf("微信运行状态为: %s\n", vals[6])
		}
	}

	time.Sleep(30 * time.Second)
	return
}

func TestDingding(t *testing.T) {
	msg := `{
		"msgtype": "text", 
		"text": {
			"content": ".我就是我, 是不一样的烟火"
		}
	}`
	resp, err := http.Post("https://oapi.dingtalk.com/robot/send?access_token=9d7888ccd788c2aed4fcc79377c02400a816cd18db6598f09ecd152c7daf8466", 
	"application/json", bytes.NewBufferString(msg))
	if err != nil {
		fmt.Println(err)
	}
	data, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	fmt.Println(string(data))
	fmt.Println("success")
}
