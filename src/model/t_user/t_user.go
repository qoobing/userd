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
package t_user

import (
	"encoding/json"
	"errors"
	"github.com/jinzhu/gorm"
	"github.com/qoobing/userd/src/common"
	. "github.com/qoobing/userd/src/const"
	"github.com/qoobing/userd/src/types"
	. "qoobing.com/utillib.golang/api"
	"qoobing.com/utillib.golang/gls"
	"qoobing.com/utillib.golang/log"
	. "qoobing.com/utillib.golang/xyz"

	"crypto/sha256"
	"encoding/hex"
	"fmt"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"strconv"
	"time"
)

const (
	TABLENAME = "user_t_user"
	TABLESQL  = "CREATE TABLE IF NOT EXISTS " + TABLENAME +
		`(
			F_user_id bigint(20) unsigned NOT NULL                    COMMENT '用户id，全局唯一',
			F_user_ids varchar(64) NOT NULL DEFAULT ''                COMMENT '关联用户id列表[一般应用可不予考虑]',
			F_status tinyint(4) NOT NULL DEFAULT '1'                  COMMENT '状态',
			F_status_ext varchar(64) NOT NULL DEFAULT ''              COMMENT '状态扩展信息',
			F_nickname varchar(64) NOT NULL DEFAULT ''                COMMENT '用户昵称',
			F_name varchar(64) NOT NULL                               COMMENT '用户名',
			F_mobile varchar(64) NOT NULL DEFAULT ''                  COMMENT '用户手机号',
			F_email varchar(64) NOT NULL DEFAULT ''                   COMMENT '用户电子邮箱',
			F_avatar varchar(256) NOT NULL DEFAULT ''                 COMMENT '用户头像',
			F_address varchar(256) NOT NULL DEFAULT ''                COMMENT '用户地址，用|进行行政区分级',
			F_birthday datetime DEFAULT '2000-01-01 00:00:00'         COMMENT '用户生日',
			F_sec_password varchar(64) NOT NULL DEFAULT ''            COMMENT '密码',
			F_sec_salt varchar(32) NOT NULL DEFAULT ''                COMMENT '密码盐值',
			F_sec_answers varchar(512) NOT NULL DEFAULT ''            COMMENT '密保问题及答案',
			F_exid_wx varchar(64) NOT NULL DEFAULT ''                 COMMENT '用户外部id：微信号',
			F_exid_wx_openid varchar(64) NOT NULL DEFAULT ''          COMMENT '用户外部id：微信openid',
			F_exid_wx_unionid varchar(64) NOT NULL DEFAULT ''         COMMENT '用户外部id：微信unionid',
			F_exid_qq varchar(64) NOT NULL DEFAULT ''                 COMMENT '用户外部id：QQ号',
			F_exid_qq_openid varchar(64) NOT NULL DEFAULT ''          COMMENT '用户外部id：QQ openid',
			F_exid_qq_unionid varchar(64) NOT NULL DEFAULT ''         COMMENT '用户外部id：QQ unionid',
			F_exid_dd varchar(64) NOT NULL DEFAULT ''                 COMMENT '用户外部id：钉钉号',
			F_exid_dd_openid varchar(64) NOT NULL DEFAULT ''          COMMENT '用户外部id：钉钉openid',
			F_exid_dd_unionid varchar(64) NOT NULL DEFAULT ''         COMMENT '用户外部id：钉钉unionid',
			F_create_time datetime DEFAULT '2000-01-01 00:00:00'      COMMENT '创建时间',
			F_modify_time datetime DEFAULT '2000-01-01 00:00:00'      COMMENT '修改时间',
			
			PRIMARY KEY (F_user_id),
			KEY I_name (F_name),
			KEY I_mobile (F_mobile),
			KEY I_email (F_email)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8;`
)

