package getRoleDetail

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
	if roles, err := t_rbac_role.FindRolesByIds(c.Mysql(), []uint64{input.Id}); err != nil {
		return c.RESULT_ERROR(ERR_PARAMETER_INVALID, err.Error())
	} else if len(roles) != 1 {
		return c.RESULT_ERROR(ERR_PRIVILEGE_NOT_EXIST, "role not exist")
	} else {
		role := roles[0]
		timestamp, _ := time.Parse(time.RFC3339, role.F_modify_time)
		output.Data = OutputData{
			Id:          role.F_id,
			Name:        role.F_name,
			Status:      role.F_status,
			Description: role.F_description,
			ModifyTime:  timestamp.Unix(),
			Dag:         getDagByRole(c.Mysql(), role),
		}
	}

	//Step 5. set success
	output.Code = 0
	output.Msg = "success"
	log.Debugf("get privilege detail success")
	return c.RESULT(output)
}

func getDagByRole(db *gorm.DB, role t_rbac_role.Role) (g dag) {
	//Step 0. current node
	curStyle := "fill: #7f7"
	curName := "R" + strconv.FormatUint(role.F_id, 10)
	nodesMap := map[string]*dagNode{curName: &dagNode{
		Name:  curName,
		Label: role.F_name,
		Style: &curStyle,
	}}
	g.Nodes = []*dagNode{nodesMap[curName]}
	edgesMap := map[string]map[string]*dagEdge{
		curName: map[string]*dagEdge{},
	}

	//Step 1. find role parents
	allroleids := []uint64{}
	ids := []t_rbac_role_map.MapId{{role.F_id, t_rbac_role_map.TARGET_TYPE_ROLE_TO_ROLE}}
	for len(ids) != 0 {
		rolemaps, _ := t_rbac_role_map.FindParentsMap(db, ids)
		ids = []t_rbac_role_map.MapId{}
		for _, rm := range rolemaps {
			name := "R" + strconv.FormatUint(rm.F_role_id, 10)
			dname := strconv.FormatUint(rm.F_target_id, 10)
			if rm.F_target_type == t_rbac_role_map.TARGET_TYPE_ROLE_TO_ROLE {
				dname = "R" + dname
			} else {
				dname = "P" + dname
			}

			if _, ok := nodesMap[name]; !ok {
				ids = append(ids, t_rbac_role_map.MapId{rm.F_role_id, t_rbac_role_map.TARGET_TYPE_ROLE_TO_ROLE})
				allroleids = append(allroleids, rm.F_role_id)
				nodesMap[name] = &dagNode{
					Name:  name,
					Label: strconv.FormatUint(rm.F_role_id, 10),
				}
				g.Nodes = append(g.Nodes, nodesMap[name])
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

	//Step 2. find role children
	roleids := []uint64{role.F_id}
	allprivilegeids := []uint64{}
	for len(roleids) != 0 {
		rolemaps, _ := t_rbac_role_map.FindChildrenMap(db, roleids)
		roleids = []uint64{}
		for _, rm := range rolemaps {
			name := "R" + strconv.FormatUint(rm.F_role_id, 10)
			dname := strconv.FormatUint(rm.F_target_id, 10)
			if rm.F_target_type == t_rbac_role_map.TARGET_TYPE_ROLE_TO_ROLE {
				dname = "R" + dname
			} else {
				dname = "P" + dname
			}
			if _, ok := nodesMap[dname]; !ok {
				nodesMap[dname] = &dagNode{
					Name:  dname,
					Label: strconv.FormatUint(rm.F_target_id, 10),
				}
				g.Nodes = append(g.Nodes, nodesMap[dname])
				if rm.F_target_type == t_rbac_role_map.TARGET_TYPE_ROLE_TO_ROLE {
					roleids = append(roleids, rm.F_target_id)
					allroleids = append(allroleids, rm.F_target_id)
				} else {
					allprivilegeids = append(allprivilegeids, rm.F_target_id)
				}
			}
			if _, ok := edgesMap[dname]; !ok {
				edgesMap[dname] = map[string]*dagEdge{}
			}
			if _, ok := edgesMap[dname][name]; !ok {
				edgesMap[dname][name] = &dagEdge{
					Src: name,
					Dst: dname,
				}
				g.Edges = append(g.Edges, edgesMap[dname][name])
			}
		}
	}

	//Step 2. find all roles
	cacheInfo := map[string]interface{}{}
	roles, _ := t_rbac_role.FindRolesByIds(db, allroleids)
	for i, r := range roles {
		name := "R" + strconv.FormatUint(r.F_id, 10)
		cacheInfo[name] = &roles[i]
	}
	privileges, _ := t_rbac_privilege.FindPrivilegeByIds(db, allprivilegeids)
	for i, p := range privileges {
		name := "P" + strconv.FormatUint(p.F_id, 10)
		cacheInfo[name] = &privileges[i]
	}
	for _, node := range g.Nodes {
		if i, ok := cacheInfo[node.Name]; !ok {
			log.Warningf("role/privilege not exist by node name [%s]", node.Name)
		} else if r, ok := i.(*t_rbac_role.Role); ok {
			node.Label = fmt.Sprintf("id:%s\nname:%s", node.Name, r.F_name)
		} else if p, ok := i.(*t_rbac_privilege.Privilege); ok {
			node.Label = fmt.Sprintf("id:%s\nname:%s", node.Name, p.F_name)
		}
	}
	return
}
