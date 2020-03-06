package generateTemplateRoleTemplate

import (
	"github.com/labstack/echo"
	"github.com/qoobing/userd/src/api/rbac/basic/checkPermission"
	"github.com/qoobing/userd/src/model/t_rbac_template"
	. "qoobing.com/utillib.golang/api"
	"qoobing.com/utillib.golang/log"
)

type Input struct {
	Description string `form:"description" json:"description"  validate:"required,min=1"`
}

type Output struct {
	Code             int                              `json:"code"`
	Msg              string                           `json:"msg"`
	TemplateRoleTree t_rbac_template.TemplateRoleTree `json:"templateRoleTree"`
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
	rootids := []uint64{983219, 983222, 983226, 983230, 983233, 983239, 983241, 983252, 983257}
	roleTree := checkPermission.GetRoleTreeByRoleIds(c.Mysql(), rootids)
	log.PrintPreety("roleTree:", roleTree)

	rootnodeids := []checkPermission.NodeId{}
	for _, rid := range rootids {
		rootnodeids = append(rootnodeids, checkPermission.NodeId{rid, 1})
	}
	output.TemplateRoleTree = roleTreeToTemplateRoleTree(roleTree, rootnodeids)

	//Step 5. set success
	output.Code = 0
	output.Msg = "success"
	return c.RESULT(output)
}

func roleTreeToTemplateRoleTree(roletree checkPermission.RoleTree, rootids []checkPermission.NodeId) (templateRoleTree t_rbac_template.TemplateRoleTree) {
	templateRoleTree = t_rbac_template.TemplateRoleTree{
		Id:       0,
		Type:     1,
		Name:     "root",
		Selected: false,
		Children: roleTreeToTemplateRoleForest(roletree, rootids),
	}
	return
}

func roleTreeToTemplateRoleForest(roletree checkPermission.RoleTree, rootids []checkPermission.NodeId) (templateRoleTree []t_rbac_template.TemplateRoleTree) {
	log.PrintPreety("roleTreeToTemplateRoleForest, input ids:", rootids)
	for _, rootid := range rootids {
		log.PrintPreety("roleTreeToTemplateRoleForest, add children for:", rootid.String())
		if role, ok := roletree[rootid]; ok {
			root := t_rbac_template.TemplateRoleTree{
				Id:       role.Id.Id,
				Type:     role.Id.Type,
				Name:     role.Name,
				Selected: false,
				Children: []t_rbac_template.TemplateRoleTree{},
			}
			crootids := []checkPermission.NodeId{}
			for _, child := range role.Children {
				crootids = append(crootids, checkPermission.NodeId{child.Id.Id, child.Id.Type})
			}
			log.Debugf("roleTreeToTemplateRoleForest, add children for:%s", rootid.String())
			root.Children = roleTreeToTemplateRoleForest(roletree, crootids)
			log.PrintPreety("roleTreeToTemplateRoleForest, add children is:", root.Children)

			templateRoleTree = append(templateRoleTree, root)
		} else {
			log.Debugf("id(%d,%d) not found in roletree", rootid.Id, rootid.Type)
		}
	}

	return
}
