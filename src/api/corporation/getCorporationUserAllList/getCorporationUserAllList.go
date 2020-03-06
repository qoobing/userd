package getCorporationUserAllList

import (
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/qoobing/userd/src/common"
	"github.com/qoobing/userd/src/const"
	"github.com/qoobing/userd/src/model/t_corp_corporation"
	"github.com/qoobing/userd/src/model/t_corp_corporation_user"
	"github.com/qoobing/userd/src/model/t_user"
	. "qoobing.com/utillib.golang/api"
	"qoobing.com/utillib.golang/log"
	"qoobing.com/utillib.golang/xyz"
	"time"
)

type Input struct {
	USER       UInfo           `json:"-"`
	UATK       string          `json:"UATK"         validate:"omitempty,min=4"`
	Page       int             `json:"page"         validate:"required,min=1"`          // 页码数
	PageCount  int             `json:"pagecount"    validate:"omitempty,min=1,max=100"` // 每页条数
	Conditions inputConditions `json:"conditions"   validate:"omitempty"`               // 查询条件
}

type inputConditions struct {
	Phone           string `json:"hone"                validate:"omitempty"` // 查询条件：注册手机
	Type            int64  `json:"type"                validate:"omitempty"` // 查询条件：用户类型
	EnableStatusStr string `json:"enableStatusStr"     validate:"omitempty"` // 查询条件：可用状态字符串（启用|禁用）
	CreateTimeStart int64  `json:"createTimeStart"     validate:"omitempty"` // 查询条件：注册时间最早时间限
	CreateTimeEnd   int64  `json:"createTimeEnd"       validate:"omitempty"` // 查询条件：注册时间最晚时间限
}

type Output struct {
	Code  int               `json:"code"`
	Msg   string            `json:"msg"`
	Count int               `json:"count"` // 总记录数
	Pages int               `json:"pages"` // 总页数
	Data  []*outputDataItem `json:"data"`
}

