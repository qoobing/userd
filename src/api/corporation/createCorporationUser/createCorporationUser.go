package createCorporationUser

import (
	"github.com/labstack/echo"
	"github.com/qoobing/userd/src/api/corporation/createCorporation"
	"github.com/qoobing/userd/src/api/corporation/getCorporationUserDetail"
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
	"yqtc.com/ubox.golib/xyz"
)

type inputRoleItem getCorporationUserDetail.OutputRoleItem

type Input struct {
	USER           UInfo           `json:"-"`
	UATK           string          `json:"UATK"           validate:"omitempty,min=4"`
	CorpId         uint64          `json:"corpid"         validate:"required,min=1"`  //企业ID
	Status         int             `json:"status"         validate:"oneof=1 2 10"`    //初始状态（1:正常；2:删除；10:禁用）
	Phone          string          `json:"phone"          validate:"required,min=11"` //手机号
	CreatePassword string          `json:"createPassword" validate:"omitempty,min=8"` //当用户不存在是，用此密码强制创建
	Roles          []inputRoleItem `json:"roles"          validate:"omitempty"`       //角色列表
	Description    string          `json:"description"    validate:"omitempty"`       //备注信息
}

type Output struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func Main(cc echo.Context) error {
	//Step 1. init apicontext
	c := cc.(ApiContext)
	defer c.PANIC_RECOVER()
	c.SetExAttribute(ATTRIBUTE_OUTPUT_FORMAT_CODE, "yes")
	c.Redis()
	c.Mysql()

	//Step 2. parameters initial
	var (
		input        Input
		output       Output
		inputRoleids = []uint64{}
		newuser      t_user.User
		err          error
	)
	if err := c.BindInput(&input); err != nil {
		return c.RESULT_PARAMETER_ERROR(err.Error())
	}
	if _, err = t_corp_corporation.FindCorporationById(c.Mysql(), input.CorpId); err != nil {
		return c.RESULT_ERROR(ERR_CORPORATION_NOT_EXIST, "Corporation Not Exist")
	}
	newuser, err = t_user.FindUser(c.Mysql(), input.Phone, NAMETYPE_MOBILE)
	if err != nil && err.Error() == USER_NOT_EXIST && input.CreatePassword == "" {
		return c.RESULT_ERROR(ERR_USER_NOT_EXIST, "User Not Exist")
	} else if err == nil && len(input.CreatePassword) != 0 {
		return c.RESULT_ERROR(ERR_CREATE_USER_ERROR, "User Have Exist, 'CreatePassword' should be empty")
	} else if err != nil && err.Error() != USER_NOT_EXIST {
		return c.RESULT_ERROR(ERR_CORPORATION_ADD_USER_FAILED, "Unkown error:"+err.Error())
	}
	if len(input.CreatePassword) == 0 {
		if _, err := t_corp_corporation_user.FindCorporationUserById(c.Mysql(), input.CorpId, newuser.F_user_id); err == nil {
			return c.RESULT_ERROR(ERR_CORPORATION_USER_DUPLICATE, "企业中已有该用户")
		}
	}
	for _, r := range input.Roles {
		inputRoleids = append(inputRoleids, r.RoleId)
	}
	scenekey := t_rbac_role.SceneKeyGYCORPID(input.CorpId)
	if roles, err := t_rbac_role.FindRolesBySceneKeyAndIds(c.Mysql(), scenekey, inputRoleids); err != nil {
		return c.RESULT_ERROR(ERR_CORPORATION_ADD_USER_FAILED, "添加用户失败")
	} else if len(inputRoleids) != len(roles) {
		return c.RESULT_ERROR(ERR_PARAMETER_INVALID, "角色错误")
	}

	//step 3. create new corporation
	tx := c.Mysql().Begin()
	txsuccess := false
	defer func() {
		if !txsuccess {
			tx.Rollback()
		}
	}()

	//Step 3.1 create new corporation user
	if len(input.CreatePassword) > 0 {
		newuser = t_user.User{
			F_name:         input.Phone,
			F_mobile:       input.Phone,
			F_email:        "",
			F_nickname:     "",
			F_avatar:       "",
			F_sec_password: input.CreatePassword,
		}
		if err := t_user.CreateUser(tx, &newuser); err != nil {
			return c.RESULT_ERROR(ERR_CORPORATION_CREATE_FAILED, err.Error())
		}
	}
	status := xyz.IF(input.Status == 0, 1, input.Status).(int)
	corpuser := t_corp_corporation_user.CorporationUser{
		F_status:         status,            //状态
		F_status_ext:     "",                //状态扩展信息
		F_corporation_id: input.CorpId,      //企业法人userid
		F_user_id:        newuser.F_user_id, //企业员工userid
		F_description:    input.Description, //备注信息
	}
	if err := corpuser.Save(tx); err != nil {
		return c.RESULT_ERROR(ERR_CORPORATION_ADD_USER_FAILED, err.Error())
	}

	//Step 3.2 create new corporation user roles
	if err := t_rbac_user_role.CreateUserRoles(tx, scenekey, newuser.F_user_id, inputRoleids); err != nil {
		return c.RESULT_ERROR(ERR_CORPORATION_CREATE_FAILED, err.Error())
	}

	///////// 临时逻辑  for 广盈通项目记录用户最后所选企业 ////////begin/////////////
	//TODO: move to frontend
	//Step 3.3 user extension info for SUIK_LAST_SELECTED_CORPID
	if input.CorpId > 1000000 {
		t_user_common_keyvalue.Set(tx, input.USER.Userid,
			createCorporation.SUIK_LAST_SELECTED_CORPID,
			strconv.FormatUint(input.CorpId, 10))
	}
	///////// 临时逻辑  for 广盈通项目记录用户最后所选企业 ////////end///////////////

	//Step 4 commit transaction
	if rtx := tx.Commit(); rtx.Error != nil {
		log.Debugf("create new corporation user failed:%s", rtx.Error.Error())
		return c.RESULT_ERROR(ERR_INNER_ERROR, "inner error, transaction commit failed")
	} else {
		txsuccess = true
		log.Debugf("create new corporation user success:%v", corpuser)
	}

	//Step 5. set success
	output.Code = 0
	output.Msg = "success"

	return c.RESULT(output)
}
