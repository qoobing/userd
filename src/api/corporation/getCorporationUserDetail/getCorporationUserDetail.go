package getCorporationUserDetail

import (
	"github.com/labstack/echo"
	. "github.com/qoobing/userd/src/const"
	"github.com/qoobing/userd/src/model/t_corp_corporation"
	"github.com/qoobing/userd/src/model/t_corp_corporation_user"
	"github.com/qoobing/userd/src/model/t_rbac_role"
	"github.com/qoobing/userd/src/model/t_rbac_user_role"
	"github.com/qoobing/userd/src/model/t_user"
	. "qoobing.com/utillib.golang/api"
	"qoobing.com/utillib.golang/log"
	"time"
)

const PHONE_FOR_GET_CORPORATION_DETAIL_DEFUALT = "10000000000"

type Input struct {
	USER   UInfo  `json:"-"`
	UATK   string `json:"UATK"              validate:"omitempty,min=4"`
	CorpId uint64 `json:"corpid"            validate:"required,min=1"`  //企业ID
	Phone  string `json:"phone"             validate:"required,len=11"` //手机号
}

type Output struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		No          string            `json:"NO"`                                     //序号
		UserId      uint64            `json:"userid"`                                 //用户ID
		Phone       string            `json:"phone"`                                  //手机号
		CreateTime  int64             `json:"createtime"`                             //创建时间
		Roles       []*OutputRoleItem `json:"roles"           validate:"omitempty"`   //角色id列表
		Description string            `json:"description"       validate:"omitempty"` //备注信息
	} `json:"data"`
}

type OutputRoleItem struct {
	RoleId   uint64 `json:"roleid"`   //角色id
	RoleName string `json:"rolename"` //角色名称
	Selected bool   `json:"selected"` //是否具有此角色
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
		input    Input
		output   Output
		corpuser t_corp_corporation_user.CorporationUser
		user     t_user.User
		err      error
	)
	if err := c.BindInput(&input); err != nil {
		return c.RESULT_PARAMETER_ERROR(err.Error())
	}
	if _, err = t_corp_corporation.FindCorporationById(c.Mysql(), input.CorpId); err != nil {
		return c.RESULT_ERROR(ERR_CORPORATION_NOT_EXIST, "Corporation Not Exist")
	}
	if input.Phone == PHONE_FOR_GET_CORPORATION_DETAIL_DEFUALT {
		log.Debugf("phone is PHONE_FOR_GET_CORPORATION_DETAIL_DEFUALT")
	} else if user, err = t_user.FindUser(c.Mysql(), input.Phone, NAMETYPE_MOBILE); err != nil && err.Error() == USER_NOT_EXIST {
		return c.RESULT_ERROR(ERR_USER_NOT_EXIST, "User Not Exist")
	}

	if input.Phone == PHONE_FOR_GET_CORPORATION_DETAIL_DEFUALT {
		//log
	} else if corpuser, err = t_corp_corporation_user.FindCorporationUserById(c.Mysql(), input.CorpId, user.F_user_id); err != nil {
		return c.RESULT_ERROR(ERR_USER_NOT_EXIST, "企业中不存在该用户")
	}

	//step 3. find all roles
	scenekey := t_rbac_role.SceneKeyGYCORPID(input.CorpId)
	if roles, err := t_rbac_role.FindRolesBySceneKey(c.Mysql(), scenekey); err != nil {
		return c.RESULT_ERROR(ERR_INNER_ERROR, "用户详情失败")
	} else {
		for _, r := range roles {
			output.Data.Roles = append(output.Data.Roles, &OutputRoleItem{
				RoleId:   r.F_id,
				RoleName: r.F_name,
				Selected: false,
			})
		}
	}

	//step 3.1 find user roles
	log.PrintPreety("output.Data.Roles", output.Data.Roles)
	userroles, err := t_rbac_user_role.FindUserRoles(c.Mysql(), user.F_user_id)
	if err == nil {
		for _, ur := range userroles {
			for _, out := range output.Data.Roles {
				if out.RoleId == ur.F_role_id {
					out.Selected = true
				}
			}
		}
	}
	log.PrintPreety("userroles", userroles)
	log.PrintPreety("output.Data.Roles", output.Data.Roles)

	//Step 5. set success
	creattime, _ := time.Parse(time.RFC3339, corpuser.F_create_time)
	output.Code = 0
	output.Msg = "success"
	output.Data.Description = corpuser.F_description
	output.Data.Phone = user.F_mobile
	output.Data.UserId = user.F_user_id
	output.Data.CreateTime = creattime.Unix()
	output.Data.No = "000" //TODO

	return c.RESULT(output)
}
