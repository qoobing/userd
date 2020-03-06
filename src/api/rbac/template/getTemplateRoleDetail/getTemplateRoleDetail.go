package getTemplateRoleDetail

import (
	"github.com/labstack/echo"
	"github.com/qoobing/userd/src/api/rbac/basic/checkPermission"
	. "github.com/qoobing/userd/src/const"
	"github.com/qoobing/userd/src/model/t_rbac_role"
	"github.com/qoobing/userd/src/model/t_rbac_role_map"
	"github.com/qoobing/userd/src/model/t_rbac_template"
	. "qoobing.com/utillib.golang/api"
	"qoobing.com/utillib.golang/log"
	"qoobing.com/utillib.golang/xyz"
)

const ROLEID_FOR_GET_TEMPLATE_DEFUALT = 0

type Input struct {
	USER       UInfo  `json:"-"`
	UATK       string `json:"UATK"         validate:"omitempty,min=4"`
	TemplateId uint64 `json:"templateId"   validate:"required,min=1"`
	RoleId     uint64 `json:"roleId"       validate:"required,min=1"`
}

type Output struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		RoleId           uint64                           `json:"roleId"`
		TemplateId       uint64                           `json:"templateId"`
		RoleName         string                           `json:"roleName"`
		RoleDescription  string                           `json:"roleDescription"`
		TemplateRoleTree t_rbac_template.TemplateRoleTree `json:"templateRoleTree"`
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

	output.Data.RoleId = input.RoleId
	output.Data.TemplateId = input.TemplateId
	if input.RoleId == ROLEID_FOR_GET_TEMPLATE_DEFUALT {
		log.Debugf("intent to get template detail, not template role tree")
	} else if roles, err := t_rbac_role.FindRolesByIds(c.Mysql(), []uint64{input.RoleId}); err != nil || len(roles) != 1 {
		return c.RESULT_ERROR(ERR_ROLE_NOT_EXIST, "role not exist")
	} else {
		output.Data.RoleName = roles[0].F_name
		output.Data.RoleDescription = roles[0].F_description
	}

	//Step 3. get template tree
	templateRoleTree, err := t_rbac_template.ReadTemplateRoleTreeFromDb(c.Mysql(), input.TemplateId)
	if err != nil {
		return c.RESULT_ERROR(ERR_TEMPLATE_ROLE_NOT_EXIST, "template not exist")
	}

	//Step 4. get role tree by roleid
	rolemaplist, err := t_rbac_role_map.FindChildrenMap(c.Mysql(), []uint64{input.RoleId})
	if err != nil {
		return c.RESULT_ERROR(ERR_INNER_ERROR, "FindChildrenMap Error:"+err.Error())
	}
	log.Debugf("FindChildrenMap %s", rolemaplist)

	//Step 5. merge jsontree
	MergeRoleTreeToTemplateRoleTree(&templateRoleTree, rolemaplist)

	//Step 6. set success
	output.Code = 0
	output.Msg = "success"
	output.Data.TemplateRoleTree = templateRoleTree

	return c.RESULT(output)
}

func MergeRoleTreeToTemplateRoleTree(templateRoleTree *t_rbac_template.TemplateRoleTree, rolemaplist []t_rbac_role_map.RoleMap) {
	selected := map[checkPermission.NodeId]*t_rbac_role_map.RoleMap{}
	for _, rolemap := range rolemaplist {
		id := checkPermission.NodeId{Id: rolemap.F_target_id, Type: rolemap.F_target_type}
		selected[id] = &rolemap
	}
	s := xyz.NewStack()
	s.Push(templateRoleTree)

	for !s.Empty() {
		cur := s.Pop().(*t_rbac_template.TemplateRoleTree)
		id := checkPermission.NodeId{Id: cur.Id, Type: cur.Type}
		if _, ok := selected[id]; ok {
			cur.Selected = true
		}

		for id, _ := range cur.Children {
			s.Push(&cur.Children[id])
		}
	}
}

func GetAllSelectedIdsFromTemplateRoleTree(templateRoleTree *t_rbac_template.TemplateRoleTree) (ids []checkPermission.NodeId) {
	s := xyz.NewStack()
	s.Push(templateRoleTree)

	for !s.Empty() {
		cur := s.Pop().(*t_rbac_template.TemplateRoleTree)
		id := checkPermission.NodeId{Id: cur.Id, Type: cur.Type}
		if cur.Selected == true {
			ids = append(ids, id)
		}

		for id, _ := range cur.Children {
			s.Push(&cur.Children[id])
		}
	}
	return
}
