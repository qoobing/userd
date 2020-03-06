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
package t_rbac_role_map

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"qoobing.com/utillib.golang/log"
	"strconv"
	"strings"
	"time"
)

const (
	TABLENAME = "user_t_rbac_role_map"
	TABLESQL  = "CREATE TABLE IF NOT EXISTS " + TABLENAME +
		`(
			F_id bigint(20) unsigned NOT NULL AUTO_INCREMENT, 
			F_status int(4)  NOT NULL DEFAULT 0,
			F_status_ext varchar(64)  NOT NULL DEFAULT "",
			F_description varchar(128) NOT NULL DEFAULT "",
			F_role_id bigint(20) unsigned NOT NULL DEFAULT 0,
			F_target_id bigint(20) unsigned NOT NULL DEFAULT 0,
			F_target_type int(4) unsigned NOT NULL DEFAULT 0,
			F_target_ext varchar(128) NOT NULL DEFAULT "",
			F_scene_key varchar(128) NOT NULL DEFAULT "",
			F_creator bigint(20) unsigned NOT NULL DEFAULT 0, 
			F_modifier bigint(20) unsigned NOT NULL DEFAULT 0, 
			F_create_time datetime NOT NULL,
			F_modify_time datetime NOT NULL,
	
			PRIMARY KEY (F_id),
			UNIQUE KEY IDX_U_SCENE_RID_TID_TTYPE_TEXT(F_scene_key,F_role_id,F_target_id, F_target_type,F_target_ext),
			INDEX IDX_N_SCENE(F_scene_key)
		) ENGINE=InnoDB  DEFAULT CHARSET=utf8 ;`
)

type RoleMap struct {
	F_id          uint64 `gorm:"column:F_id;primary_key"` //角色id，全局唯一
	F_status      int    `gorm:"column:F_status"`         //状态
	F_status_ext  string `gorm:"column:F_status_ext"`     //状态扩展信息
	F_description string `gorm:"column:F_description"`    //描述
	F_role_id     uint64 `gorm:"column:F_role_id"`        //角色id
	F_target_type int    `gorm:"column:F_target_type"`    //映射类型
	F_target_id   uint64 `gorm:"column:F_target_id"`      //映射目标id，类型根据F_target_type确定
	F_target_ext  string `gorm:"column:F_target_ext"`     //映射目标附加参数
	F_scene_key   string `gorm:"column:F_scene_key"`      //场景key
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

	TARGET_TYPE_INVALID           = 0 //非法状态，不应该存在
	TARGET_TYPE_ROLE_TO_ROLE      = 1 //角色映射到角色
	TARGET_TYPE_ROLE_TO_PRIVILEGE = 2 //角色映射到权限
)

type MapId struct {
	Id   uint64
	Type int
}

func (rolemap *RoleMap) TableName() string {
	return TABLENAME
}

func (rolemap *RoleMap) CheckTargetType(target_type int) (err error) {
	switch target_type {
	case TARGET_TYPE_ROLE_TO_ROLE:
		return nil
	case TARGET_TYPE_ROLE_TO_PRIVILEGE:
		return nil
	default:
		err = fmt.Errorf("Inavlid target type, MUST be one of " +
			strconv.Itoa(TARGET_TYPE_ROLE_TO_ROLE) + ":TARGET_TYPE_ROLE_TO_ROLE," +
			strconv.Itoa(TARGET_TYPE_ROLE_TO_PRIVILEGE) + ":TARGET_TYPE_ROLE_TO_PRIVILEGE")
		return err
	}
	return
}

func (rolemap *RoleMap) BeforeCreate(scope *gorm.Scope) error {
	currentTime := time.Now().Local()
	nowFormat := currentTime.Format("2006-01-02 15:04:05.000")
	scope.SetColumn("F_create_time", nowFormat)
	scope.SetColumn("F_modify_time", nowFormat)
	return nil
}

func (rolemap *RoleMap) BeforeUpdate(scope *gorm.Scope) (err error) {
	currentTime := time.Now().Local()
	nowFormat := currentTime.Format("2006-01-02 15:04:05.000")
	scope.SetColumn("F_modify_time", nowFormat)
	return nil
}

func (rolemap *RoleMap) Save(db *gorm.DB) (err error) {
	rdb := db.Save(rolemap)
	if rdb.Error != nil {
		panic("save rolemap error:" + rdb.Error.Error())
	} else {
		err = nil
	}
	return nil
}

