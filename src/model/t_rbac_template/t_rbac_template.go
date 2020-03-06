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
package t_rbac_template

import (
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"time"
)

const (
	TABLENAME = "user_t_rbac_template"
	TABLESQL  = "CREATE TABLE IF NOT EXISTS " + TABLENAME +
		`(
			F_id bigint(20) unsigned NOT NULL AUTO_INCREMENT, 
			F_status int(4)  NOT NULL DEFAULT 0,
			F_description varchar(128) NOT NULL DEFAULT "",
			F_role_tree text NOT NULL,
			F_create_time datetime NOT NULL,
			F_modify_time datetime NOT NULL,
	
			PRIMARY KEY (F_id)
		) ENGINE=InnoDB  DEFAULT CHARSET=utf8 ;`
)

type Template struct {
	F_id          uint64 `gorm:"column:F_id;primary_key"` //全局唯一id
	F_status      int    `gorm:"column:F_status"`         //状态
	F_description string `gorm:"column:F_description"`    //模版描述
	F_role_tree   string `gorm:"column:F_role_tree"`      //模版角色树
	F_create_time string `gorm:"column:F_create_time"`    //创建时间
	F_modify_time string `gorm:"column:F_modify_time"`    //修改时间
}

type TemplateRoleTree struct {
	Id       uint64             `json:"id"`
	Type     int                `json:"type"`
	Name     string             `json:"name"`
	Selected bool               `json:"selected"`
	Children []TemplateRoleTree `json:"children"`
}

const (
	STATUS_INVALID   = 0  //非法状态，不应该存在
	STATUS_OK        = 1  //正常状态
	STATUS_DELETED   = 2  //已删除
	STATUS_FORBIDDEN = 10 //已禁止
)

func (template *Template) TableName() string {
	return TABLENAME
}

func (template *Template) BeforeCreate(scope *gorm.Scope) error {
	currentTime := time.Now().Local()
	nowFormat := currentTime.Format("2006-01-02 15:04:05.000")
	scope.SetColumn("F_create_time", nowFormat)
	scope.SetColumn("F_modify_time", nowFormat)
	return nil
}

func (template *Template) BeforeUpdate(scope *gorm.Scope) (err error) {
	currentTime := time.Now().Local()
	nowFormat := currentTime.Format("2006-01-02 15:04:05.000")
	scope.SetColumn("F_modify_time", nowFormat)
	return nil
}

func (template *Template) Save(db *gorm.DB) (err error) {
	rdb := db.Save(template)
	if rdb.Error != nil {
		panic("save privilege error:" + rdb.Error.Error())
	} else {
		err = nil
	}
	return nil
}

func ReadTemplateRoleTreeFromDb(db *gorm.DB, templateId uint64) (templateRoleTree TemplateRoleTree, err error) {
	var (
		rdb      *gorm.DB
		template Template
	)

	rdb = db.Where("F_id = ? AND F_status = ?", templateId, STATUS_OK).First(&template)
	if rdb.RecordNotFound() {
		return templateRoleTree, gorm.ErrRecordNotFound
	} else if rdb.Error != nil {
		panic("find privilege error:" + rdb.Error.Error())
	}

	if unerr := json.Unmarshal([]byte(template.F_role_tree), &templateRoleTree); unerr != nil {
		err = fmt.Errorf("ReadTemplateRoleTreeFromDb error:%s", unerr.Error())
		return templateRoleTree, err
	}

	return templateRoleTree, nil
}
