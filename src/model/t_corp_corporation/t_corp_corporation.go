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
package t_corp_corporation

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"log"
	"strings"
	"time"
)

const (
	TABLENAME = "user_t_corp_corporation"
	TABLESQL  = "CREATE TABLE IF NOT EXISTS " + TABLENAME +
		`(
			F_id bigint(20) unsigned NOT NULL                         COMMENT '企业法人userid', 
			F_status int(4)  NOT NULL DEFAULT 0                       COMMENT '状态',
			F_status_ext varchar(64)  NOT NULL DEFAULT ""             COMMENT '状态扩展信息',
			F_type bigint NOT NULL DEFAULT 0                          COMMENT '公司类型,按bit位解释（64:保留；63:系统管理；1:广盈交易；2:广盈仓储；3:广盈供应链；...）',
			F_name varchar(128) NOT NULL DEFAULT ""                   COMMENT '公司名称',
			F_description varchar(8192) NOT NULL DEFAULT ""           COMMENT '公司简介',
			F_website varchar(128) NOT NULL DEFAULT ""                COMMENT '公司官网',
			F_logo varchar(128) NOT NULL DEFAULT ""                   COMMENT '公司图标',
			F_addr_province varchar(32) NOT NULL DEFAULT ""           COMMENT '公司地址-省',
			F_addr_city varchar(32) NOT NULL DEFAULT ""               COMMENT '公司地址-市',
			F_addr_district varchar(32) NOT NULL DEFAULT ""           COMMENT '公司地址-区',
			F_addr_detail varchar(128) NOT NULL DEFAULT ""            COMMENT '公司地址-详细地址',
			F_registered_time datetime DEFAULT '2000-01-01 00:00:00'  COMMENT '注册年份',
			F_registered_capital bigint(20) NOT NULL DEFAULT 0        COMMENT '注册资本(分)',
			F_registered_business varchar(255) NOT NULL DEFAULT ""    COMMENT '主营业务',
			F_contact_name varchar(64) NOT NULL DEFAULT ""            COMMENT '联系人',
			F_contact_phone varchar(32) NOT NULL DEFAULT ""           COMMENT '联系人电话',
			F_contact_email varchar(128) NOT NULL DEFAULT ""          COMMENT '联系人邮箱',
			F_administrator bigint(20) unsigned NOT NULL DEFAULT 0    COMMENT '管理员账号id',
	        F_creator bigint(20) unsigned NOT NULL DEFAULT 0          COMMENT '创建者',
			F_modifier bigint(20) unsigned NOT NULL DEFAULT 0         COMMENT '修改者',
			F_create_time datetime NOT NULL                           COMMENT '创建时间',
			F_modify_time datetime NOT NULL                           COMMENT '修改时间',
	
			PRIMARY KEY (F_id),
			UNIQUE KEY (F_name),
			INDEX (F_administrator),
			INDEX (F_create_time),
			INDEX (F_modify_time)
		) ENGINE=InnoDB  DEFAULT CHARSET=utf8 ;`
)

