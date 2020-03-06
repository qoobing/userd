package createCorporation

import (
	"github.com/labstack/echo"
	. "github.com/qoobing/userd/src/const"
	"github.com/qoobing/userd/src/model/t_corp_corporation"
	"github.com/qoobing/userd/src/model/t_corp_corporation_user"
	"github.com/qoobing/userd/src/model/t_rbac_role"
	"github.com/qoobing/userd/src/model/t_rbac_user_role"
	"github.com/qoobing/userd/src/model/t_user"
	"github.com/qoobing/userd/src/model/t_user_common_keyvalue"
	. "qoobing.com/utillib.golang/api"
	"qoobing.com/utillib.golang/log"
	"strconv"
	"time"
)

type Input struct {
	USER                UInfo  `json:"-"`
	UATK                string `json:"UATK"                    validate:"omitempty,min=4"`
	Name                string `json:"name"                    validate:"required,min=1"`         //公司名称
	Description         string `json:"description"             validate:"required,min=1"`         //公司简介
	Logo                string `json:"logo"                    validate:"omitempty,min=1"`        //公司图标
	Website             string `json:"website"                 validate:"omitempty,min=1"`        //权限参数逻辑表达式
	Addr_province       string `json:"addr_province"           validate:"required,min=1,max=64"`  //公司地址-省
	Addr_city           string `json:"addr_city"               validate:"required,min=1,max=64"`  //公司地址-市
	Addr_district       string `json:"addr_district"           validate:"omitempty,min=1,max=64"` //公司地址-区
	Addr_detail         string `json:"addr_detail"             validate:"required,min=5,max=128"` //公司地址-详细地址
	Registered_time     int64  `json:"registered_time"         validate:"required,min=0"`         //注册年份
	Registered_capital  int64  `json:"registered_capital"      validate:"required,min=0"`         //注册资本(分)
	Registered_business string `json:"registered_business"     validate:"omitempty,min=0"`        //主营业务
	Contact_name        string `json:"contact_name"            validate:"required,min=1"`         //联系人
	Contact_phone       string `json:"contact_phone"           validate:"required,min=1"`         //联系人电话
	Contact_email       string `json:"contact_email"           validate:"omitempty,min=1"`        //联系人邮箱
}

type Output struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		CorpId uint64 `json:"corpid"`
	} `json:"data"`
}

const (
	SUIK_LAST_SELECTED_CORPID = "SUIK_$#@88#$2_GYT_LAST_SELECTED_CORPID"
)

