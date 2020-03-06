package createRole

import (
	"github.com/labstack/echo"
	. "github.com/qoobing/userd/src/const"
	"github.com/qoobing/userd/src/model/t_rbac_role"
	. "qoobing.com/utillib.golang/api"
	"qoobing.com/utillib.golang/log"
)

type Input struct {
	USER        UInfo  `json:"UATK"`
	UATK        string `json:"UATK"         validate:"omitempty,min=4"`
	Name        string `form:"name"         json:"name"         validate:"required,min=1"`
	Description string `form:"description"  json:"description"  validate:"required,min=1"`
}

type Output struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func Main(cc echo.Context) error {
	//Step 1. init apicontext
	c := cc.(ApiContext)
	c.SetExAttribute(ATTRIBUTE_OUTPUT_FORMAT_CODE, "yes")
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

	//Step 3. check duplicate
	scenekey := "INNER"
	if _, err := t_rbac_role.FindRoleByUniqueKey(c.Mysql(), "INNER", input.Name); err == nil {
		return c.RESULT_ERROR(ERR_PRIVILEGE_DUPLICATE,
			"重复的角色:scenekey="+scenekey+",name="+input.Name)
	}

	role := t_rbac_role.Role{
		F_status:      t_rbac_role.ROLE_STATUS_OK, //状态
		F_name:        input.Name,                 //角色名称
		F_description: input.Description,          //角色描述
		F_scene_key:   scenekey,                   //角色场景
		F_creator:     input.USER.Userid,          //记录创建者id
		F_modifier:    input.USER.Userid,          //记录修改者id
	}

	//Step 4. create new privilege
	if err := role.Save(c.Mysql()); err != nil {
		return c.RESULT_ERROR(ERR_PRIVILEGE_SAVE_FAILED, err.Error())
	}
	log.Debugf("create new role:%v", role)

	//Step 5. set success
	output.Code = 0
	output.Msg = "success"

	return c.RESULT(output)
}
