/***********************************************************************
// Copyright qoobing.com @2017 The source code.
// Copyright (c) 2009-2016 The Bitcoin Core developers
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php.
//******
// Filename:
// Description:
// Author:
// CreateTime:
/***********************************************************************/
package t_rbac_user_role

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"qoobing.com/utillib.golang/log"
	"strings"
	"time"
)

const (
	TABLENAME = "user_t_rbac_user_role"
	TABLESQL  = "CREATE TABLE IF NOT EXISTS " + TABLENAME +
		`(
			F_id bigint(20) unsigned NOT NULL AUTO_INCREMENT, 
			F_status int(4)  NOT NULL DEFAULT 0,
			F_status_ext varchar(64)  NOT NULL DEFAULT "",
			F_user_id bigint(20) unsigned NOT NULL DEFAULT 0,
			F_role_id bigint(20) unsigned NOT NULL DEFAULT 0,
			F_scene_key varchar(128) NOT NULL DEFAULT "",
			F_create_time datetime NOT NULL,
			F_modify_time datetime NOT NULL,
	
			PRIMARY KEY (F_id),
			INDEX (F_scene_key)
		) ENGINE=InnoDB  DEFAULT CHARSET=utf8 ;`
)

type UserRole struct {
	F_id          uint64 `gorm:"column:F_id;primary_key"` //全局唯一id
	F_status      int    `gorm:"column:F_status"`         //状态
	F_status_ext  string `gorm:"column:F_status_ext"`     //状态扩展信息
	F_user_id     uint64 `gorm:"column:F_user_id"`        //角色id
	F_role_id     uint64 `gorm:"column:F_role_id"`        //用户id
	F_scene_key   string `gorm:"column:F_scene_key"`      //场景值
	F_create_time string `gorm:"column:F_create_time"`    //创建时间
	F_modify_time string `gorm:"column:F_modify_time"`    //修改时间
}

const (
	USERROLE_STATUS_INVALID   = 0  //非法状态，不应该存在
	USERROLE_STATUS_OK        = 1  //正常状态
	USERROLE_STATUS_DELETED   = 2  //已删除
	USERROLE_STATUS_FORBIDDEN = 10 //已禁止
)

func (userrole *UserRole) TableName() string {
	return TABLENAME
}

func (userrole *UserRole) BeforeCreate(scope *gorm.Scope) error {
	currentTime := time.Now().Local()
	nowFormat := currentTime.Format("2006-01-02 15:04:05.000")
	scope.SetColumn("F_create_time", nowFormat)
	scope.SetColumn("F_modify_time", nowFormat)
	return nil
}

func (userrole *UserRole) BeforeUpdate(scope *gorm.Scope) (err error) {
	currentTime := time.Now().Local()
	nowFormat := currentTime.Format("2006-01-02 15:04:05.000")
	scope.SetColumn("F_modify_time", nowFormat)
	return nil
}

func (userrole *UserRole) Save(db *gorm.DB) (err error) {
	rdb := db.Save(userrole)
	if rdb.Error != nil {
		panic("save privilege error:" + rdb.Error.Error())
	} else {
		err = nil
	}
	return nil
}

func CreateUserRoles(db *gorm.DB, scenekey string, userid uint64, rolesid []uint64) (err error) {
	if len(rolesid) <= 0 {
		log.Debugf("Nothing need to do with [%s]", "CreateRoleMaps")
		return nil
	}

	fieldsNum := 7
	valueStrings := make([]string, 0, len(rolesid))
	valueArgs := make([]interface{}, 0, len(rolesid)*fieldsNum)
	valueLocat := "?"
	for i := 1; i < fieldsNum; i++ {
		valueLocat += ",?"
	}
	valueLocat = fmt.Sprintf("(%s)", valueLocat)
	currentTime := time.Now().Local()
	nowFormat := currentTime.Format("2006-01-02 15:04:05.000")

	for _, rid := range rolesid {
		valueStrings = append(valueStrings, valueLocat)
		valueArgs = append(valueArgs, USERROLE_STATUS_OK) //状态
		valueArgs = append(valueArgs, "OK")               //状态扩展信息
		valueArgs = append(valueArgs, userid)             //用户id
		valueArgs = append(valueArgs, rid)                //角色id
		valueArgs = append(valueArgs, scenekey)           //场景值
		valueArgs = append(valueArgs, nowFormat)          //创建时间
		valueArgs = append(valueArgs, nowFormat)          //修改时间
	}
	stmt := fmt.Sprintf(`INSERT INTO %s (
			F_status,
			F_status_ext,
			F_user_id,
			F_role_id,
			F_scene_key,
			F_create_time,
			F_modify_time
		) VALUES %s`,
		TABLENAME,
		strings.Join(valueStrings, ","))

	rdb := db.Exec(stmt, valueArgs...)
	if rdb.Error != nil {
		panic("create user role error:" + rdb.Error.Error())
	} else {
		err = nil
	}
	return
}

func UpdateUserRoleToStatus(db *gorm.DB, userroles []UserRole) (err error) {
	if len(userroles) <= 0 {
		log.Debugf("Nothing need to do with [%s]", "UpdateUserRoleToStatus")
		return nil
	}
	var (
		ids         = []uint64{}
		currentTime = time.Now().Local()
		target      = map[string]interface{}{
			"F_status":      userroles[0].F_status,
			"F_modify_time": currentTime.Format("2006-01-02 15:04:05.000"),
		}
	)

	for _, rm := range userroles {
		ids = append(ids, rm.F_id)
	}
	rdb := db.Model(&UserRole{}).Where("F_id IN (?)", ids).Update(target)

	if rdb.Error != nil {
		panic("UpdateUserRoleToStatus error:" + rdb.Error.Error())
	} else {
		err = nil
	}
	return nil
}

func FindUserRolesByScenekeyAndUserids(db *gorm.DB, scenekey string, userid uint64) (userroles []UserRole, err error) {
	var rdb *gorm.DB
	rdb = db.Where("F_scene_key = ? AND F_user_id = ? AND F_status = ? ", scenekey, userid, USERROLE_STATUS_OK).Find(&userroles)
	if rdb.RecordNotFound() {
		return userroles, nil
	} else if rdb.Error != nil {
		panic("find userroles error:" + rdb.Error.Error())
	} else {
		err = nil
	}
	return
}

func FindUserRoles(db *gorm.DB, userid uint64) (userroles []UserRole, err error) {
	var rdb *gorm.DB
	rdb = db.Where("F_user_id = ? AND F_status = ?", userid, USERROLE_STATUS_OK).Find(&userroles)
	if rdb.RecordNotFound() {
		return userroles, nil
	} else if rdb.Error != nil {
		panic("find userroles error:" + rdb.Error.Error())
	} else {
		err = nil
	}
	return
}

func FindUsersRolesByScenekeyAndUserids(db *gorm.DB, scenekey string, userids []uint64) (userroles []UserRole, err error) {
	var rdb *gorm.DB
	rdb = db.Where("F_scene_key = ? AND F_user_id IN (?) AND F_status = ?", scenekey, userids, USERROLE_STATUS_OK).Find(&userroles)
	if rdb.RecordNotFound() {
		return userroles, nil
	} else if rdb.Error != nil {
		panic("find userroles error:" + rdb.Error.Error())
	} else {
		err = nil
	}
	return
}
