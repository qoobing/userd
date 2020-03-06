package createPrivilege

import (
	"github.com/labstack/echo"
	. "github.com/qoobing/userd/src/const"
	"github.com/qoobing/userd/src/model/t_rbac_privilege"
	. "qoobing.com/utillib.golang/api"
	"qoobing.com/utillib.golang/log"
)

type Input struct {
	USER        UInfo  `json:"-"`
	UATK        string `json:"UATK"         validate:"omitempty,min=4"`
	Name        string `form:"name"         json:"name"         validate:"required,min=1"`
	Description string `form:"description"  json:"description"  validate:"required,min=1"`
	Uri         string `form:"uri"          json:"uri"          validate:"required,min=1"`
	Expression  string `form:"expression"   json:"expression"   validate:"omitempty,min=8"`
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
	if _, err := t_rbac_privilege.FindPrivilegeByUniqueKey(c.Mysql(), input.Name, input.Uri, input.Expression); err == nil {
		return c.RESULT_ERROR(ERR_PRIVILEGE_DUPLICATE,
			"重复的权限:name="+input.Name+",uri="+input.Uri+",expression="+input.Expression)
	}

	privilege := t_rbac_privilege.Privilege{
		F_status:      t_rbac_privilege.PRIVILEGE_STATUS_OK, //状态
		F_name:        input.Name,                           //权限名称
		F_description: input.Description,                    //权限描述
		F_uri:         input.Uri,                            //uri正则表达式
		F_expression:  input.Expression,                     //权限参数逻辑表达式
		F_creator:     input.USER.Userid,                    //记录创建者id
		F_modifier:    input.USER.Userid,                    //记录修改者id
	}

	//Step 4. create new privilege
	privilege.Save(c.Mysql())
	if err := privilege.Save(c.Mysql()); err != nil {
		return c.RESULT_ERROR(ERR_PRIVILEGE_SAVE_FAILED, err.Error())
	}
	log.Debugf("create new privilege:%v", privilege)

	//Step 5. set success
	output.Code = 0
	output.Msg = "success"

	return c.RESULT(output)
}