func CreateRoleMaps(db *gorm.DB, rolemaps []*RoleMap) (err error) {
	if len(rolemaps) <= 0 {
		log.Debugf("Nothing need to do with [%s]", "CreateRoleMaps")
		return nil
	}

	fieldsNum := 12
	valueStrings := make([]string, 0, len(rolemaps))
	valueArgs := make([]interface{}, 0, len(rolemaps)*fieldsNum)
	valueLocat := "?"
	for i := 1; i < fieldsNum; i++ {
		valueLocat += ",?"
	}
	valueLocat = fmt.Sprintf("(%s)", valueLocat)
	currentTime := time.Now().Local()
	nowFormat := currentTime.Format("2006-01-02 15:04:05.000")

	for _, rms := range rolemaps {
		valueStrings = append(valueStrings, valueLocat)
		valueArgs = append(valueArgs, rms.F_status)
		valueArgs = append(valueArgs, rms.F_status_ext)
		valueArgs = append(valueArgs, rms.F_description)
		valueArgs = append(valueArgs, rms.F_role_id)
		valueArgs = append(valueArgs, rms.F_target_type)
		valueArgs = append(valueArgs, rms.F_target_id)
		valueArgs = append(valueArgs, rms.F_target_ext)
		valueArgs = append(valueArgs, rms.F_scene_key)
		valueArgs = append(valueArgs, rms.F_creator)
		valueArgs = append(valueArgs, rms.F_modifier)
		valueArgs = append(valueArgs, nowFormat)
		valueArgs = append(valueArgs, nowFormat)
	}
	stmt := fmt.Sprintf(`INSERT INTO %s (
			F_status,
			F_status_ext,
			F_description,
			F_role_id,
			F_target_type,
			F_target_id,
			F_target_ext,
			F_scene_key,
			F_creator,
			F_modifier,
			F_create_time,
			F_modify_time
		) VALUES %s`,
		TABLENAME,
		strings.Join(valueStrings, ","))

	rdb := db.Exec(stmt, valueArgs...)
	if rdb.Error != nil {
		panic("create rolemaps error:" + rdb.Error.Error())
	} else {
		err = nil
	}

	//for _, rolemap := range rolemaps {
	//	rdb := db.Create(rolemap)
	//	if rdb.Error != nil {
	//		panic("create rolemap error:" + rdb.Error.Error())
	//	} else {
	//		err = nil
	//	}
	//}
	return nil
}

func UpdateRoleMapsToStatus(db *gorm.DB, rolemaps []*RoleMap) (err error) {
	if len(rolemaps) <= 0 {
		log.Debugf("Nothing need to do with [%s]", "UpdateRoleMapsToStatus")
		return nil
	}
	var (
		ids         = []uint64{}
		currentTime = time.Now().Local()
		target      = map[string]interface{}{
			"F_status":      rolemaps[0].F_status,
			"F_modifier":    rolemaps[0].F_modifier,
			"F_modify_time": currentTime.Format("2006-01-02 15:04:05.000"),
		}
	)

	for _, rm := range rolemaps {
		ids = append(ids, rm.F_id)
	}
	rdb := db.Model(&RoleMap{}).Where("F_id IN (?)", ids).Update(target)

	if rdb.Error != nil {
		panic("UpdateRoleMapsToStatus error:" + rdb.Error.Error())
	} else {
		err = nil
	}
	return nil
}

func FindChildrenMap(db *gorm.DB, roleids []uint64) (rolemap []RoleMap, err error) {
	var rdb *gorm.DB
	rdb = db.Where("F_role_id IN (?) AND F_status = ?", roleids, ROLE_STATUS_OK).Find(&rolemap)
	if rdb.RecordNotFound() {
		return rolemap, nil
	} else if rdb.Error != nil {
		panic("find rolemap error:" + rdb.Error.Error())
	} else {
		err = nil
	}
	return
}

func FindParentsMap(db *gorm.DB, ids []MapId) (rolemaps []RoleMap, err error) {
	if len(ids) == 0 {
		return
	}
	var rdb *gorm.DB
	var tids = []uint64{}
	for _, id := range ids {
		tids = append(tids, id.Id)
	}
	rdb = db.Where("F_target_id IN (?) AND F_target_type = ?", tids, ids[0].Type).Find(&rolemaps)
	if rdb.RecordNotFound() {
		return rolemaps, nil
	} else if rdb.Error != nil {
		panic("find rolemap error:" + rdb.Error.Error())
	} else {
		err = nil
	}
	return
}

func FindPrivilegeParents(db *gorm.DB, id uint64) (rolemaps []RoleMap, err error) {
	var rdb *gorm.DB
	rdb = db.Where("F_target_id = ? AND F_target_type = ?", id, TARGET_TYPE_ROLE_TO_PRIVILEGE).Find(&rolemaps)
	if rdb.RecordNotFound() {
		return rolemaps, nil
	} else if rdb.Error != nil {
		panic("find rolemap error:" + rdb.Error.Error())
	} else {
		err = nil
	}
	return
}

func FindRoleParents(db *gorm.DB, id uint64) (rolemaps []RoleMap, err error) {
	var rdb *gorm.DB
	rdb = db.Where("F_target_id = ? AND F_target_type = ?", id, TARGET_TYPE_ROLE_TO_ROLE).Find(&rolemaps)
	if rdb.RecordNotFound() {
		return rolemaps, nil
	} else if rdb.Error != nil {
		panic("find rolemap error:" + rdb.Error.Error())
	} else {
		err = nil
	}
	return
}
