package getRoleList

import (
	"github.com/labstack/echo"
	"github.com/qoobing/userd/src/model/t_rbac_role"
	. "qoobing.com/utillib.golang/api"
	"qoobing.com/utillib.golang/log"
	"qoobing.com/utillib.golang/xyz"
	"time"
)

type Input struct {
	USER      UInfo  `json:"-"`
	UATK      string `json:"UATK"         validate:"omitempty,min=4"`
	Page      int    `json:"page"         validate:"required,min=1"`            // 页码数
	PageCount int    `json:"pagecount"    validate:"omitempty,min=1,max=10000"` // 每页条数
}

type Output struct {
	Code  int              `json:"code"`
	Msg   string           `json:"msg"`
	Count int              `json:"count"` // 总记录数
	Pages int              `json:"pages"` // 总页数
	Data  []OutputDataItem `json:"data"`
}

type OutputDataItem struct {
	Id          uint64 `json:"id"`          //id
	Name        string `json:"name"`        //名称
	Status      int    `json:"status"`      //状态
	Description string `json:"description"` //描述
	ModifyTime  int64  `json:"modifytime"`  //修改时间
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
	input.PageCount = xyz.IF(input.PageCount <= 0, 20, input.PageCount).(int)

	//step 3. get corporation users
	offset := (input.Page - 1) * input.PageCount

	//Step 3. find role
	if roles, count, err := t_rbac_role.GetRoleList(c.Mysql(), offset, input.PageCount); err != nil {
		return c.RESULT_ERROR(ERR_PARAMETER_INVALID, err.Error())
	} else {
		output.Pages = 1
		output.Count = count
		log.PrintPreety("roles=", roles)
		for _, role := range roles {
			timestamp, _ := time.Parse(time.RFC3339, role.F_modify_time)
			newdataitem := OutputDataItem{
				Id:          role.F_id,
				Name:        role.F_name,
				Status:      role.F_status,
				Description: role.F_description,
				ModifyTime:  timestamp.Unix(),
			}
			output.Data = append(output.Data, newdataitem)
		}
	}
	//Step 5. set success
	output.Code = 0
	output.Msg = "success"
	log.Debugf("get role list success")
	return c.RESULT(output)
}
