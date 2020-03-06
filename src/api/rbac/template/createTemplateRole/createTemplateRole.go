package createTemplateRole

import (
	"github.com/labstack/echo"
	"github.com/qoobing/userd/src/api/rbac/template/getTemplateRoleDetail"
	"github.com/qoobing/userd/src/common"
	. "github.com/qoobing/userd/src/const"
	"github.com/qoobing/userd/src/model/t_rbac_role"
	"github.com/qoobing/userd/src/model/t_rbac_role_map"
	"github.com/qoobing/userd/src/model/t_rbac_template"
	. "qoobing.com/utillib.golang/api"
	"qoobing.com/utillib.golang/log"
)

type Input struct {
	USER             UInfo                            `json:"-"`
	UATK             string                           `json:"UATK"         validate:"omitempty,min=4"`
	Name             string                           `json:"name"         validate:"required,min=1"`
	Description      string                           `json:"description"  validate:"required,min=1"`
	SceneKey         string                           `json:"scenekey"     validate:"required,min=1"`
	TemplateId       uint64                           `json:"templateid"   validate:"required,min=1"`
	TemplateRoleTree t_rbac_template.TemplateRoleTree `json:"templatetree" validate:"required,min=3"`
}

type Output struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		RoleId uint64 `json:"roleid"`
	} `json:"data"`
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
		input  Input
		output Output
	)
	if err := c.BindInput(&input); err != nil {
		return c.RESULT_PARAMETER_ERROR(err.Error())
	}
	if err := t_rbac_role.CheckSceneKeyFormat(input.SceneKey); err != nil {
		return c.RESULT_ERROR(ERR_PARAMETER_INVALID, err.Error())
	}

	//Step 3. create new role
	txsuccess := false
	tx := c.Mysql().Begin()
	defer func() {
		if !txsuccess {
			tx.Rollback()
		}
	}()

	role := t_rbac_role.Role{
		F_status:      t_rbac_role.ROLE_STATUS_OK, //状态
		F_name:        input.Name,                 //角色名称
		F_description: input.Description,          //角色描述
		F_scene_key:   input.SceneKey,             //角色适用场景
		F_creator:     input.USER.Userid,          //记录创建者id
		F_modifier:    input.USER.Userid,          //记录修改者id
	}
	var retrytimes = 0
	for {
		role.F_id = common.GenerateId("ROLE_ID")
		if roles, err := t_rbac_role.FindRolesByIds(c.Mysql(), []uint64{role.F_id}); err == nil && len(roles) == 0 {
			break
		} else if retrytimes > 3 {
			return c.RESULT_ERROR(ERR_INNER_ERROR, "Generate roleId failed")
		}
		retrytimes++
	}

	//Step 4. create all rolemap
	log.PrintPreety("input.TemplateRoleTree", input.TemplateRoleTree)
	ids := getTemplateRoleDetail.GetAllSelectedIdsFromTemplateRoleTree(&input.TemplateRoleTree)
	rolemaps := []*t_rbac_role_map.RoleMap{}
	for _, target_id := range ids {
		newrolemap := &t_rbac_role_map.RoleMap{
			F_status:      t_rbac_role_map.ROLE_STATUS_OK,
			F_status_ext:  "",
			F_role_id:     role.F_id,
			F_scene_key:   input.SceneKey,
			F_target_type: target_id.Type,
			F_target_id:   target_id.Id,
			F_target_ext:  "",
			F_creator:     input.USER.Userid,
			F_modifier:    input.USER.Userid,
		}
		rolemaps = append(rolemaps, newrolemap)
	}

	if err := role.Save(tx); err != nil {
		return c.RESULT_ERROR(ERR_ROLE_SAVE_FAILED, err.Error())
	}
	if err := t_rbac_role_map.CreateRoleMaps(tx, rolemaps); err != nil {
		return c.RESULT_ERROR(ERR_ROLE_MAP_SAVE_FAILED, err.Error())
	}

	//Step 5 commit create new role
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
	output.Data.RoleId = role.F_id

	return c.RESULT(output)
}
