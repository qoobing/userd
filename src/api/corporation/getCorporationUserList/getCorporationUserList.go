package getCorporationUserList

import (
	"fmt"
	"github.com/labstack/echo"
	"github.com/qoobing/userd/src/model/t_corp_corporation_user"
	"github.com/qoobing/userd/src/model/t_rbac_role"
	"github.com/qoobing/userd/src/model/t_rbac_user_role"
	"github.com/qoobing/userd/src/model/t_user"
	. "qoobing.com/utillib.golang/api"
	"qoobing.com/utillib.golang/xyz"
	"time"
)

type Input struct {
	USER      UInfo  `json:"-"`
	UATK      string `json:"UATK"         validate:"omitempty,min=4"`
	CorpId    uint64 `json:"corpid"       validate:"required,min=1"`          // 企业ID
	Page      int    `json:"page"         validate:"required,min=1"`          // 页码数
	PageCount int    `json:"pagecount"    validate:"omitempty,min=1,max=100"` // 每页条数
}

type Output struct {
	Code  int               `json:"code"`
	Msg   string            `json:"msg"`
	Count int               `json:"count"` // 总记录数
	Pages int               `json:"pages"` // 总页数
	Data  []*outputDataItem `json:"data"`
}

type outputDataItem struct {
	No         string            `json:"NO"`         //序号
	Id         uint64            `json:"id"`         //用户ID
	Name       string            `json:"name"`       //名称
	Status     int               `json:"status"`     //状态
	StatusStr  string            `json:"statusstr"`  //状态
	Phone      string            `json:"phone"`      //手机
	Roles      map[uint64]string `json:"roles"`      //所属角色列表
	CreateTime int64             `json:"createtime"` //创建时间
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
		outusers = map[uint64]*outputDataItem{}
	)
	if err := c.BindInput(&input); err != nil {
		return c.RESULT_PARAMETER_ERROR(err.Error())
	}
	input.PageCount = xyz.IF(input.PageCount <= 0, 20, input.PageCount).(int)

	//step 3. get corporation users
	var (
		offset    = (input.Page - 1) * input.PageCount
		limit     = input.PageCount
		no        = (input.Page - 1) * input.PageCount
		corpusers = []t_corp_corporation_user.CorporationUser{}
		users     = []t_user.User{}
		err       error
	)
	if corpusers, err = t_corp_corporation_user.GetCorporationUsers(c.Mysql(), input.CorpId, offset, limit); err != nil {
		return c.RESULT_ERROR(ERR_INNER_ERROR, err.Error())
	}
	userids := []uint64{}
	for _, cu := range corpusers {
		_, ok := outusers[cu.F_user_id]
		if !ok {
			no++
			userids = append(userids, cu.F_user_id)
			timestamp, _ := time.Parse(time.RFC3339, cu.F_create_time)
			outusers[cu.F_user_id] = &outputDataItem{
				No:         fmt.Sprintf("%03d", no),
				Id:         cu.F_user_id,
				Name:       "",
				Status:     cu.F_status,
				StatusStr:  cu.GetCorporationUserStatusStr(),
				Phone:      "",
				Roles:      map[uint64]string{},
				CreateTime: timestamp.Unix(),
			}
		}
	}

	//Step 4. get users detail
	if users, err = t_user.GetUsersByUserids(c.Mysql(), userids); err == nil {
		for _, u := range users {
			outusers[u.F_user_id].Name = u.F_name
			outusers[u.F_user_id].Phone = u.F_mobile
		}
	}

	//Step 5. get users roles
	rolesname := map[uint64]string{}
	userroles := []t_rbac_user_role.UserRole{}
	scenekey := t_rbac_role.SceneKeyGYCORPID(input.CorpId)
	if userroles, err = t_rbac_user_role.FindUsersRolesByScenekeyAndUserids(c.Mysql(), scenekey, userids); err != nil {
		return c.RESULT_ERROR(ERR_INNER_ERROR, err.Error())
	}
	for _, ur := range userroles {
		rolesname[ur.F_role_id] = ""
		outusers[ur.F_user_id].Roles[ur.F_role_id] = ""
	}
	roleids := []uint64{}
	for rid, _ := range rolesname {
		roleids = append(roleids, rid)
	}
	roles := []t_rbac_role.Role{}
	if roles, err = t_rbac_role.FindRolesByIds(c.Mysql(), roleids); err != nil {
		return c.RESULT_ERROR(ERR_INNER_ERROR, err.Error())
	}
	for _, role := range roles {
		rolesname[role.F_id] = role.F_name
	}
	for _, ou := range outusers {
		for rid, _ := range ou.Roles {
			ou.Roles[rid] = rolesname[rid]
		}
		output.Data = append(output.Data, ou)
	}

	//Step 5. set success
	output.Code = 0
	output.Msg = "success"

	return c.RESULT(output)
}
