package updateTemplateRole

import (
	"github.com/labstack/echo"
	"github.com/qoobing/userd/src/api/rbac/basic/checkPermission"
	"github.com/qoobing/userd/src/api/rbac/template/getTemplateRoleDetail"
	. "github.com/qoobing/userd/src/const"
	"github.com/qoobing/userd/src/model/t_rbac_role"
	"github.com/qoobing/userd/src/model/t_rbac_role_map"
	"github.com/qoobing/userd/src/model/t_rbac_template"
	. "qoobing.com/utillib.golang/api"
	"qoobing.com/utillib.golang/log"
	"time"
)

type Input struct {
	USER             UInfo                            `json:"-"`
	UATK             string                           `json:"UATK"         validate:"omitempty,min=4"`
	Name             string                           `json:"name"         validate:"required,min=1"`
	Description      string                           `json:"description"  validate:"required,min=1"`
	RoleId           uint64                           `json:"roleid"       validate:"required,min=1"`
	TemplateId       uint64                           `json:"templateid"   validate:"required,min=1"`
	TemplateRoleTree t_rbac_template.TemplateRoleTree `json:"templatetree" validate:"required,min=3"`
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
		input  Input
		output Output
	)
	if err := c.BindInput(&input); err != nil {
		return c.RESULT_PARAMETER_ERROR(err.Error())
	}
	newrole := t_rbac_role.Role{}
	newroleUpdate := map[string]interface{}{}
	if roles, err := t_rbac_role.FindRolesByIds(c.Mysql(), []uint64{input.RoleId}); err != nil || len(roles) != 1 {
		return c.RESULT_ERROR(ERR_ROLE_NOT_EXIST, "role not exist")
	} else {
		newrole = roles[0]
		currentTime := time.Now().Local()
		nowFormat := currentTime.Format("2006-01-02 15:04:05.000")
		log.PrintPreety("newrole:", newrole)
		newroleUpdate = map[string]interface{}{
			"F_name":        input.Name,
			"F_description": input.Description,
			"F_modify_time": nowFormat,
		}
	}
	//
	// TODO
	//      checkInnerPermission

	//Step 4. found old roles child role
	var (
		err         error
		oldrolemaps []t_rbac_role_map.RoleMap
		oldids      = []uint64{}
	)
	if oldrolemaps, err = t_rbac_role_map.FindChildrenMap(c.Mysql(), []uint64{input.RoleId}); err != nil {
		return c.RESULT_ERROR(ERR_INNER_ERROR, "Find RoleMaps failed")
	} else {
		for _, oldrolemap := range oldrolemaps {
			oldids = append(oldids, oldrolemap.F_id)
		}
	}

	//Step 5. get new ids
	ids := getTemplateRoleDetail.GetAllSelectedIdsFromTemplateRoleTree(&input.TemplateRoleTree)

	//Step 6. diff
	inserts, deletes, updates := getDiffIds(oldrolemaps, ids)

	//Step 7. save
	txsuccess := false
	tx := c.Mysql().Begin()
	defer func() {
		if !txsuccess {
			tx.Rollback()
		}
	}()

	//Step 7.1 New
	rolemaps := []*t_rbac_role_map.RoleMap{}
	for _, newrolemap := range inserts {
		newrolemap.F_role_id = input.RoleId
		newrolemap.F_creator = input.USER.Userid
		newrolemap.F_modifier = input.USER.Userid
		rolemaps = append(rolemaps, newrolemap)
	}
	t_rbac_role_map.CreateRoleMaps(tx, rolemaps)

	//Step 7.2 Update
	rolemaps = []*t_rbac_role_map.RoleMap{}
	for _, newrolemap := range updates {
		newrolemap.F_modifier = input.USER.Userid
		rolemaps = append(rolemaps, newrolemap)
	}
	t_rbac_role_map.UpdateRoleMapsToStatus(tx, rolemaps)

	//Step 7.3 Delete
	rolemaps = []*t_rbac_role_map.RoleMap{}
	for _, newrolemap := range deletes {
		newrolemap.F_modifier = input.USER.Userid
		rolemaps = append(rolemaps, newrolemap)
	}
	t_rbac_role_map.UpdateRoleMapsToStatus(tx, rolemaps)

	//Step 7.4 Update role
	if err := tx.Model(&newrole).Updates(newroleUpdate).Error; err != nil {
		log.Debugf("update role failed:%s", err.Error())
		return c.RESULT_ERROR(ERR_INNER_ERROR, "inner error, update role failed")
	}

	//Step 8. set success
	output.Code = 0
	output.Msg = "success"
	if rtx := tx.Commit(); rtx.Error != nil {
		log.Debugf("update role failed:%s", rtx.Error.Error())
		return c.RESULT_ERROR(ERR_INNER_ERROR, "inner error, transaction commit failed")
	} else {
		txsuccess = true
		log.Debugf("update role success")
	}
	return c.RESULT(output)
}

func getDiffIds(oldrolemaps []t_rbac_role_map.RoleMap, newids []checkPermission.NodeId) (inserts, deletes, updates []*t_rbac_role_map.RoleMap) {
	//Step 1. add to map
	newidscache := map[checkPermission.NodeId]checkPermission.NodeId{}
	for _, id := range newids {
		newidscache[id] = id
	}
	log.PrintPreety("getDiffIds: newids", newids)
	log.PrintPreety("getDiffIds: oldrolemaps", oldrolemaps)

	//Step 2. travel old
	for _, rolemap := range oldrolemaps {
		id := checkPermission.NodeId{rolemap.F_target_id, rolemap.F_target_type}
		log.Debugf("travel oldrolemaps, id:%s", id.String())
		if rolemap.F_status == t_rbac_role_map.ROLE_STATUS_OK {
			if _, ok := newidscache[id]; ok {
				//Nothing
				log.Debugf("新旧都有，不需做任何事情")
			} else {
				//Delete
				log.Debugf("新无旧有，加入deletes")
				deletes = append(deletes, &t_rbac_role_map.RoleMap{
					F_id:     rolemap.F_id,
					F_status: t_rbac_role_map.ROLE_STATUS_DELETED,
				})
			}
		} else {
			if _, ok := newidscache[id]; ok {
				//Update
				log.Debugf("新有旧已删，加入updates")
				updates = append(updates, &t_rbac_role_map.RoleMap{
					F_id:     rolemap.F_id,
					F_status: t_rbac_role_map.ROLE_STATUS_OK,
				})
			} else {
				//Nothing
				log.Debugf("新已删旧已删，不需做任何事情")
			}
		}

		//delete from newrolemap
		delete(newidscache, id)
	}

	//Step 3. travel new remind
	for rid, _ := range newidscache {
		inserts = append(inserts, &t_rbac_role_map.RoleMap{
			F_target_id:   rid.Id,
			F_target_type: rid.Type,
			F_status:      t_rbac_role_map.ROLE_STATUS_OK,
		})
	}

	return inserts, deletes, updates
}