type Corporation struct {
	F_id                  uint64 `gorm:"column:F_id"`                  //企业法人userid，全局唯一
	F_status              int    `gorm:"column:F_status"`              //状态
	F_status_ext          string `gorm:"column:F_status_ext"`          //状态扩展信息
	F_type                int64  `gorm:"column:F_type"`                //公司公司类型,按bit位解释（64:保留；63:系统管理；1:广盈交易；2:广盈仓储；3:广盈供应链；...）
	F_name                string `gorm:"column:F_name"`                //公司名称
	F_description         string `gorm:"column:F_description"`         //公司简介
	F_logo                string `gorm:"column:F_logo"`                //公司图标
	F_addr_province       string `gorm:"column:F_addr_province"`       //公司地址-省
	F_addr_city           string `gorm:"column:F_addr_city"`           //公司地址-市
	F_addr_district       string `gorm:"column:F_addr_district"`       //公司地址-区
	F_addr_detail         string `gorm:"column:F_addr_detail"`         //公司地址-详细地址
	F_registered_time     string `gorm:"column:F_registered_time"`     //注册年份
	F_registered_capital  int64  `gorm:"column:F_registered_capital"`  //注册资本(分)
	F_Registered_business string `gorm:"column:F_registered_business"` //主营业务
	F_website             string `gorm:"column:F_website"`             //权限参数逻辑表达式
	F_contact_name        string `gorm:"column:F_contact_name"`        //联系人
	F_contact_phone       string `gorm:"column:F_contact_phone"`       //联系人电话
	F_contact_email       string `gorm:"column:F_contact_email"`       //联系人邮箱
	F_administrator       uint64 `gorm:"column:F_administrator"`       //管理员
	F_creator             uint64 `gorm:"column:F_creator"`             //记录创建者id
	F_modifier            uint64 `gorm:"column:F_modifier"`            //记录修改者id
	F_create_time         string `gorm:"column:F_create_time"`         //创建时间
	F_modify_time         string `gorm:"column:F_modify_time"`         //修改时间
}

const (
	STATUS_INVALID       = 0  //非法状态，不应该存在
	STATUS_OK            = 1  //正常/新创建/草稿
	STATUS_DELETED       = 2  //已删除
	STATUS_VERIFY_WAIT   = 3  //待审核
	STATUS_VERIFY_DOING  = 4  //审核中
	STATUS_VERIFY_FAILED = 8  //审核失败
	STATUS_VERIFIED      = 9  //已审核通过
	STATUS_FORBIDDEN     = 10 //已禁止
)

func (corp *Corporation) TableName() string {
	return TABLENAME
}

func (corp *Corporation) BeforeCreate(scope *gorm.Scope) error {
	currentTime := time.Now().Local()
	nowFormat := currentTime.Format("2006-01-02 15:04:05.000")
	scope.SetColumn("F_create_time", nowFormat)
	scope.SetColumn("F_modify_time", nowFormat)
	return nil
}

func (corp *Corporation) BeforeUpdate(scope *gorm.Scope) (err error) {
	currentTime := time.Now().Local()
	nowFormat := currentTime.Format("2006-01-02 15:04:05.000")
	scope.SetColumn("F_modify_time", nowFormat)
	return nil
}

func (corp *Corporation) Save(db *gorm.DB) (err error) {
	rdb := db.Save(corp)
	if rdb.Error != nil {
		panic("save corporation error:" + rdb.Error.Error())
	} else {
		err = nil
	}
	return nil
}

var (
	_TYPE_TABLE = map[uint]string{
		1:  "交易会员",
		2:  "仓储会员",
		3:  "供应链会员",
		63: "系统管理",
		64: "保留",
	}
	_UNTYPE_TABLE = map[string]uint{}
)

func TypeToTypeStr(typ int64) string {
	var cur uint64 = uint64(typ)
	typestrarr := []string{}
	for i := uint(1); i <= 64; i++ {
		if (0x01 & cur) != 0 {
			if typestr, ok := _TYPE_TABLE[i]; ok {
				typestrarr = append(typestrarr, typestr)
			} else {
				log.Fatalf("UNKOWN type bit:(%d's bit %dth)", typ, i)
			}
		}
		cur = cur >> 1
	}
	return strings.Join(typestrarr, ",")
}

func TypeStrToType(typestr string) int64 {
	if len(_UNTYPE_TABLE) == 0 {
		for i, t := range _TYPE_TABLE {
			_UNTYPE_TABLE[t] = i
		}
	}

	ret := int64(0)
	typestrarr := strings.Split(typestr, ",")
	for _, str := range typestrarr {
		if i, ok := _UNTYPE_TABLE[str]; ok {
			ret |= 1 << i
		} else {
			log.Fatalf("UNKOWN typestr %s", str)
		}
	}
	return ret
}

