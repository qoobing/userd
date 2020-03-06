/***********************************************************************
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php.
//******
// Filename:
// Description:
// Author:
// CreateTime:
/***********************************************************************/
package checkPermission

import (
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"net/url"
	. "github.com/qoobing/userd/src/config"
	. "github.com/qoobing/userd/src/const"
	"github.com/qoobing/userd/src/model/t_rbac_privilege"
	"github.com/qoobing/userd/src/model/t_rbac_role"
	"github.com/qoobing/userd/src/model/t_rbac_role_map"
	"github.com/qoobing/userd/src/model/t_rbac_user_role"
	"github.com/qoobing/userd/src/model/t_user"
	. "qoobing.com/utillib.golang/api"
	"qoobing.com/utillib.golang/log"
	"qoobing.com/utillib.golang/val"
)

type Input struct {
	Userid uint64 `form:"userid" json:"userid"    validate:"required,min=1"`
	Uri    string `form:"uri"    json:"uri"       validate:"required,min=1"`
	Args   string `form:"args"   json:"args"      validate:"omitempty,min=1"`
}

type Output struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
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

	//Step 3. 获取基础信息
	user, err := t_user.GetUserByUserid(c.Mysql(), input.Userid)
	if err != nil {
		log.Debugf("failed GetUserByUserid(%d)", input.Userid)
	} else {
		log.Debugf("success GetUserByUserid(%d), user:[%+v]", input.Userid, user)
	}
	args := map[string]interface{}{}
	if input.Args == "" {
		//log.Debugf("Not args input")
	} else if err := json.Unmarshal([]byte(input.Args), &args); err == nil {
		//log.Debugf("args input is json")
	} else if values, err := url.ParseQuery(input.Args); err == nil {
		//log.Debugf("args input is querystring")
		for k, varray := range values {
			if varray != nil {
				args[k] = varray[0]
			}
		}
	}

	//Step 4. 获取用户角色树
	roletree, err := GetUserRoleTreeFromDb(c.Mysql(), input.Userid)
	if err != nil {
		return c.RESULT_ERROR(ERR_PERMISSION_DENIED, "获取用户角色树错误")
	}
	log.PrintPreety("roletree", roletree)

	//Step 5. 获取所有权限
	privilegeids := GetPrivilegeIds(roletree)
	if len(privilegeids) == 0 {
		return c.RESULT_ERROR(ERR_PERMISSION_DENIED, "用户无任何权限")
	}
	privileges, err := t_rbac_privilege.FindPrivilegeByIds(c.Mysql(), privilegeids)
	if err != nil {
		return c.RESULT_ERROR(ERR_PERMISSION_DENIED, "获取权限列表失败")
	}

	//Step 6. 判断用户有权限
	pass := false
	for _, privilege := range privileges {
		if privilege.F_uri == input.Uri {
			if ok, err := CheckPrivilege(privilege.F_expression, user, args); ok && err == nil {
				pass = true
			}
		}
	}

	//Step 4. set output
	if pass {
		output.Code = 0
		output.Msg = "success"
	} else {
		output.Code = ERR_PERMISSION_DENIED
		output.Msg = "Permission denied"
	}

	return c.RESULT(output)
}

type NodeId struct {
	Id   uint64
	Type int
}
type NodeInfo struct {
	Id       NodeId
	Name     string
	Parents  RoleTree
	Children RoleTree
}
type RoleTree map[NodeId]*NodeInfo

func (roletree RoleTree) String() string {
	ret := "roletree:\n"
	for id, rolemap := range roletree {
		ret += fmt.Sprintf("==  %s->%s,parents:%d,children:%d\n",
			id.String(), rolemap.Name, len(rolemap.Parents), len(rolemap.Children))
	}
	ret += "=="
	return ret
}

func (nodeid *NodeId) String() string {
	if nodeid.Type == 1 {
		return fmt.Sprintf("ROLE%d", nodeid.Id)
	} else {
		return fmt.Sprintf("PRIVILEGE%d", nodeid.Id)
	}
}

func GetUserRoleTreeFromDb(db *gorm.DB, userid uint64) (roletree RoleTree, err error) {
	//Step 1. 获取用户所有角色ID
	roleids := []uint64{}
	roles := map[uint64]interface{}{}
	userroles, err := t_rbac_user_role.FindUserRoles(db, userid)
	for _, userrole := range userroles {
		if id, ok := roles[userrole.F_role_id]; ok {
			log.Warningf("User role %d duplicated", id)
		} else {
			roleids = append(roleids, userrole.F_role_id)
			roles[userrole.F_role_id] = userrole.F_role_id
		}
	}

	//Step 2. 获取子树
	roletree = GetRoleTreeByRoleIds(db, roleids)
	return
}

