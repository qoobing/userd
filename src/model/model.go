package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/jinzhu/gorm"
	"github.com/qoobing/userd/src/config"
	. "github.com/qoobing/userd/src/const"
	"github.com/qoobing/userd/src/model/t_corp_corporation"
	"github.com/qoobing/userd/src/model/t_corp_corporation_user"
	"github.com/qoobing/userd/src/model/t_login_scene"
	"github.com/qoobing/userd/src/model/t_rbac_privilege"
	"github.com/qoobing/userd/src/model/t_rbac_role"
	"github.com/qoobing/userd/src/model/t_rbac_role_map"
	"github.com/qoobing/userd/src/model/t_rbac_template"
	"github.com/qoobing/userd/src/model/t_rbac_user_role"
	"github.com/qoobing/userd/src/model/t_user"
	"github.com/qoobing/userd/src/model/t_user_common_keyvalue"
	"github.com/qoobing/userd/src/types"
	. "qoobing.com/utillib.golang/api"
	"qoobing.com/utillib.golang/log"
	"qoobing.com/utillib.golang/xyz"
	"strconv"
)

func InitDatabase() {
	db, err := gorm.Open("mysql", config.Config().Database)
	defer db.Close()
	if err != nil {
		log.Fatalf("connect mysql[%s] failed [%s]", config.Config().Database, err)
		panic("connect mysql failed")
	}
	tables := []string{
		t_user.TABLESQL,
		t_user_common_keyvalue.TABLESQL,
		t_login_scene.TABLESQL,
		t_rbac_privilege.TABLESQL,
		t_rbac_role.TABLESQL,
		t_rbac_role_map.TABLESQL,
		t_rbac_template.TABLESQL,
		t_rbac_user_role.TABLESQL,
		t_corp_corporation.TABLESQL,
		t_corp_corporation_user.TABLESQL,
	}
	for _, table := range tables {
		rdb := db.Exec(table)
		if rdb.Error != nil {
			log.Fatalf("create table failed [%s], sql:[%s]", rdb.Error.Error(), table)
		}
	}
	log.Debugf("init database done")
}

func SetUserAccessTokenData(rds redis.Conn, user t_user.User, scene t_login_scene.LoginScene) (name string, err error) {
	uid := "0000" + strconv.FormatUint(user.F_user_id, 10)
	name = "V1" + "S" + fmt.Sprintf("%04d", scene.F_scene_id)
	name = name + "R" + xyz.GetRandomString(20)
	name = name + "U" + uid[len(uid)-3:]
	err = UpdateUserAccessTokenData(rds, user, name)
	return name, err
}

func UpdateUserAccessTokenData(rds redis.Conn, user t_user.User, name string) (err error) {
	type UserAccessToken struct {
		name  string
		value string
		uinfo types.Userinfo
	}
	uatk := UserAccessToken{
		name:  name,
		value: "",
		uinfo: types.Userinfo{
			Name:           user.F_name,
			Nickname:       user.F_nickname,
			Avatar:         user.F_avatar,
			Lastupdatetime: user.F_name,
			Userid:         user.F_user_id,
			Loginstate:     0,
		},
	}
	data, err := json.Marshal(uatk.uinfo)
	if err != nil {
		return err
	}
	uatk.value = string(data[:])

	_, err = rds.Do("SETEX", "USER_ACCESS_TOKEN-"+uatk.name, config.Config().UatkTimeout, uatk.value)
	if err != nil {
		return err
	}
	log.Debugf("SETEX USER_ACCESS_TOKEN-%s success, value:[%s]", uatk.name, uatk.value)
	useridstr := fmt.Sprintf("{%d}", user.F_user_id)
	_, err = rds.Do("HSET", "USER_INFO-"+useridstr, "USER_ACCESS_TOKEN", "USER_ACCESS_TOKEN-"+uatk.name)
	return nil
}

func DelUserAccessTokenData(rds redis.Conn, user t_user.User) (err error) {
	useridstr := fmt.Sprintf("{%d}", user.F_user_id)
	uatkname, err := rds.Do("HGET", "USER_INFO-"+useridstr, "USER_ACCESS_TOKEN")
	if err != nil {
		return err
	}
	if uatkname == nil || uatkname == "" {
		return nil
	}

	_, err = rds.Do("DEL", uatkname)
	if err != nil {
		return err
	}
	log.Debugf("DEL %s success", uatkname)
	return nil
}