type User struct {
	F_user_id         uint64 `gorm:"column:F_user_id"`         //用户id，全局唯一
	F_user_ids        string `gorm:"column:F_user_ids"`        //关联用户id列表【一般应用可不予考虑】
	F_status          int    `gorm:"column:F_status"`          //用户状态
	F_status_ext      string `gorm:"column:F_status_ext"`      //用户状态扩展信息
	F_nickname        string `gorm:"column:F_nickname"`        //用户昵称
	F_name            string `gorm:"column:F_name"`            //用户名
	F_mobile          string `gorm:"column:F_mobile"`          //用户手机
	F_email           string `gorm:"column:F_email"`           //用户电子邮箱
	F_avatar          string `gorm:"column:F_avatar"`          //用户头像
	F_address         string `gorm:"column:F_address"`         //用户地址，用|进行行政区分级
	F_birthday        string `gorm:"column:F_birthday"`        //用户生日
	F_sec_password    string `gorm:"column:F_sec_password"`    //密码
	F_sec_salt        string `gorm:"column:F_sec_salt"`        //密码盐值
	F_sec_answers     string `gorm:"column:F_sec_answers"`     //密保问题及答案
	F_exid_wx         string `gorm:"column:F_exid_wx"`         //用户外部id：微信号
	F_exid_wx_openid  string `gorm:"column:F_exid_wx_openid"`  //用户外部id：微信openid
	F_exid_wx_unionid string `gorm:"column:F_exid_wx_unionid"` //用户外部id：微信unionid
	F_exid_qq         string `gorm:"column:F_exid_qq"`         //用户外部id：QQ号
	F_exid_qq_openid  string `gorm:"column:F_exid_qq_openid"`  //用户外部id：QQ openid
	F_exid_qq_unionid string `gorm:"column:F_exid_qq_unionid"` //用户外部id：QQ unionid
	F_exid_dd         string `gorm:"column:F_exid_dd"`         //用户外部id：钉钉号
	F_exid_dd_openid  string `gorm:"column:F_exid_dd_openid"`  //用户外部id：钉钉openid
	F_exid_dd_unionid string `gorm:"column:F_exid_dd_unionid"` //用户外部id：钉钉unionid
	F_create_time     string `gorm:"column:F_create_time"`     //创建时间
	F_modify_time     string `gorm:"column:F_modify_time"`     //修改时间
}

const (
	USER_STATUS_INVALID   = 0   //非法状态，不应该存在
	USER_STATUS_OK        = 1   //正常状态
	USER_STATUS_DELETED   = 2   //正常状态
	USER_STATUS_FORBIDDEN = 10  //已禁止
	USER_STATUS_MERGED    = 100 //已合并到主账号状态，F_status_ext指示出合并到的主账号
)

var (
	ERROR_USER_NOT_EXIST     = errors.New(USER_NOT_EXIST)
	ERROR_USER_MORE_THAN_ONE = errors.New(USER_FOUND_MORE_THAN_ONE)
)

func (u *User) TableName() string {
	return "t_user"
}

func (user *User) BeforeCreate(scope *gorm.Scope) error {
	currentTime := time.Now().Local()
	nowFormat := currentTime.Format("2006-01-02 15:04:05.000")
	scope.SetColumn("F_user_id", common.GenerateId("USER_ID"))
	scope.SetColumn("F_create_time", nowFormat)
	scope.SetColumn("F_modify_time", nowFormat)
	return nil
}

func (user *User) BeforeUpdate(scope *gorm.Scope) (err error) {
	currentTime := time.Now().Local()
	nowFormat := currentTime.Format("2006-01-02 15:04:05.000")
	scope.SetColumn("F_modify_time", nowFormat)
	return nil
}

func FindUser(db *gorm.DB, inputname string, nametype int) (user User, err error) {
	var rdb *gorm.DB
	var nm = inputname
	if nm == "" {
		return user, ERROR_USER_NOT_EXIST
	}
	switch nametype {
	case NAMETYPE_NAME: //name
		rdb = db.Where("F_name = ?", nm)
	case NAMETYPE_MOBILE: //mobile
		rdb = db.Where("F_mobile = ?", nm)
	case NAMETYPE_EMAIL: //email
		rdb = db.Where("F_email = ?", nm)
	case NAMETYPE_ANY: //all
		rdb = db.Where("F_name = ? OR F_mobile = ? OR F_email = ?", nm, nm, nm)
	default:
		panic("find user error: unknown nametype")
	}
	users := []User{}
	rdb = rdb.Limit(10).Find(&users)

	if rdb.RecordNotFound() || len(users) <= 0 {
		return user, ERROR_USER_NOT_EXIST
	} else if rdb.Error != nil {
		panic("find user error:" + rdb.Error.Error())
	}

	if len(users) > 1 {
		return user, ERROR_USER_MORE_THAN_ONE
	}
	user = users[0]

	return user, nil
}