type outputDataItem struct {
	Id          uint64 `json:"id"`          //用户ID
	Name        string `json:"name"`        //名称
	Status      int    `json:"status"`      //状态
	StatusStr   string `json:"statusstr"`   //状态字符串
	Phone       string `json:"phone"`       //手机
	CreateTime  int64  `json:"createtime"`  //创建时间
	CorpCount   int    `json:"corpcount"`   //关联公司数量
	CorpType    int64  `json:"corptype"`    //类型
	CorpTypeStr string `json:"corptypestr"` //类型
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
		corpusers = []t_corp_corporation_user.CorporationUser{}
		users     = []t_user.User{}
		err       error
	)
	output.Code = 0
	output.Msg = "success"
	var rdb *gorm.DB = c.Mysql()
	if input.Conditions.Phone != "" {
		if user, err := t_user.FindUser(c.Mysql(), input.Conditions.Phone, _const.NAMETYPE_MOBILE); err != nil && err.Error() != USER_NOT_EXIST {
			return c.RESULT_ERROR(ERR_INNER_ERROR, err.Error())
		} else if err == t_user.ERROR_USER_NOT_EXIST {
			return c.RESULT(output)
		} else {
			rdb = rdb.Where("F_creator = ?", user.F_user_id)
		}
	}
	if input.Conditions.Type != 0 {
		rdb = rdb.Where("F_type = ?", input.Conditions.Type)
	}
	if input.Conditions.EnableStatusStr != "" {
		rdb = rdb.Where("F_status IN (?)", t_corp_corporation.GetEnableStringStatuses(input.Conditions.EnableStatusStr))
	}
	if input.Conditions.CreateTimeStart != 0 {
		tstr := time.Unix(input.Conditions.CreateTimeStart, 0).Format("2006-01-02 15:04:05")
		rdb = rdb.Where("F_create_time >= ?", tstr)
	}
	if input.Conditions.CreateTimeEnd != 0 {
		tstr := time.Unix(input.Conditions.CreateTimeEnd, 0).Format("2006-01-02 15:04:05")
		rdb = rdb.Where("F_create_time <= ?", tstr)
	}

	dbcount := rdb.Model(&corpusers).Select("F_id").Group("F_user_id")
	if err = dbcount.Count(&output.Count).Error; err != nil {
		return c.RESULT_ERROR(ERR_INNER_ERROR, err.Error())
	} else {
		output.Pages = output.Count/input.PageCount + 1
	}

	userids := []uint64{}
	//sql := fmt.Sprintf("SELECT %s FROM %s GROUP BY F_create_time ORDER BY F_create_time OFFSET ? LIMIT ?",
	//	"F_user_id, F_create_time",
	//	t_corp_corporation_user.TABLENAME)
	//
	//rows, err := rdb.Raw(sql, offset, limit).Rows() // (*sql.Rows, error)
	//defer rows.Close()
	//for rows.Next() {
	//	F_user_id := uint64(0)
	//	F_create_time := ""
	//	rows.Scan(&F_user_id, &F_create_time)
	//	userids = append(userids, F_user_id)
	//}

	rdb.Model(t_corp_corporation_user.TABLENAME).QueryExpr()
	dbuserids := rdb.Model(&corpusers).Group("F_user_id").
		Order("F_create_time DESC").Offset(offset).Limit(limit)
	if err = dbuserids.Find(&corpusers).Error; err != nil {
		return c.RESULT_ERROR(ERR_INNER_ERROR, err.Error())
	} else {
		for _, cu := range corpusers {
			if _, ok := outusers[cu.F_user_id]; !ok {
				userids = append(userids, cu.F_user_id)
			}
		}
	}

	if err := rdb.Where("F_user_id IN (?)", userids).Find(&corpusers).Error; err != nil {
		log.Fatalf("find corporation all users error:%s", err.Error())
		return c.RESULT(output)
	}

	corpusermap := map[uint64]*t_corp_corporation_user.CorporationUser{}
	userids = []uint64{}
	for i, cu := range corpusers {
		if _, ok := outusers[cu.F_user_id]; !ok {
			userids = append(userids, cu.F_user_id)
			outusers[cu.F_user_id] = &outputDataItem{
				Id:          cu.F_user_id,
				Name:        "",
				Status:      0,
				StatusStr:   "",
				Phone:       "",
				CreateTime:  0,
				CorpCount:   1,
				CorpType:    cu.F_corporation_type,
				CorpTypeStr: "",
			}
		} else {
			outusers[cu.F_user_id].CorpCount++
			outusers[cu.F_user_id].CorpType |= cu.F_corporation_type
		}
		if _, ok := corpusermap[cu.F_user_id]; !ok {
			corpusermap[cu.F_user_id] = &corpusers[i]
		}
	}

	//Step 4. get users detail
	usermap := map[uint64]*t_user.User{}
	if users, err = t_user.GetUsersByUserids(c.Mysql(), userids); err == nil {
		for i, u := range users {
			usermap[u.F_user_id] = &users[i]
		}
	}

	////Step 5. merge data
	for _, ou := range outusers {
		if _, ok := usermap[ou.Id]; ok {
			ou.Name = usermap[ou.Id].F_name
			ou.Status = usermap[ou.Id].F_status
			ou.StatusStr = getUserStatusStr(ou.Status)
			ou.Phone = usermap[ou.Id].F_name
			ou.CreateTime = common.GetTimeStamp(usermap[ou.Id].F_create_time)
			ou.CorpTypeStr = t_corp_corporation.TypeToTypeStr(ou.CorpType)
			output.Data = append(output.Data, ou)
		} else {
			log.Fatalf("Can not found output user by id: [%d]", ou.Id)
			ou.Name = "数据已删除"
			ou.Status = 0
			ou.StatusStr = getUserStatusStr(ou.Status)
			ou.Phone = "数据已删除"
			ou.CreateTime = 0
			ou.CorpTypeStr = ""
			output.Data = append(output.Data, ou)
		}
	}

	//Step 5. set success
	output.Code = 0
	output.Msg = "success"

	return c.RESULT(output)
}

func getUserStatusStr(status int) string {
	switch status {
	case t_user.USER_STATUS_OK:
		return "启用"
	case t_user.USER_STATUS_FORBIDDEN:
		return "禁用"
	default:
		return "-"
	}
	return "-"
}
