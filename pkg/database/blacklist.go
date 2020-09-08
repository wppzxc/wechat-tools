package database

import "github.com/labstack/gommon/log"

func CreateBlackList(user *User) error {
	blackList := new(BlackList)
	blackList.Wxid = user.Wxid
	tx := dbConn.Begin()
	err := tx.Create(blackList).Error
	if err != nil {
		log.Error(err)
		tx.Rollback()
		return err
	}
	tx.Commit()
	return err
}

func GetBlackListByWxid(wxid string) (*BlackList, error) {
	blackList := new(BlackList)
	if err := dbConn.Where("wxid = ? ", wxid).Find(blackList).Error; err != nil {
		return nil, err
	}
	return blackList, nil
}

func GetAllBlackLists() ([]BlackList, error) {
	blacklists := make([]BlackList, 0)
	if err := dbConn.Find(&blacklists).Error; err != nil {
		return nil, err
	}
	return blacklists, nil
}

func DeleteBlackLists(users []*User) error {
	tx := dbConn.Begin()
	for _, bl := range users {
		if err := dbConn.Where("wxid = ? ", bl.Wxid).Delete(BlackList{}).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	tx.Commit()
	return nil
}
