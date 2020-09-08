package database

import (
	"k8s.io/klog"
)

func GetGroupUserByWxid(groupWxid string, wxid string) (*User, error) {
	user := new(User)
	if err := dbConn.Where("wxid = ? AND group_wxid = ? ", wxid, groupWxid).Find(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func GetGroupUsersByRole(groupWxid string, role string) ([]User, error) {
	users := make([]User, 0)
	if err := dbConn.Where("role = ? AND group_wxid = ? ", role, groupWxid).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func GetGroupUsersByRoles(groupWxid string, roles []string) ([]User, error) {
	users := make([]User, 0)
	if err := dbConn.Where("role in (?) AND group_wxid = ? ", roles, groupWxid).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func CreateUser(user *User) error {
	tx := dbConn.Begin()
	err := tx.Create(user).Error
	if err != nil {
		klog.Error(err)
		tx.Rollback()
		return err
	}
	tx.Commit()
	return err
}

func CreateUsers(users []*User) error {
	tx := dbConn.Begin()
	var err error
	for _, user := range users {
		err = tx.Create(user).Error
		if err != nil {
			klog.Error(err)
			tx.Rollback()
			return err
		}
	}
	tx.Commit()
	return err
}

func UpdateGroupUserByWxid(user User, groupWxid string, wxid string) error {
	user.ID = 0
	if err := dbConn.Model(user).Where("wxid = ? AND group_wxid = ? ", wxid, groupWxid).Update(user).Error; err != nil {
		klog.Error(err)
		return err
	}
	return nil
}

func DeleteGroupUserByWxid(groupWxid string, wxid string) error {
	tx := dbConn.Begin()
	if err := dbConn.Where("wxid = ? AND group_wxid = ? ", wxid, groupWxid).Delete(User{}).Error; err != nil {
		return err
	}
	tx.Commit()
	return nil
}