var roleCache = map[NodeId]t_rbac_role.Role{}
var privCache = map[NodeId]t_rbac_privilege.Privilege{}

func getName(db *gorm.DB, nid NodeId) string {
	if nid.Type == t_rbac_role_map.TARGET_TYPE_ROLE_TO_ROLE {
		if v, ok := roleCache[nid]; ok {
			return v.F_name
		} else if roles, err := t_rbac_role.FindRolesByIds(db, []uint64{nid.Id}); err == nil && len(roles) == 1 {
			roleCache[nid] = roles[0]
			return roles[0].F_name
		} else {
			return "ROLE_NOT_EXIST"
		}
	} else if nid.Type == t_rbac_role_map.TARGET_TYPE_ROLE_TO_PRIVILEGE {
		if v, ok := privCache[nid]; ok {
			return v.F_name
		} else if privileges, err := t_rbac_privilege.FindPrivilegeByIds(db, []uint64{nid.Id}); err == nil && len(privileges) == 1 {
			privCache[nid] = privileges[0]
			return privileges[0].F_name
		} else {
			return "PRIVILEGE_NOT_EXIST"
		}
	}
	return "UNKOWN_TYPE_ERROR"
}

func GetRoleTreeByRoleIds(db *gorm.DB, roleids []uint64) (roletree RoleTree) {
	roletree = RoleTree{}
	roles := map[uint64]interface{}{}
	for _, rid := range roleids {
		roles[rid] = rid
	}

	for len(roleids) != 0 {
		rolemaps, _ := t_rbac_role_map.FindChildrenMap(db, roleids)
		roleids = []uint64{}

		for _, rolemap := range rolemaps {
			//父节点ID
			pnid := NodeId{rolemap.F_role_id, t_rbac_role_map.TARGET_TYPE_ROLE_TO_ROLE}
			//子节点ID
			cnid := NodeId{rolemap.F_target_id, rolemap.F_target_type}

			//父节点不在树内， 新建一个顶级节点
			if _, ok := roletree[pnid]; !ok {
				roletree[pnid] = &NodeInfo{
					Id:       pnid,
					Name:     getName(db, pnid),
					Children: RoleTree{},
					Parents:  RoleTree{},
				}
			}
			//添加子节点
			if child, ok := roletree[cnid]; ok {
				//子节点已存在，直接添加
				roletree[pnid].Children[cnid] = child
				roletree[cnid].Parents[pnid] = roletree[pnid]
			} else {
				//子节点不存在，新建添加
				roletree[cnid] = &NodeInfo{
					Id:       cnid,
					Name:     getName(db, cnid),
					Children: RoleTree{},
					Parents:  RoleTree{},
				}
				roletree[pnid].Children[cnid] = roletree[cnid]
				roletree[cnid].Parents[pnid] = roletree[pnid]
			}

			//获取下一层级的节点信息
			if rolemap.F_target_type == t_rbac_role_map.TARGET_TYPE_ROLE_TO_ROLE {
				target_id := rolemap.F_target_id
				if id, ok := roles[target_id]; ok {
					log.Debugf("role %d cached", id)
				} else {
					roleids = append(roleids, target_id)
					roles[target_id] = target_id
				}
			}
		}
	}
	return
}

func GetPrivilegeIds(roletree RoleTree) (privilegeids []uint64) {
	cached := map[uint64]interface{}{}
	for _, node := range roletree {
		if node.Id.Type == t_rbac_role_map.TARGET_TYPE_ROLE_TO_PRIVILEGE {
			if _, ok := cached[node.Id.Id]; ok {
				log.Debugf("privilege %d cached", node.Id.Id)
			} else {
				privilegeids = append(privilegeids, node.Id.Id)
				cached[node.Id.Id] = node.Id.Id
			}
		}
	}
	return
}

func GetRoleIds(roletree RoleTree) (roleids []uint64) {
	cached := map[uint64]interface{}{}
	for _, node := range roletree {
		if node.Id.Type == t_rbac_role_map.TARGET_TYPE_ROLE_TO_ROLE {
			if _, ok := cached[node.Id.Id]; ok {
				log.Debugf("role %d cached", node.Id.Id)
			} else {
				roleids = append(roleids, node.Id.Id)
				cached[node.Id.Id] = node.Id.Id
			}
		}
	}
	return
}