func FindUserByAuth20id(db *gorm.DB, id string, idtype int) (user User, err error) {
	var rdb *gorm.DB

	switch idtype {
	case AUTH20ID_TYPE_WX_OPENID:
		rdb = db.Where("F_exid_wx_openid = ?", id).First(&user)
	case AUTH20ID_TYPE_WX_UNIONID:
		rdb = db.Where("F_exid_wx_unionid = ?", id).First(&user)
	case AUTH20ID_TYPE_QQ_OPENID:
		rdb = db.Where("F_exid_qq_openid = ?", id).First(&user)
	case AUTH20ID_TYPE_QQ_UNIONID:
		rdb = db.Where("F_exid_qq_unionid = ?", id).First(&user)
	case AUTH20ID_TYPE_DING_OPENID:
		rdb = db.Where("F_exid_dd_openid = ?", id).First(&user)
	case AUTH20ID_TYPE_DING_UNIONID:
		rdb = db.Where("F_exid_dd_unionid = ?", id).First(&user)
	default:
		panic("find user error: unknown idtype:" + strconv.Itoa(idtype))
	}

	if rdb.RecordNotFound() {
		err = ERROR_USER_NOT_EXIST
	} else if rdb.Error != nil {
		panic("find user error:" + rdb.Error.Error())
	} else {
		err = nil
	}

	return user, err
}

func CreateUser(db *gorm.DB, user *User) (err error) {
	ASSERT(user.F_user_id == 0, "create user must have no userid")
	ASSERT(user.F_name != "", "create user must have user 'name'")

	if user.F_sec_password != "" {
		user.F_sec_salt = GetRandomString(30)
		str := user.F_sec_password + user.F_sec_salt
		h := sha256.New()
		h.Write([]byte(str))
		user.F_sec_password = hex.EncodeToString(h.Sum(nil))
	}

	if user.F_status == USER_STATUS_INVALID {
		user.F_status = USER_STATUS_OK
	}
	if user.F_birthday == "" {
		user.F_birthday = "1987-08-11 10:09:08"
	}

	rdb := db.Create(user)
	return rdb.Error
}

func GetUserByUserid(db *gorm.DB, id uint64) (user User, err error) {
	var rdb *gorm.DB
	rdb = db.Where("F_user_id = ?", id).First(&user)

	if rdb.RecordNotFound() {
		err = ERROR_USER_NOT_EXIST
	} else if rdb.Error != nil {
		panic("find user error:" + rdb.Error.Error())
	} else {
		err = nil
	}

	return user, err
}

func GetUsersByUserids(db *gorm.DB, ids []uint64) (users []User, err error) {
	var rdb *gorm.DB
	rdb = db.Where("F_user_id IN (?)", ids).Find(&users)

	if rdb.RecordNotFound() {
		err = nil
	} else if rdb.Error != nil {
		panic("find user error:" + rdb.Error.Error())
	}

	return users, err
}

func UpdateUserPassword(db *gorm.DB, user User) (err error) {
	ASSERT(user.F_user_id != 0, "update user info must have userid")

	user.F_sec_salt = GetRandomString(30)
	str := user.F_sec_password + user.F_sec_salt
	h := sha256.New()
	h.Write([]byte(str))
	user.F_sec_password = hex.EncodeToString(h.Sum(nil))
	u := map[string]interface{}{
		"F_sec_password": user.F_sec_password,
		"F_sec_salt":     user.F_sec_salt,
	}

	rdb := db.Model(&user).Where("F_user_id = ?", user.F_user_id).Update(u)
	return rdb.Error
}

func UpdateUserMobile(db *gorm.DB, user User) (err error) {
	ASSERT(user.F_user_id != 0, "update user info must have userid")

	u := map[string]interface{}{
		"F_mobile": user.F_mobile,
	}

	rdb := db.Model(&user).Where("F_user_id = ?", user.F_user_id).Update(u)
	return rdb.Error
}

func UpdateUserEmail(db *gorm.DB, user User) (err error) {
	ASSERT(user.F_user_id != 0, "update user info must have userid")

	u := map[string]interface{}{
		"F_email": user.F_email,
	}

	rdb := db.Model(&user).Where("F_user_id = ?", user.F_user_id).Update(u)
	return rdb.Error
}

