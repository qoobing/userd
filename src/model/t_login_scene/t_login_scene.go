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
package t_login_scene

import (
	"errors"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

const (
	TABLENAME = "user_t_login_scene"
	TABLESQL  = "CREATE TABLE IF NOT EXISTS " + TABLENAME +
		`(
			F_scene_id bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '用户登陆场景id，全局唯一',
			F_scene_name varchar(64) NOT NULL DEFAULT '' COMMENT '场景名称，全局唯一',
			F_redirect_url varchar(128) NOT NULL DEFAULT '' COMMENT '登陆成功重定向url',
			F_privilege varchar(256) NOT NULL DEFAULT '' COMMENT '场景权限',
			F_create_time datetime NOT NULL DEFAULT '2000-01-01 00:00:00' COMMENT '创建时间',
			F_modify_time datetime NOT NULL DEFAULT '2000-01-01 00:00:00' COMMENT '修改时间',
			PRIMARY KEY (F_scene_id),
			UNIQUE KEY I_scene_name (F_scene_name)
		) ENGINE=InnoDB DEFAULT CHARSET=gbk;`
)

type LoginScene struct {
	F_scene_id     uint64 `gorm:"column:F_scene_id"`     //用户登陆场景id，全局唯一
	F_scene_name   string `gorm:"column:F_scene_name"`   //场景名称，全局唯一
	F_redirect_url string `gorm:"column:F_redirect_url"` //登陆成功重定向url
	F_privilege    string `gorm:"column:F_privilege"`    //场景权限
	F_create_time  string `gorm:"column:F_create_time"`  //创建时间
	F_modify_time  string `gorm:"column:F_modify_time"`  //修改时间
}

func (scene *LoginScene) TableName() string {
	return TABLENAME
}

type ATTRIBUTE_NAME string

const (
	ATTRIBUTE_NAME_AUTO_REGISTER_MOBILE_LOGIN = "auto_register_mobile_login"
)

func (scene *LoginScene) GetAttribute(name ATTRIBUTE_NAME, defaultvalue interface{}) interface{} {
	//TODO: read from database
	if name == ATTRIBUTE_NAME_AUTO_REGISTER_MOBILE_LOGIN {
		if defaultvalue != nil {
			return defaultvalue
		} else {
			return false
		}
	}
	return nil
}

func FindLoginScene(db *gorm.DB, scene_name string) (scene LoginScene, err error) {
	if scene_name == "" {
		scene_name = "default"
	}

	var rdb = db.Where("F_scene_name = ?", scene_name).First(&scene)

	if rdb.RecordNotFound() {
		err = errors.New("scene not exist")
	} else if rdb.Error != nil {
		panic("find login_scene error:" + rdb.Error.Error())
	} else {
		err = nil
	}

	return scene, err
}