func GetUserAccessTokenData(rds redis.Conn, uatkname string) (uinfo types.Userinfo, err error) {
	strfn := log.TRACE_INTO("get uatk data")
	defer log.TRACE_EXIT(strfn, "get uatk data")

	jsondata, err := rds.Do("GET", "USER_ACCESS_TOKEN-"+uatkname)
	if err != nil {
		return uinfo, errors.New("GET USER_ACCESS_TOKEN-" + uatkname + " error:" + err.Error())
	} else if jsondata == nil {
		return uinfo, errors.New("GET USER_ACCESS_TOKEN-" + uatkname + " return nil[not found]")
	}

	err = json.Unmarshal(jsondata.([]byte), &uinfo)
	if err != nil {
		return uinfo, errors.New("unmarshal error:" + err.Error())
	}
	log.Debugf("Get USER_ACCESS_TOKEN success: %v", uinfo)
	return uinfo, nil
}

func GetUserAccessTokenTTL(rds redis.Conn, uatkname string) (ttl int64, err error) {
	strfn := log.TRACE_INTO("get uatk ttl")
	defer log.TRACE_EXIT(strfn, "get uatk ttl")

	ttldata, err := rds.Do("TTL", "USER_ACCESS_TOKEN-"+uatkname)
	if err != nil {
		return 0, errors.New("TTL USER_ACCESS_TOKEN-" + uatkname + " error:" + err.Error())
	} else if ttldata == nil {
	}
	log.Debugf("ttldata=%v", ttldata, ttldata)
	ttl = ttldata.(int64)
	if ttl == -1 {
		return 0, errors.New("TTL USER_ACCESS_TOKEN-" + uatkname + " return -1[no expire time]")
	} else if ttl == -2 {
		return 0, errors.New("TTL USER_ACCESS_TOKEN-" + uatkname + " return -2[key not exist]")
	}

	log.Debugf("TTL USER_ACCESS_TOKEN success: %v", ttl)
	return ttl, nil
}

func GetLoginInfo(c ApiContext, rds redis.Conn) (uinfo types.Userinfo, err error) {
	//Step 1. get redis key from input
	uatkname := c.QueryParam("UATK")
	if uatkname == "" {
		uatkname = c.FormValue("UATK")
	}
	if uatkname == "" {
		if cookie, err := c.Cookie(COOKIE_NAME_USERINFO); err == nil {
			uatkname = cookie.Value
		} else {
			return uinfo, errors.New("have no 'UATK' parameter and 'UATK' cookie")
		}
	}

	//Step 2. get userinfo from redis
	uinfo, err = GetUserAccessTokenData(rds, uatkname)
	return uinfo, err
}

func EstimateNametype(name string) (nametype int) {
	if config.Config().MobileCompiledReg.MatchString(name) {
		return NAMETYPE_MOBILE
	} else if config.Config().EmailCompiledReg.MatchString(name) {
		return NAMETYPE_EMAIL
	} else {
		return NAMETYPE_NAME
	}
}

//func GetDiffField(dbdata interface{}, newdata interface{}) (diff map[string]interface{}) {
//	a := reflect.ValueOf(dbdata)
//	if a.Kind() == reflect.Ptr {
//		dbdata = reflect.Indirect(a)
//	}
//	if a.Kind() != reflect.Struct {
//		panic("ERROR: Type is not struct")
//	}
//
//	b := reflect.ValueOf(newdata)
//	if b.Kind() == reflect.Ptr {
//		dbdata = reflect.Indirect(a)
//	}
//	if b.Kind() != reflect.Struct {
//		panic("ERROR: Type is not struct")
//	}
//
//	x := reflect.New(a.Type()).Elem()
//
//	for i := 0; i < b.NumField(); i++ {
//
//		field := a.Type().Field(i)
//		tname := field.Name
//
//		af := a.Field(i)
//		bf := x.FieldByName(field.Name)
//
//		af.
//
//		fpath := copyAppend(path, tname)
//
//		err := nd.diff(fpath, xf, af)
//		if err != nil {
//			return err
//		}
//	}
//
//	for i := 0; i < len(nd.cl); i++ {
//		(d.cl) = append(d.cl, swapChange(t, nd.cl[i]))
//	}
//}
