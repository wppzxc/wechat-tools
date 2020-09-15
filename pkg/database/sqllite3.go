package database

import (
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
)

const (
	// SpecType
	WhiteUser = "whiteUser"
	BlackUser = "blackUser"

	// ManageType
	GroupManager  = "GroupManager"
	SystemManager = "systemManager"
)

var dbConn *gorm.DB

const (
	UserRoleOwner   = "owner"
	UserRoleManager = "manager"
	UserRoleVip     = "vip"
	UserRoleNormal  = "normal"
	UserRoleBlack   = "black"
	UserRoleWhite   = "white"
)

// 用户
type User struct {
	gorm.Model
	GroupWxid        string `gorm:"type:varchar(72)"`
	NickName         string `gorm:"type:varchar(255)" json:"nick_name"`
	Wxid             string `gorm:"type:varchar(72)"`
	InviteUserNumber int    `gorm:"type:int(10)"`
	Alerted          bool
	Active           bool
	// 角色可以是：normal(普通用户), manager(二级管理员), owner(一级管理员)
	Role string `gorm:"type:varchar(16)"`
}

// 黑白名单用户
type BlackList struct {
	gorm.Model
	Wxid string `gorm:"type:varchar(72)"`
}

type WhiteList struct {
	gorm.Model
	Wxid string `gorm:"type:varchar(72)"`
}

func InitDB() *gorm.DB {
	var err error
	dbConn, err = gorm.Open("sqlite3", "./wechat-tools.db")
	if err != nil {
		panic(err)
	}
	dbConn.SetLogger(dbLogger{})
	dbConn.LogMode(true)
	dbConn.AutoMigrate(User{})
	dbConn.AutoMigrate(BlackList{})
	dbConn.AutoMigrate(WhiteList{})
	return dbConn
}

func GetDB() *gorm.DB {
	return dbConn
}