func GetVcodeTypeName(vcodetype int) (vcodetypename string) {
	switch vcodetype {
	case VCODEADDRESSTYPE_MOBILE_LOGIN:
		return "MOBILE_LOGIN_VCODE"
	case VCODEADDRESSTYPE_EMAIL_LOGIN:
		return "EMAIL_LOGIN_VCODE"
	case VCODEADDRESSTYPE_MOBILE_REGISTER:
		return "MOBILE_REGISTER_VCODE"
	case VCODEADDRESSTYPE_EMAIL_REGISTER:
		return "EMAIL_REGISTER_VCODE"
	case VCODEADDRESSTYPE_MOBILE_PASSWORD:
		return "MOBILE_PASSWORD_VCODE"
	case VCODEADDRESSTYPE_EMAIL_PASSWORD:
		return "EMAIL_PASSWORD_VCODE"
	case VCODEADDRESSTYPE_SAFE_MOBILE_PASSWORD:
		return "SAFE_MOBILE_PASSWORD_VCODE"
	case VCODEADDRESSTYPE_SAFE_EMAIL_PASSWORD:
		return "SAFE_EMAIL_PASSWORD_VCODE"
	}
	panic(fmt.Sprintf("unkown vcodetype:%d", vcodetype))
}

func VerifyUserPassword(user User, sha256pwd string) bool {
	log.Debugf("VerifyUserPassword, userid:%d", user.F_user_id)
	str := sha256pwd + user.F_sec_salt
	h := sha256.New()
	h.Write([]byte(str))

	if hex.EncodeToString(h.Sum(nil)) == user.F_sec_password {
		log.Debugf("verify user access by password success")
		return true
	}
	return false
}

func VerifyMobileVcode(mobile, vcodekey, vcode string, vcodetype int) (pass bool) {
	log.Debugf("VerifyMobileVcode")
	rds := gls.GetGlsValueNotNil("redis").(*RedisConn)
	rdskey := "USER_MOBILE_VCODE-" + GetVcodeTypeName(vcodetype) + "-" + vcodekey
	jsondata, err := rds.Do("GET", rdskey)
	if err != nil {
		log.Debugf("verify mobile vcode failed: redis GET [%s] error:%s", rdskey, err.Error())
		return false
	} else if jsondata == nil {
		log.Debugf("verify mobile vcode failed: redis GET [%s] return nil[not found]", rdskey)
		return false
	}

	var vc types.VerifyCodeInfo
	err = json.Unmarshal(jsondata.([]byte), &vc)
	if err != nil {
		log.Debugf("verify mobile vcode failed:  unmarshal error:" + err.Error())
		return false
	} else if vc.Name != mobile {
		log.Debugf("verify mobile vcode failed: input name not match saved name")
		return false
	} else if vc.Vcode != vcode {
		log.Debugf("verify mobile vcode failed: input name not match saved name")
		return false
	}

	log.Debugf("verify user access by mobile vcode success")
	return true
}

func VerifyEmailVcode(email, vcodekey, vcode string, vcodetype int) (pass bool) {
	log.Debugf("VerifyEmailVcode")
	rds := gls.GetGlsValueNotNil("redis").(*RedisConn)
	rdskey := "USER_EMAIL_VCODE-" + GetVcodeTypeName(vcodetype) + "-" + vcodekey
	jsondata, err := rds.Do("GET", rdskey)
	if err != nil {
		log.Debugf("verify email vcode failed: GET USER_EMAIL_VCODE-" + vcodekey + " error:" + err.Error())
		return false
	} else if jsondata == nil {
		log.Debugf("verify email vcode failed: GET USER_EMAIL_VCODE-" + vcodekey + " return nil[not found]")
		return false
	}

	var vc types.VerifyCodeInfo
	err = json.Unmarshal(jsondata.([]byte), &vc)
	if err != nil {
		log.Debugf("verify email vcode failed:  unmarshal error:" + err.Error())
		return false
	} else if vc.Name != email {
		log.Debugf("verify email vcode failed: input name not match saved name")
		return false
	}

	log.Debugf("verify user access by email vcode success")
	return true
}

func VerifyUatkCookie(user User, uatk string) (pass bool) {
	log.Debugf("verify user access by uatk cookie success")
	return false
}
