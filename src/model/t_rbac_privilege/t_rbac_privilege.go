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
package t_rbac_privilege

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"qoobing.com/utillib.golang/log"
	"time"
)

const (
	TABLENAME = "user_t_rbac_privilege"
	TABLESQL  = "CREATE TABLE IF NOT EXISTS " + TABLENAME +
		`(
			F_id bigint(20) unsigned NOT NULL AUTO_INCREMENT, 
			F_status int(4)  NOT NULL DEFAULT 0,
			F_status_ext varchar(64)  NOT NULL DEFAULT "",
			F_name varchar(128) NOT NULL DEFAULT "",
			F_description varchar(128) NOT NULL DEFAULT "",
			F_uri varchar(128) NOT NULL DEFAULT "",
			F_expression varchar(128) NOT NULL DEFAULT "",
			F_creator bigint(20) unsigned NOT NULL DEFAULT 0, 
			F_modifier bigint(20) unsigned NOT NULL DEFAULT 0, 
			F_create_time datetime NOT NULL,
			F_modify_time datetime NOT NULL,
	
			PRIMARY KEY (F_id),
			UNIQUE KEY (F_name,F_uri,F_expression),
			INDEX (F_uri)
		) ENGINE=InnoDB  DEFAULT CHARSET=utf8 ;`
)

type Privilege struct {
	F_id          uint64 `gorm:"column:F_id;primary_key"` //权限id，全局唯一
	F_status      int    `gorm:"column:F_status"`         //状态
	F_status_ext  string `gorm:"column:F_status_ext"`     //状态扩展信息
	F_name        string `gorm:"column:F_name"`           //权限名称
	F_description string `gorm:"column:F_description"`    //权限描述
	F_uri         string `gorm:"column:F_uri"`            //uri正则表达式
	F_expression  string `gorm:"column:F_expression"`     //权限参数逻辑表达式
	F_creator     uint64 `gorm:"column:F_creator"`        //记录创建者id
	F_modifier    uint64 `gorm:"column:F_modifier"`       //记录修改者id
	F_create_time string `gorm:"column:F_create_time"`    //创建时间
	F_modify_time string `gorm:"column:F_modify_time"`    //修改时间
}

const (
	PRIVILEGE_STATUS_INVALID   = 0  //非法状态，不应该存在
	PRIVILEGE_STATUS_OK        = 1  //正常状态
	PRIVILEGE_STATUS_DELETED   = 2  //已删除
	PRIVILEGE_STATUS_FORBIDDEN = 10 //已禁止
)

func (privilege *Privilege) TableName() string {
	return TABLENAME
}

func (privilege *Privilege) BeforeCreate(scope *gorm.Scope) error {
	currentTime := time.Now().Local()
	nowFormat := currentTime.Format("2006-01-02 15:04:05.000")
	scope.SetColumn("F_create_time", nowFormat)
	scope.SetColumn("F_modify_time", nowFormat)
	return nil
}

func (privilege *Privilege) BeforeUpdate(scope *gorm.Scope) (err error) {
	currentTime := time.Now().Local()
	nowFormat := currentTime.Format("2006-01-02 15:04:05.000")
	scope.SetColumn("F_modify_time", nowFormat)
	return nil
}

func (privilege *Privilege) Save(db *gorm.DB) (err error) {
	rdb := db.Save(privilege)
	if rdb.Error != nil {
		panic("save privilege error:" + rdb.Error.Error())
	} else {
		err = nil
	}
	return nil
}

func FindPrivilegeByUniqueKey(db *gorm.DB, name, uri, expression string) (privilege Privilege, err error) {
	var rdb *gorm.DB
	rdb = db.Where("F_name = ? AND F_uri = ? AND F_expression = ?", name, uri, expression).First(&privilege)
	if rdb.RecordNotFound() {
		return privilege, gorm.ErrRecordNotFound
	} else if rdb.Error != nil {
		panic("find privilege error:" + rdb.Error.Error())
	} else {
		err = nil
	}
	return
}

func FindPrivilegeByName(db *gorm.DB, name string) (privilege Privilege, err error) {
	var rdb *gorm.DB
	rdb = db.Where("F_name = ?", name).First(&privilege)
	if rdb.RecordNotFound() {
		return privilege, gorm.ErrRecordNotFound
	} else if rdb.Error != nil {
		panic("find privilege error:" + rdb.Error.Error())
	} else {
		err = nil
	}
	return
}

func FindPrivilegeByUriExpression(db *gorm.DB, uri, expression string) (privilege Privilege, err error) {
	var rdb *gorm.DB
	rdb = db.Where("F_uri = ? AND F_expression = ?", uri, expression).First(&privilege)
	if rdb.RecordNotFound() {
		return privilege, gorm.ErrRecordNotFound
	} else if rdb.Error != nil {
		panic("find privilege error:" + rdb.Error.Error())
	} else {
		err = nil
	}
	return
}

func FindPrivilegeIdsByUriAndArgs(db *gorm.DB, uri, args string) (ids []string, err error) {
	var (
		rdb        *gorm.DB
		privileges = []Privilege{}
	)
	rdb = db.Where("F_uri = ?", uri).Find(&privileges)
	log.Fatalf("UNFINISHED")
	if rdb.RecordNotFound() {
		return nil, gorm.ErrRecordNotFound
	} else if rdb.Error != nil {
		panic("find privilege error:" + rdb.Error.Error())
	} else {
		err = nil
	}
	return
}

func FindPrivilegeByIds(db *gorm.DB, ids []uint64) (privileges []Privilege, err error) {
	var (
		rdb *gorm.DB
	)
	rdb = db.Where("F_id IN (?)", ids).Find(&privileges)

	if rdb.RecordNotFound() {
		return
	} else if rdb.Error != nil {
		panic("find privilege error:" + rdb.Error.Error())
	} else {
		err = nil
	}
	return
}

func GetPrivilegeList(db *gorm.DB, offset, limit int) (privileges []Privilege, count int, err error) {
	var (
		rdb *gorm.DB
	)
	rdb = db.Offset(offset).Limit(limit).Order("F_modify_time DESC").Find(&privileges)

	if rdb.RecordNotFound() {
		return
	} else if rdb.Error != nil {
		panic("find privilege error:" + rdb.Error.Error())
	} else {
		err = nil
	}
	db.Model(&Privilege{}).Count(&count)

	return
}