func Main(cc echo.Context) error {
	//Step 1. init apicontext
	c := cc.(ApiContext)
	defer c.PANIC_RECOVER()
	c.SetExAttribute(ATTRIBUTE_OUTPUT_FORMAT_CODE, "yes")
	c.Redis()
	c.Mysql()

	//Step 2. parameters initial
	var (
		input  Input
		output Output
	)
	if err := c.BindInput(&input); err != nil {
		return c.RESULT_PARAMETER_ERROR(err.Error())
	}
	if _, err := t_corp_corporation.FindCorporationByName(c.Mysql(), input.Name); err == nil {
		return c.RESULT_ERROR(ERR_CORPORATION_DUPLICATE, "重复的企业注册:"+input.Name)
	}

	//step 3. create new corporation
	tx := c.Mysql().Begin()
	txsuccess := false
	defer func() {
		if !txsuccess {
			tx.Rollback()
		}
	}()

	//Step 3.1 create new user
	user := t_user.User{
		F_name:         input.Name,
		F_mobile:       "",
		F_email:        "",
		F_nickname:     "",
		F_avatar:       input.Logo,
		F_sec_password: "",
	}
	if err := t_user.CreateUser(tx, &user); err != nil {
		return c.RESULT_ERROR(ERR_CORPORATION_CREATE_FAILED, err.Error())
	}

	//Step 3.2 create new corporation
	f_registered_time := time.Unix(input.Registered_time, 0).Format("2006-01-02 15:04:05")
	corp := t_corp_corporation.Corporation{
		F_id:                  user.F_user_id,               //企业法人userid，全局唯一
		F_status:              t_corp_corporation.STATUS_OK, //状态
		F_status_ext:          "",                           //状态扩展信息
		F_name:                input.Name,                   //公司名称
		F_description:         input.Description,            //公司简介
		F_logo:                input.Logo,                   //公司图标
		F_addr_province:       input.Addr_province,          //公司地址-省
		F_addr_city:           input.Addr_city,              //公司地址-市
		F_addr_district:       input.Addr_district,          //公司地址-区
		F_addr_detail:         input.Addr_detail,            //公司地址-详细地址
		F_registered_time:     f_registered_time,            //注册年份
		F_registered_capital:  input.Registered_capital,     //注册资本(分)
		F_Registered_business: input.Registered_business,    //主营业务
		F_website:             input.Website,                //权限参数逻辑表达式
		F_contact_name:        input.Contact_name,           //联系人
		F_contact_phone:       input.Contact_phone,          //联系人电话
		F_contact_email:       input.Contact_email,          //联系人邮箱
		F_administrator:       input.USER.Userid,            //管理员
		F_creator:             input.USER.Userid,            //记录创建者id
		F_modifier:            input.USER.Userid,            //记录修改者id
	}
	if err := corp.Save(tx); err != nil {
		return c.RESULT_ERROR(ERR_CORPORATION_CREATE_FAILED, err.Error())
	}

	//Step 3.3 create corporation inner role
	role := t_rbac_role.Role{
		F_status:      t_rbac_role.ROLE_STATUS_OK,              //状态
		F_name:        "零权限",                                   //角色名称
		F_description: "内置角色",                                  //角色描述
		F_scene_key:   t_rbac_role.SceneKeyGYCORPID(corp.F_id), //角色适用场景
		F_creator:     input.USER.Userid,                       //记录创建者id
		F_modifier:    input.USER.Userid,                       //记录修改者id
	}
	if err := role.CreateRole(tx); err != nil {
		return c.RESULT_ERROR(ERR_CORPORATION_CREATE_FAILED, err.Error())
	}
	//Step 3.4 create new corporation admin user
	corpuser := t_corp_corporation_user.CorporationUser{
		F_status:         t_corp_corporation_user.STATUS_OK, //状态
		F_status_ext:     "",                                //状态扩展信息
		F_corporation_id: corp.F_id,                         //企业法人userid
		F_user_id:        input.USER.Userid,                 //企业员工userid
		F_description:    input.Description,                 //备注信息
	}
	if err := corpuser.Save(tx); err != nil {
		return c.RESULT_ERROR(ERR_CORPORATION_ADD_USER_FAILED, err.Error())
	}

	//Step 3.5 create new corporation admin user roles
	scenekey := t_rbac_role.SceneKeyGYCORPID(corp.F_id)
	// |--------------------------------
	// | roleid |   描述
	// | 983261 |前台企业中心模块>公共模块
	// | 983268 |前台企业中心模块>买家中心
	// | 983273 |前台企业中心模块>卖家中心
	// |--------------------------------
	adminroleids := []uint64{983261}
	if err := t_rbac_user_role.CreateUserRoles(tx, scenekey, input.USER.Userid, adminroleids); err != nil {
		return c.RESULT_ERROR(ERR_CORPORATION_CREATE_FAILED, err.Error())
	}

	///////// 临时逻辑  for 广盈通项目记录用户最后所选企业 ////////begin/////////////
	//TODO: move to frontend
	//Step 3.6 user extension info for SUIK_$#@88#$2_LAST_SELECT_KEY
	t_user_common_keyvalue.Set(tx, input.USER.Userid, SUIK_LAST_SELECTED_CORPID, strconv.FormatUint(corp.F_id, 10))
	///////// 临时逻辑  for 广盈通项目记录用户最后所选企业 ////////end///////////////

	//Step 4 commit transaction
	if rtx := tx.Commit(); rtx.Error != nil {
		log.Debugf("create new role failed:%s", rtx.Error.Error())
		return c.RESULT_ERROR(ERR_INNER_ERROR, "inner error, transaction commit failed")
	} else {
		txsuccess = true
		log.Debugf("create new role success:%v", role)
	}

	//Step 5. set success
	output.Code = 0
	output.Msg = "success"
	output.Data.CorpId = corp.F_id

	return c.RESULT(output)
}
