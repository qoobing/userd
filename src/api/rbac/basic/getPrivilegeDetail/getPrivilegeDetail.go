package getPrivilegeDetail

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	. "github.com/qoobing/userd/src/const"
	"github.com/qoobing/userd/src/model/t_rbac_privilege"
	"github.com/qoobing/userd/src/model/t_rbac_role"
	"github.com/qoobing/userd/src/model/t_rbac_role_map"
	. "qoobing.com/utillib.golang/api"
	"qoobing.com/utillib.golang/log"
	"strconv"
	"time"
)

type Input struct {
	USER UInfo  `json:"-"`
	UATK string `json:"UATK"         validate:"omitempty,min=4"`
	Id   uint64 `json:"id"`
}

type Output struct {
	Code int        `json:"code"`
	Msg  string     `json:"msg"`
	Data OutputData `json:"data"`
}

type OutputData struct {
	Id          uint64 `json:"id"`          //id
	Name        string `json:"name"`        //名称
	Status      int    `json:"status"`      //状态
	Description string `json:"description"` //描述
	ModifyTime  int64  `json:"modifytime"`  //修改时间
	Dag         dag    `json:"dag"`         //dag
}
type dag struct {
	Nodes []*dagNode `json:"nodes"` //节点
	Edges []*dagEdge `json:"edges"` //边
}

type dagNode struct {
	Name  string  `json:"name"`
	Label string  `json:"label"`
	Style *string `json:"style,omitempty"`
}

type dagEdge struct {
	Src   string  `json:"src"`
	Dst   string  `json:"dst"`
	Label *string `json:"label,omitempty"`
	Style *string `json:"style,omitempty"`
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

	//Step 3. find role
	if privileges, err := t_rbac_privilege.FindPrivilegeByIds(c.Mysql(), []uint64{input.Id}); err != nil {
		return c.RESULT_ERROR(ERR_PARAMETER_INVALID, err.Error())
	} else if len(privileges) != 1 {
		return c.RESULT_ERROR(ERR_PRIVILEGE_NOT_EXIST, "privilege not exist")
	} else {
		privilege := privileges[0]
		timestamp, _ := time.Parse(time.RFC3339, privilege.F_modify_time)
		output.Data = OutputData{
			Id:          privilege.F_id,
			Name:        privilege.F_name,
			Status:      privilege.F_status,
			Description: privilege.F_description,
			ModifyTime:  timestamp.Unix(),
			Dag:         getDagByPrivilege(c.Mysql(), privilege),
		}
	}

	//Step 5. set success
	output.Code = 0
	output.Msg = "success"
	log.Debugf("get privilege detail success")
	return c.RESULT(output)
}

func getDagByPrivilege(db *gorm.DB, privilege t_rbac_privilege.Privilege) (g dag) {
	//Step 0. current node
	curStyle := "fill: #7f7"
	curName := "P" + strconv.FormatUint(privilege.F_id, 10)
	nodesMap := map[string]*dagNode{curName: &dagNode{
		Name:  curName,
		Label: privilege.F_name,
		Style: &curStyle,
	}}
	g.Nodes = []*dagNode{nodesMap[curName]}
	edgesMap := map[string]map[string]*dagEdge{
		curName: map[string]*dagEdge{},
	}

	//Step 1. find privilege parents
	allroleids := []uint64{}
	ids := []t_rbac_role_map.MapId{{privilege.F_id, t_rbac_role_map.TARGET_TYPE_ROLE_TO_PRIVILEGE}}
	for len(ids) != 0 {
		rolemaps, _ := t_rbac_role_map.FindParentsMap(db, ids)
		ids = []t_rbac_role_map.MapId{}
		for _, rm := range rolemaps {
			name := "R" + strconv.FormatUint(rm.F_role_id, 10)
			dname := strconv.FormatUint(rm.F_target_id, 10)

			if _, ok := nodesMap[name]; !ok {
				ids = append(ids, t_rbac_role_map.MapId{rm.F_role_id, t_rbac_role_map.TARGET_TYPE_ROLE_TO_ROLE})
				allroleids = append(allroleids, rm.F_role_id)
				nodesMap[name] = &dagNode{
					Name:  name,
					Label: strconv.FormatUint(rm.F_role_id, 10),
				}
				g.Nodes = append(g.Nodes, nodesMap[name])
			}

			if rm.F_target_type == t_rbac_role_map.TARGET_TYPE_ROLE_TO_ROLE {
				dname = "R" + dname
			} else {
				dname = "P" + dname
			}

			if _, ok := edgesMap[dname][name]; !ok {
				edgesMap[dname][name] = &dagEdge{
					Src: name,
					Dst: dname,
				}
				g.Edges = append(g.Edges, edgesMap[dname][name])
				if _, ok := edgesMap[name]; !ok {
					edgesMap[name] = map[string]*dagEdge{}
				}
			}
		}
	}

	//Step 2. find all roles
	rolesInfo := map[string]*t_rbac_role.Role{}
	roles, _ := t_rbac_role.FindRolesByIds(db, allroleids)
	for i, r := range roles {
		name := "R" + strconv.FormatUint(r.F_id, 10)
		rolesInfo[name] = &roles[i]
	}
	for _, node := range g.Nodes {
		if r, ok := rolesInfo[node.Name]; ok {
			node.Label = fmt.Sprintf("id:%d\nname:%s", r.F_id, r.F_name)
		} else {
			log.Warningf("role not exist by id [%s]", node.Label)
		}
	}
	return
}
