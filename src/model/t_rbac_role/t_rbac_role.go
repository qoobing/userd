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
package t_rbac_role

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/qoobing/userd/src/common"
	"regexp"
	"time"
)

const (
	TABLENAME = "user_t_rbac_role"
	TABLESQL  = "CREATE TABLE IF NOT EXISTS " + TABLENAME +
		`(
			F_id bigint(20) unsigned NOT NULL AUTO_INCREMENT, 
			F_status int(4)  NOT NULL DEFAULT 0,
			F_status_ext varchar(64)  NOT NULL DEFAULT "",
			F_name varchar(128) NOT NULL DEFAULT "",
			F_description varchar(128) NOT NULL DEFAULT "",
			F_scene_key varchar(128) NOT NULL DEFAULT "",
			F_creator bigint(20) unsigned NOT NULL DEFAULT 0, 
			F_modifier bigint(20) unsigned NOT NULL DEFAULT 0, 
			F_create_time datetime NOT NULL,
			F_modify_time datetime NOT NULL,
	
			PRIMARY KEY (F_id),
			UNIQUE KEY (F_scene_key,F_name),
			INDEX (F_scene_key)
		) ENGINE=InnoDB  DEFAULT CHARSET=utf8 ;`
)

type Role struct {
	F_id          uint64 `gorm:"column:F_id;primary_key"` //角色id，全局唯一
	F_status      int    `gorm:"column:F_status"`         //状态
	F_status_ext  string `gorm:"column:F_status_ext"`     //状态扩展信息
	F_name        string `gorm:"column:F_name"`           //权限名称
	F_description string `gorm:"column:F_description"`    //描述
	F_scene_key   string `gorm:"column:F_scene_key"`      //适用场景
	F_creator     uint64 `gorm:"column:F_creator"`        //记录创建者id
	F_modifier    uint64 `gorm:"column:F_modifier"`       //记录修改者id
	F_create_time string `gorm:"column:F_create_time"`    //创建时间
	F_modify_time string `gorm:"column:F_modify_time"`    //修改时间
}

const (
	ROLE_STATUS_INVALID   = 0  //非法状态，不应该存在
	ROLE_STATUS_OK        = 1  //正常状态
	ROLE_STATUS_DELETED   = 2  //已删除
	ROLE_STATUS_FORBIDDEN = 10 //已禁止
)

func (role *Role) TableName() string {
	return TABLENAME
}

func (role *Role) BeforeCreate(scope *gorm.Scope) error {
	if role.F_id == 0 {
		var retrytimes = 0
		for {
			role.F_id = common.GenerateId("ROLE_ID")
			if roles, err := FindRolesByIds(scope.DB(), []uint64{role.F_id}); err == nil && len(roles) == 0 {
				break
			} else if retrytimes > 3 {
				return fmt.Errorf("Generate roleId failed, retry times=%d", retrytimes)
			}
			retrytimes++
		}
		scope.SetColumn("F_id", role.F_id)
	}

	currentTime := time.Now().Local()
	nowFormat := currentTime.Format("2006-01-02 15:04:05.000")
	scope.SetColumn("F_create_time", nowFormat)
	scope.SetColumn("F_modify_time", nowFormat)
	return nil
}

func (role *Role) BeforeUpdate(scope *gorm.Scope) (err error) {
	currentTime := time.Now().Local()
	nowFormat := currentTime.Format("2006-01-02 15:04:05.000")
	scope.SetColumn("F_modify_time", nowFormat)
	return nil
}

func (role *Role) CreateRole(db *gorm.DB) (err error) {
	rdb := db.Create(role)
	return rdb.Error
}

func (role *Role) Save(db *gorm.DB) (err error) {
	rdb := db.Save(role)
	if rdb.Error != nil {
		panic("save role error:" + rdb.Error.Error())
	} else {
		err = nil
	}
	return nil
}

var (
	SceneKeyGYCORPID    = func(id uint64) string { return fmt.Sprintf("GYCORPID#%d", id) }
	SceneKeyRegGYCORPID = regexp.MustCompile(`^GYCORPID#([\d]+)$`)
)

func CheckSceneKeyFormat(scenekey string) (err error) {
	if arr := SceneKeyRegGYCORPID.FindStringSubmatch(scenekey); len(arr) == 2 {
		//log
	} else {
		return fmt.Errorf("parameter invalid, SceneKey Must be like:`" + SceneKeyRegGYCORPID.String() + "`")
	}
	return nil
}

func FindRolesByIds(db *gorm.DB, roleids []uint64) (roles []Role, err error) {
	var (
		rdb *gorm.DB
	)
	if len(roleids) == 0 {
		return
	}
	rdb = db.Where("F_id IN (?)", roleids).Find(&roles)

	if rdb.RecordNotFound() {
		return
	} else if rdb.Error != nil {
		panic("find role error:" + rdb.Error.Error())
	} else {
		err = nil
	}
	return
}

func FindRolesBySceneKey(db *gorm.DB, scenekey string) (roles []Role, err error) {
	var (
		rdb *gorm.DB
	)
	rdb = db.Where("F_scene_key = ?", scenekey).Order("F_modify_time DESC").Find(&roles)

	if rdb.RecordNotFound() {
		return
	} else if rdb.Error != nil {
		panic("find role error:" + rdb.Error.Error())
	} else {
		err = nil
	}
	return
}

func FindRolesBySceneKeyAndIds(db *gorm.DB, scenekey string, roleids []uint64) (roles []Role, err error) {
	var (
		rdb *gorm.DB
	)
	if len(roleids) == 0 {
		return
	}
	rdb = db.Where("F_scene_key = ? AND F_id IN (?)", scenekey, roleids).
		Order("F_modify_time DESC").
		Find(&roles)

	if rdb.RecordNotFound() {
		return
	} else if rdb.Error != nil {
		panic("find role error:" + rdb.Error.Error())
	} else {
		err = nil
	}
	return
}

func FindRoleByUniqueKey(db *gorm.DB, scenekey, name string) (role Role, err error) {
	var rdb *gorm.DB
	rdb = db.Where("F_name = ? AND F_scene_key = ?", name, scenekey).First(&role)
	if rdb.RecordNotFound() {
		return role, gorm.ErrRecordNotFound
	} else if rdb.Error != nil {
		panic("find privilege error:" + rdb.Error.Error())
	} else {
		err = nil
	}
	return
}

func GetRoleList(db *gorm.DB, offset, limit int) (roles []Role, count int, err error) {
	var (
		rdb *gorm.DB
	)
	rdb = db.Offset(offset).Limit(limit).Order("F_modify_time DESC").Find(&roles)

	if rdb.RecordNotFound() {
		return
	} else if rdb.Error != nil {
		panic("find privilege error:" + rdb.Error.Error())
	} else {
		err = nil
	}
	db.Model(&Role{}).Count(&count)

	return
}
