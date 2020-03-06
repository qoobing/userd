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
package t_user_common_keyvalue

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"time"
)

const (
	TABLENAME = "user_t_user_common_keyvalue"
	TABLESQL  = "CREATE TABLE IF NOT EXISTS " + TABLENAME +
		`(
			F_user_id bigint(20) unsigned NOT NULL                     COMMENT '用户id',
			F_key   varchar(96) NOT NULL                               COMMENT '关键词，须一定复杂度，以免不同项目冲突',
			F_value varchar(4096) NOT NULL                             COMMENT '数据值，不同项目场景独立解析其含义',
			F_status int(4)  NOT NULL DEFAULT 0                        COMMENT '状态',
			F_modify_time datetime NOT NULL                            COMMENT '修改时间',
	
			PRIMARY KEY (F_user_id,F_key),
			INDEX (F_user_id)
		) ENGINE=InnoDB  DEFAULT CHARSET=utf8 ;`
)

type UserCommomKeyValue struct {
	F_user_id     uint64 `gorm:"column:F_user_id"`     //用户id
	F_key         string `gorm:"column:F_key"`         //备注信息
	F_value       string `gorm:"column:F_value"`       //备注信息
	F_status      int    `gorm:"column:F_status"`      //状态
	F_create_time string `gorm:"column:F_create_time"` //创建时间
	F_modify_time string `gorm:"column:F_modify_time"` //修改时间
}

const (
	STATUS_INVALID = 0 //非法状态，不应该存在
	STATUS_OK      = 1 //正常状态
	STATUS_DELETED = 2 //已删除
)

func (ckv *UserCommomKeyValue) TableName() string {
	return TABLENAME
}

func (ckv *UserCommomKeyValue) BeforeCreate(scope *gorm.Scope) error {
	currentTime := time.Now().Local()
	nowFormat := currentTime.Format("2006-01-02 15:04:05.000")
	scope.SetColumn("F_modify_time", nowFormat)
	return nil
}

func (ckv *UserCommomKeyValue) BeforeUpdate(scope *gorm.Scope) (err error) {
	currentTime := time.Now().Local()
	nowFormat := currentTime.Format("2006-01-02 15:04:05.000")
	scope.SetColumn("F_modify_time", nowFormat)
	return nil
}

func Get(db *gorm.DB, userid uint64, key string) (value string, err error) {
	var rdb *gorm.DB
	ckv := UserCommomKeyValue{}
	rdb = db.First(&ckv)

	if rdb.RecordNotFound() {
		return "", nil
	} else if rdb.Error != nil {
		panic("find user error:" + rdb.Error.Error())
	}

	return ckv.F_value, nil
}

func GetMulti(db *gorm.DB, userid uint64, keys []string) (values map[string]string, err error) {
	var rdb *gorm.DB
	ckvs := []UserCommomKeyValue{}
	rdb = db.Find(&ckvs)

	if rdb.RecordNotFound() || len(ckvs) == 0 {
		return values, nil
	} else if rdb.Error != nil {
		panic("find user error:" + rdb.Error.Error())
	}
	values = map[string]string{}
	for _, ckv := range ckvs {
		values[ckv.F_key] = ckv.F_value
	}

	return values, nil
}

//func SetIfNotExist(db *gorm.DB, userid uint64, key, value string) (err error){
//
// }

func Set(db *gorm.DB, userid uint64, key, value string) (err error) {
	var rdb *gorm.DB
	currentTime := time.Now().Local()
	nowFormat := currentTime.Format("2006-01-02 15:04:05.000")

	sqlstr := "INSERT INTO " + TABLENAME +
		"(F_user_id, F_key, F_value, F_modify_time) VALUES (?, ?, ?, ?) ON DUPLICATE KEY UPDATE F_user_id = ?, F_key = ?"

	rdb = db.Exec(sqlstr, userid, key, value, nowFormat, userid, key)

	if rdb.Error != nil {
		panic("set 'common key value' error:" + rdb.Error.Error())
	}

	return nil
}