type Status int

func (st *Status) String() string {
	switch *st {
	case STATUS_OK: //正常状态
		return "草稿"
	case STATUS_DELETED: //已删除
		return "已删除"
	case STATUS_VERIFY_WAIT: //待审核
		return "待审核"
	case STATUS_VERIFY_DOING: //审核中
		return "审核中"
	case STATUS_VERIFY_FAILED: //审核失败
		return "审核失败"
	case STATUS_VERIFIED: //已审核通过
		return "已审核通过"
	case STATUS_FORBIDDEN: //已禁止
		return "已禁用"
	}
	return fmt.Sprintf("UNKOWN_STATUS:%d", *st)
}

func (st *Status) EnableString() string {
	switch *st {
	case STATUS_OK, STATUS_VERIFY_WAIT, STATUS_VERIFY_DOING, STATUS_VERIFY_FAILED, STATUS_VERIFIED:
		return "启用"
	case STATUS_FORBIDDEN, STATUS_DELETED:
		return "禁用"
	}
	return fmt.Sprintf("UNKOWN_STATUS:%d", *st)
}

func GetEnableStringStatuses(enableString string) (statuses []Status) {
	switch enableString {
	case "启用":
		return []Status{STATUS_OK, STATUS_VERIFY_WAIT, STATUS_VERIFY_DOING, STATUS_VERIFY_FAILED, STATUS_VERIFIED}
	case "禁用":
		return []Status{STATUS_FORBIDDEN, STATUS_DELETED}
	}
	return
}

func (st *Status) VerifyString() string {
	switch *st {
	case STATUS_OK: //正常状态
		return "未提交"
	case STATUS_VERIFY_WAIT: //待审核
		return "待审核"
	case STATUS_VERIFY_DOING: //审核中
		return "审核中"
	case STATUS_VERIFY_FAILED: //审核失败
		return "审核失败"
	case STATUS_VERIFIED: //已审核通过
		return "审核通过"
	case STATUS_DELETED: //已删除
		return "已删除"
	case STATUS_FORBIDDEN: //已禁止
		return "已禁用"
	}
	return fmt.Sprintf("UNKOWN_STATUS:%d", *st)
}

func FindCorporationByName(db *gorm.DB, name string) (corp Corporation, err error) {
	var rdb *gorm.DB
	rdb = db.Where("F_name = ?", name).First(&corp)
	if rdb.RecordNotFound() {
		return corp, gorm.ErrRecordNotFound
	} else if rdb.Error != nil {
		panic("find corporation error:" + rdb.Error.Error())
	} else {
		err = nil
	}
	return
}

func FindCorporations(db *gorm.DB, conditions map[string]interface{}) (corps Corporation, err error) {
	var rdb *gorm.DB
	rdb = db.Where(conditions).Find(&corps)
	if rdb.RecordNotFound() {
		return corps, nil
	} else if rdb.Error != nil {
		panic("find corporation error:" + rdb.Error.Error())
	} else {
		err = nil
	}
	return
}

func FindCorporationById(db *gorm.DB, id uint64) (corp Corporation, err error) {
	var rdb *gorm.DB
	rdb = db.Where("F_id = ?", id).First(&corp)
	if rdb.RecordNotFound() {
		return corp, gorm.ErrRecordNotFound
	} else if rdb.Error != nil {
		panic("find corporation error:" + rdb.Error.Error())
	} else {
		err = nil
	}
	return
}

func FindCorporationsByIds(db *gorm.DB, corpids []uint64) (corps []Corporation, err error) {
	var rdb *gorm.DB
	rdb = db.Where("F_id IN (?)", corpids).Find(&corps)
	if rdb.RecordNotFound() {
		return corps, nil
	} else if rdb.Error != nil {
		panic("find corporation error:" + rdb.Error.Error())
	} else {
		err = nil
	}
	return
}
