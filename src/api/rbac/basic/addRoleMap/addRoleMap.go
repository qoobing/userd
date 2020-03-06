package addRoleMap

import (
	"github.com/labstack/echo"
	. "github.com/qoobing/userd/src/const"
	"github.com/qoobing/userd/src/model/t_rbac_privilege"
	"github.com/qoobing/userd/src/model/t_rbac_role"
	"github.com/qoobing/userd/src/model/t_rbac_role_map"
	. "qoobing.com/utillib.golang/api"
	"qoobing.com/utillib.golang/log"
	"strconv"
)

type Input struct {
	USER        UInfo  `json:"-"`
	UATK        string `json:"UATK"                             validate:"omitempty,min=4"`
	Description string `form:"description"  json:"description"  validate:"required,min=1"`
	RoleId      uint64 `form:"roleId"       json:"roleId"       validate:"required,min=1"`
	TargetId    uint64 `form:"targetId"     json:"targetId"     validate:"required,min=1"`
	TargetType  int    `form:"targetType"   json:"targetType"   validate:"required,min=1"`
	TargetExt   string `form:"targetExt"    json:"targetExt"    validate:"omitempty,min=1"`
}

type Output struct {
	Eno int    `json:"eno"`
	Err string `json:"err"`
}

func Main(cc echo.Context) error {
	//Step 1. init apicontext
	c := cc.(ApiContext)
	defer c.PANIC_RECOVER()
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
	//Step 2.1 parameters check for TargetType
	if err := (&t_rbac_role_map.RoleMap{}).CheckTargetType(input.TargetType); err != nil {
		return c.RESULT_PARAMETER_ERROR(err.Error())
	}

	//Step 2.2 parameters check for RoleId
	if roles, err := t_rbac_role.FindRolesByIds(c.Mysql(), []uint64{input.RoleId}); err != nil || len(roles) != 1 {
		return c.RESULT_PARAMETER_ERROR("role not exist by id " + strconv.FormatUint(input.RoleId, 10))
	}

	//Step 2.3 parameters check for TargetId
	switch input.TargetType {
	case t_rbac_role_map.TARGET_TYPE_ROLE_TO_ROLE:
		if input.RoleId == input.TargetId {
			return c.RESULT_PARAMETER_ERROR("role id must be different from target id when target type is 'TARGET_TYPE_ROLE_TO_ROLE'")
		}
		if roles, err := t_rbac_role.FindRolesByIds(c.Mysql(), []uint64{input.TargetId}); err != nil || len(roles) != 1 {
			return c.RESULT_PARAMETER_ERROR("role not exist by id " + strconv.FormatUint(input.TargetId, 10))
		}
	case t_rbac_role_map.TARGET_TYPE_ROLE_TO_PRIVILEGE:
		if privileges, err := t_rbac_privilege.FindPrivilegeByIds(c.Mysql(), []uint64{input.TargetId}); err != nil || len(privileges) != 1 {
			return c.RESULT_PARAMETER_ERROR("privilege not exist by id " + strconv.FormatUint(input.TargetId, 10))
		}
	}

	rolemap := t_rbac_role_map.RoleMap{
		F_status:      t_rbac_privilege.PRIVILEGE_STATUS_OK, //状态
		F_description: input.Description,                    //描述
		F_role_id:     input.RoleId,
		F_target_type: input.TargetType,
		F_target_id:   input.TargetId,
		F_target_ext:  input.TargetExt,
		F_creator:     input.USER.Userid, //记录创建者id
		F_modifier:    input.USER.Userid, //记录修改者id
	}

	//Step 4. create new privilege
	if err := rolemap.Save(c.Mysql()); err != nil {
		return c.RESULT_ERROR(ERR_PRIVILEGE_SAVE_FAILED, err.Error())
	}
	log.Debugf("add new rolemap success:%v", rolemap)

	//Step 5. set success
	output.Eno = 0
	output.Err = "success"

	return c.RESULT(output)
}