func CheckPrivilege(logicexp string, user t_user.User, args map[string]interface{}) (ok bool, err error) {
	if len(logicexp) == 0 {
		return true, nil
	}

	//Step 1. get need tokens
	tokens, expression, err := GetLogicexpTokensAndExpression(logicexp)
	if err != nil {
		return false, err
	}

	//step 2. get all token value
	for token, value := range args {
		tokens[token] = value
	}
	for token, _ := range tokens {
		switch token {
		case "alluser":
			tokens["alluser"] = true

		case "inneruser":
			inneruser, err := EnvIsInnerUser(user)
			if err != nil {
				log.Panicf("check inneruser failed:%s, treat as not inneruser", err.Error())
			}
			tokens["inneruser"] = inneruser

		case "linuxpamuser":
			ispamuser, err := EnvIsPamUser(user)
			if err != nil {
				log.Panicf("check inneruser failed:%s, treat as not inneruser", err.Error())
			}
			tokens["linuxpamuser"] = ispamuser

		default:
			log.Panicf("not supported token[%s]", token)
		}
	}
	log.PrintPreety("tokens:", tokens)

	//step 3. valuate
	result, err := expression.Evaluate(tokens)
	if err != nil {
		log.Panicf("CheckScenePrivilege Evaluate(%+v) failed:%s", tokens, err.Error())
		return false, err
	}

	return result == true, err
}

func GetLogicexpTokensAndExpression(logicexp string) (tokens map[string]interface{}, expr *govaluate.EvaluableExpression, err error) {
	//Step 1. get need tokens
	tokens = map[string]interface{}{}
	expression, err := govaluate.NewEvaluableExpression(logicexp)
	if err != nil {
		log.Panicf("CheckScenePrivilege NewEvaluableExpression(%s) failed:%s",
			logicexp, err.Error())
		return nil, nil, err
	}
	for _, tk := range expression.Tokens() {
		if tk.Kind == govaluate.VARIABLE {
			tokens[tk.Value.(string)] = 1
		}
	}
	return tokens, expression, nil
}

func EnvIsInnerUser(user t_user.User) (inneruser bool, err error) {
	mysqlcnf := Config().Auth20Login.DingMysqlConf
	mysql, err := gorm.Open("mysql", mysqlcnf)
	if err != nil {
		log.Fatalf("connect mysql[%s] failed [%s]", mysqlcnf, err.Error())
		return false, err
	}
	defer mysql.Close()
	type Result struct {
		F_user_id string
	}
	var (
		tmp = Result{}
		sql = "SELECT F_user_id FROM t_ding_user WHERE F_unionid = ? AND F_user_status = 1"
	)
	rdb := mysql.Raw(sql, user.F_exid_dd_unionid).Scan(&tmp)
	if rdb.RecordNotFound() {
		log.Debugf("can not found user ddid=%s in t_ding_user", user.F_exid_dd_unionid)
		return false, err
	} else if err = rdb.Error; err != nil {
		log.Debugf("find user from t_ding_user error: %s", err.Error())
		return false, err
	}

	return true, nil
}

func EnvIsPamUser(user t_user.User) (inneruser bool, err error) {
	mysqlcnf := Config().Auth20Login.PamMysqlConf

	mysql, err := gorm.Open("mysql", mysqlcnf)
	if err != nil {
		log.Fatalf("connect mysql[%s] failed [%s]", mysqlcnf, err.Error())
		return false, err
	}
	defer mysql.Close()
	type Result struct {
		F_user_id string
	}
	var (
		tmp = Result{}
		sql = "SELECT F_user_id FROM t_pam_user WHERE F_user_id = ? AND F_expire = -1"
	)
	rdb := mysql.Raw(sql, user.F_user_id).Scan(&tmp)
	if rdb.RecordNotFound() {
		log.Debugf("can not found user user_id=%d in t_pam_user", user.F_user_id)
		return false, err
	} else if err = rdb.Error; err != nil {
		log.Debugf("find user from t_pam_user error: %s", err.Error())
		return false, err
	}

	log.Debugf("EnvIsPamUser, sql: %s; result:%v", sql, tmp)
	return true, nil
}

func EnvGetPamUserInfo(userid uint64) (user map[string]string, err error) {
	mysqlcnf := Config().Auth20Login.PamMysqlConf
	mysql, err := gorm.Open("mysql", mysqlcnf)
	if err != nil {
		log.Fatalf("connect mysql[%s] failed [%s]", mysqlcnf, err.Error())
		return user, err
	}
	defer mysql.Close()
	var (
		sql = `
			SELECT 
				F_username 	as username,
				F_password 	as password,
				F_phrase 	as phrase,
				F_uid 		as uid,	
				F_gid 		as gid,	
			FROM 
				t_pam_user 
			WHERE 
				F_user_id = ? 
				AND F_expire = -1
			LIMIT 1
		`
	)
	rdb := mysql.Raw(sql, userid).Scan(&user)
	if rdb.RecordNotFound() {
		log.Debugf("can not found user user_id=%s in t_pam_user", userid)
		return user, fmt.Errorf("user not found by user_id: %d", userid)
	} else if err = rdb.Error; err != nil {
		log.Debugf("find user from t_pam_user error: %s", err.Error())
		return user, err
	}

	return user, nil
}
