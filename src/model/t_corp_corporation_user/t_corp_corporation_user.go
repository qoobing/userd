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
package t_corp_corporation_user

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"time"
)

const (
	TABLENAME = "user_t_corp_corporation_user"
	TABLESQL  = "CREATE TABLE IF NOT EXISTS " + TABLENAME +
		`(
			F_id bigint(20) unsigned NOT NULL AUTO_INCREMENT           COMMENT '唯一id', 
			F_status int(4)  NOT NULL DEFAULT 0                        COMMENT '状态',
			F_status_ext varchar(64)  NOT NULL DEFAULT ""              COMMENT '状态扩展信息',
			F_corporation_id bigint(20) unsigned NOT NULL              COMMENT '企业法人userid',
			F_corporation_type bigint(20) unsigned NOT NULL DEFAULT 0  COMMENT '企业类型，冗余企业表的type字段',
			F_user_id bigint(20) unsigned NOT NULL                     COMMENT '企业员工userid',
			F_description varchar(128) NOT NULL DEFAULT ""             COMMENT '备注信息',
			F_create_time datetime NOT NULL                            COMMENT '创建时间',
			F_modify_time datetime NOT NULL                            COMMENT '修改时间',
	
			PRIMARY KEY (F_id),
			UNIQUE KEY (F_corporation_id, F_user_id),
			INDEX (F_corporation_id),
			INDEX (F_user_id)
		) ENGINE=InnoDB  DEFAULT CHARSET=utf8 ;`
)

type CorporationUser struct {
	F_id               uint64 `gorm:"column:F_id;primary_key"`   //唯一id
	F_status           int    `gorm:"column:F_status"`           //状态
	F_status_ext       string `gorm:"column:F_status_ext"`       //状态扩展信息
	F_corporation_id   uint64 `gorm:"column:F_corporation_id"`   //企业法人userid
	F_corporation_type int64  `gorm:"column:F_corporation_type"` //企业类型，冗余企业表的type字段
	F_user_id          uint64 `gorm:"column:F_user_id"`          //企业员工userid
	F_description      string `gorm:"column:F_description"`      //备注信息
	F_create_time      string `gorm:"column:F_create_time"`      //创建时间
	F_modify_time      string `gorm:"column:F_modify_time"`      //修改时间
}

const (
	STATUS_INVALID   = 0  //非法状态，不应该存在
	STATUS_OK        = 1  //正常状态
	STATUS_DELETED   = 2  //已删除
	STATUS_FORBIDDEN = 10 //已禁止
)

func (corpuser *CorporationUser) TableName() string {
	return TABLENAME
}

func (corpuser *CorporationUser) BeforeCreate(scope *gorm.Scope) error {
	currentTime := time.Now().Local()
	nowFormat := currentTime.Format("2006-01-02 15:04:05.000")
	scope.SetColumn("F_create_time", nowFormat)
	scope.SetColumn("F_modify_time", nowFormat)
	return nil
}

func (corpuser *CorporationUser) BeforeUpdate(scope *gorm.Scope) (err error) {
	currentTime := time.Now().Local()
	nowFormat := currentTime.Format("2006-01-02 15:04:05.000")
	scope.SetColumn("F_modify_time", nowFormat)
	return nil
}

func (corpuser *CorporationUser) Save(db *gorm.DB) (err error) {
	rdb := db.Save(corpuser)
	if rdb.Error != nil {
		panic("save corporation user error:" + rdb.Error.Error())
	} else {
		err = nil
	}
	return nil
}

func (corpuser *CorporationUser) GetCorporationUserStatusStr() string {
	switch corpuser.F_status {
	case STATUS_OK:
		return "启用"
	case STATUS_FORBIDDEN:
		return "禁用"
	}
	return ""
}

func GetCorporationUsers(db *gorm.DB, corpid uint64, offset, limit int) (corpusers []CorporationUser, err error) {
	var (
		rdb *gorm.DB
	)
	rdb = db.Where("F_status IN (?) AND F_corporation_id = ?", []int{STATUS_OK, STATUS_FORBIDDEN}, corpid).
		Order("F_modify_time DESC").
		Offset(offset).
		Limit(limit).
		Find(&corpusers)

	if rdb.RecordNotFound() {
		return
	} else if rdb.Error != nil {
		panic("find corporation users error:" + rdb.Error.Error())
	}

	return
}

func FindCorporationUserById(db *gorm.DB, corpid, userid uint64) (corpuser CorporationUser, err error) {
	var (
		rdb *gorm.DB
	)
	rdb = db.Where("F_corporation_id = ? AND F_user_id = ?", corpid, userid).Find(&corpuser)

	if rdb.RecordNotFound() {
		return corpuser, rdb.Error
	} else if rdb.Error != nil {
		panic("find user corporation error:" + rdb.Error.Error())
	}

	return
}

func FindUserCorporations(db *gorm.DB, userid uint64) (usercorps []CorporationUser, err error) {
	var (
		rdb *gorm.DB
	)
	rdb = db.Where("F_user_id = ? AND F_status IN (?) AND F_corporation_id > 1000000", userid, []int{STATUS_OK, STATUS_FORBIDDEN}).Find(&usercorps)

	if rdb.RecordNotFound() {
		return usercorps, rdb.Error
	} else if rdb.Error != nil {
		panic("find user corporation error:" + rdb.Error.Error())
	}

	return
}
