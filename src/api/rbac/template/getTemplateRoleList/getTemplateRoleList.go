package getTemplateRoleList

import (
	"github.com/labstack/echo"
	"github.com/qoobing/userd/src/model/t_rbac_role"
	. "qoobing.com/utillib.golang/api"
	"qoobing.com/utillib.golang/log"
	"time"
)

type Input struct {
	USER     UInfo  `json:"-"`
	UATK     string `json:"UATK"         validate:"omitempty,min=4"`
	SceneKey string `json:"scenekey"     validate:"required,min=1"`
}

type Output struct {
	Code  int              `json:"code"`
	Msg   string           `json:"msg"`
	Count int              `json:"count"` // 总记录数
	Pages int              `json:"pages"` // 总页数
	Data  []OutputDataItem `json:"data"`
}
type OutputDataItem struct {
	Id          uint64 `json:"id"`          //角色id
	Name        string `json:"name"`        //名称
	Status      int    `json:"status"`      //角色状态
	Description string `json:"description"` //角色描述
	UserCount   int    `json:"usercount"`   //账号数量
	CreateTime  int64  `json:"createtime"`  //创建时间
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

	//Step 3. find role
	if roles, err := t_rbac_role.FindRolesBySceneKey(c.Mysql(), input.SceneKey); err != nil {
		return c.RESULT_ERROR(ERR_PARAMETER_INVALID, err.Error())
	} else {
		output.Pages = 1
		output.Count = len(roles)
		log.PrintPreety("roles=", roles)
		for _, role := range roles {
			timestamp, _ := time.Parse(time.RFC3339, role.F_create_time)
			newdataitem := OutputDataItem{
				Id:          role.F_id,
				Name:        role.F_name,
				Status:      role.F_status,
				Description: role.F_description,
				UserCount:   0,
				CreateTime:  timestamp.Unix(),
			}
			output.Data = append(output.Data, newdataitem)
		}
	}

	//Step 5. set success
	output.Code = 0
	output.Msg = "success"
	log.Debugf("get template list success")
	return c.RESULT(output)
}
