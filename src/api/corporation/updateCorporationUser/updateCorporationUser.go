package updateCorporationUser

import (
	"github.com/labstack/echo"
	"github.com/qoobing/userd/src/api/corporation/getCorporationUserDetail"
	. "github.com/qoobing/userd/src/const"
	"github.com/qoobing/userd/src/model/t_corp_corporation"
	"github.com/qoobing/userd/src/model/t_corp_corporation_user"
	"github.com/qoobing/userd/src/model/t_rbac_role"
	"github.com/qoobing/userd/src/model/t_rbac_user_role"
	"github.com/qoobing/userd/src/model/t_user"
	. "qoobing.com/utillib.golang/api"
	"qoobing.com/utillib.golang/log"
)

type inputRoleItem getCorporationUserDetail.OutputRoleItem

type Input struct {
	USER        UInfo           `json:"-"`
	UATK        string          `json:"UATK"         validate:"omitempty,min=4"`
	CorpId      uint64          `json:"corpid"       validate:"required,min=1"`        //企业ID
	Phone       string          `json:"phone"        validate:"required,min=11"`       //手机号
	Status      int             `json:"status"       validate:"required,oneof=1 2 10"` //目标状态（1:正常；2:删除；10:禁用）
	Roles       []inputRoleItem `json:"roles"      validate:"omitempty"`               //角色列表
	Description string          `json:"description"  validate:"omitempty"`             //备注信息
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
		corpuser     t_corp_corporation_user.CorporationUser
		user         t_user.User
		err          error
	)
	if err := c.BindInput(&input); err != nil {
		return c.RESULT_PARAMETER_ERROR(err.Error())
	}
	if _, err = t_corp_corporation.FindCorporationById(c.Mysql(), input.CorpId); err != nil {
		return c.RESULT_ERROR(ERR_CORPORATION_NOT_EXIST, "Corporation Not Exist")
	}
	if user, err = t_user.FindUser(c.Mysql(), input.Phone, NAMETYPE_MOBILE); err != nil && err.Error() == USER_NOT_EXIST {
		return c.RESULT_ERROR(ERR_USER_NOT_EXIST, "User Not Exist")
	}
	if corpuser, err = t_corp_corporation_user.FindCorporationUserById(c.Mysql(), input.CorpId, user.F_user_id); err != nil {
		return c.RESULT_ERROR(ERR_CORPORATION_USER_DUPLICATE, "企业中没有该用户")
	}
	for _, r := range input.Roles {
		inputRoleids = append(inputRoleids, r.RoleId)
	}
	scenekey := t_rbac_role.SceneKeyGYCORPID(input.CorpId)
	if roles, err := t_rbac_role.FindRolesBySceneKeyAndIds(c.Mysql(), scenekey, inputRoleids); err != nil {
		return c.RESULT_ERROR(ERR_CORPORATION_UPDATE_USER_FAILED, "修改用户失败")
	} else if len(inputRoleids) != len(roles) {
		return c.RESULT_ERROR(ERR_PARAMETER_INVALID, "角色不存在或者不属于当前场景")
	}

	//step 3. create new corporation
	tx := c.Mysql().Begin()
	txsuccess := false
	defer func() {
		if !txsuccess {
			tx.Rollback()
		}
	}()

	//Step 3.1 update corporation user
	err = tx.Model(&corpuser).Where("F_id = ?", corpuser.F_id).
		Update(map[string]interface{}{
			"F_status":      input.Status,
			"F_description": input.Description,
		}).Error
	if err != nil {
		return c.RESULT_ERROR(ERR_CORPORATION_UPDATE_USER_FAILED, err.Error())
	}

	//Step 3.2 update corporation user roles
	//Step 3.2.1 get all user roles
	existuserroles := []t_rbac_user_role.UserRole{}
	if existuserroles, err = t_rbac_user_role.FindUserRolesByScenekeyAndUserids(tx, scenekey, user.F_user_id); err != nil {
		return c.RESULT_ERROR(ERR_INNER_ERROR, "inner error:"+err.Error())
	}
	selectedroles := []uint64{}
	if len(input.Roles) == 0 {
		for _, r := range existuserroles {
			if r.F_status == t_rbac_user_role.USERROLE_STATUS_OK {
				selectedroles = append(selectedroles, r.F_role_id)
			}
		}
	} else {
		for _, r := range input.Roles {
			if r.Selected {
				selectedroles = append(selectedroles, r.RoleId)
			}
		}
	}

	//Step 3.2.2 diff with input roles
	creates, updates, deletes := getDiff(existuserroles, selectedroles)
	log.PrintPreety("existuserroles:", existuserroles)
	log.PrintPreety("selectedroles:", selectedroles)
	log.PrintPreety("creates:", creates)
	log.PrintPreety("updates:", updates)
	log.PrintPreety("deletes:", deletes)
	if err := t_rbac_user_role.CreateUserRoles(tx, scenekey, user.F_user_id, creates); err != nil {
		return c.RESULT_ERROR(ERR_CORPORATION_UPDATE_USER_FAILED, err.Error())
	}
	if err := t_rbac_user_role.UpdateUserRoleToStatus(tx, updates); err != nil {
		return c.RESULT_ERROR(ERR_CORPORATION_UPDATE_USER_FAILED, err.Error())
	}
	if err := t_rbac_user_role.UpdateUserRoleToStatus(tx, deletes); err != nil {
		return c.RESULT_ERROR(ERR_CORPORATION_UPDATE_USER_FAILED, err.Error())
	}

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

func getDiff(existuserroles []t_rbac_user_role.UserRole, inputroleids []uint64) (creates []uint64, updates, deletes []t_rbac_user_role.UserRole) {
	incache := map[uint64]uint64{}
	excache := map[uint64]uint64{}
	for _, inroleid := range inputroleids {
		incache[inroleid] = inroleid
	}

	for _, eur := range existuserroles {
		if _, ok := incache[eur.F_role_id]; ok {
			if eur.F_status == t_rbac_user_role.USERROLE_STATUS_OK {
				//旧有新也有，nothing
				excache[eur.F_role_id] = eur.F_role_id
			} else {
				//旧没有新有
				creates = append(creates, eur.F_role_id)
			}
		} else {
			if eur.F_status == t_rbac_user_role.USERROLE_STATUS_OK {
				//旧有新没有，delete
				deletes = append(deletes, t_rbac_user_role.UserRole{
					F_id:     eur.F_id,
					F_status: t_rbac_user_role.USERROLE_STATUS_DELETED, //状态
				})
			} else {
				//旧状态不对，新没有，update
				updates = append(updates, t_rbac_user_role.UserRole{
					F_id:     eur.F_id,
					F_status: t_rbac_user_role.USERROLE_STATUS_OK, //状态
				})
			}
		}
	}

	for _, inroleid := range inputroleids {
		if _, ok := excache[inroleid]; ok {
			//旧有新也有，nothing
		} else {
			//旧有新没有，create
			creates = append(creates, inroleid)
		}
	}
	return
}
