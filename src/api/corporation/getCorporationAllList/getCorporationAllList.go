package getCorporationAllList

import (
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/qoobing/userd/src/const"
	"github.com/qoobing/userd/src/model/t_corp_corporation"
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
	CreatorPhone    string `json:"creatorPhone"        validate:"omitempty"` // 查询条件：注册手机
	AdminPhone      string `json:"adminPhone"          validate:"omitempty"` // 查询条件：管理员手机
	Name            string `json:"name"                validate:"omitempty"` // 查询条件：公司名称
	Type            int64  `json:"type"                validate:"omitempty"` // 查询条件：公司类型
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
	CorpId              uint64 `json:"corpid"`              //企业ID
	CorpNo              string `json:"corpNo"`              //企业编号
	CorpName            string `json:"corpName"`            //企业名称
	AdminId             uint64 `json:"adminId"`             //管理员ID
	AdminName           string `json:"adminName"`           //管理员名称(手机号)
	CorpType            int64  `json:"corpType"`            //企业类型
	CorpTypeStr         string `json:"corpTypeStr"`         //企业类型字符串
	CorpStatus          int    `json:"corpStatus"`          //企业状态
	CorpStatusStr       string `json:"corpStatusStr"`       //企业状态解释-字符串
	CorpEnableStatusStr string `json:"corpEnableStatusStr"` //企业状态解释-可用状态字符串（启用|禁用）
	CorpVerifyStatusStr string `json:"corpVerifyStatusStr"` //企业状态解释-审核状态字符串（审核中|审核通过）
	CorpVerifiedTime    uint64 `json:"corpVerifiedTime"`    //企业审核认证通过时间
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
	input.PageCount = xyz.IF(input.PageCount <= 0, 20, input.PageCount).(int)

	//step 3. get corporation users
	var (
		offset   = (input.Page - 1) * input.PageCount
		limit    = input.PageCount
		corps    = []t_corp_corporation.Corporation{}
		users    = []t_user.User{}
		outusers = map[uint64]*t_user.User{}
		err      error
	)
	output.Code = 0
	output.Msg = "success"
	var rdb *gorm.DB = c.Mysql()
	if input.Conditions.CreatorPhone != "" {
		if user, err := t_user.FindUser(c.Mysql(), input.Conditions.CreatorPhone, _const.NAMETYPE_MOBILE); err != nil && err.Error() != USER_NOT_EXIST {
			return c.RESULT_ERROR(ERR_INNER_ERROR, err.Error())
		} else if err == t_user.ERROR_USER_NOT_EXIST {
			return c.RESULT(output)
		} else {
			rdb = rdb.Where("F_creator = ?", user.F_user_id)
		}
	}
	if input.Conditions.AdminPhone != "" {
		if user, err := t_user.FindUser(c.Mysql(), input.Conditions.CreatorPhone, _const.NAMETYPE_MOBILE); err != nil && err.Error() != USER_NOT_EXIST {
			return c.RESULT_ERROR(ERR_INNER_ERROR, err.Error())
		} else if err == t_user.ERROR_USER_NOT_EXIST {
			return c.RESULT(output)
		} else {
			rdb = rdb.Where("F_administrator = ?", user.F_user_id)
		}
	}
	if input.Conditions.Name != "" {
		rdb = rdb.Where("F_name LIKE ?", input.Conditions.Name)
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

	if err := rdb.Find(&corps).Order("F_create_time DESC").Offset(offset).Limit(limit).Error; err != nil {
		log.Fatalf("find corporations error:%s", err.Error())
		return c.RESULT(output)
	}

	userids := []uint64{}
	for _, cu := range corps {
		if _, ok := outusers[cu.F_creator]; !ok {
			userids = append(userids, cu.F_creator)
			outusers[cu.F_creator] = &t_user.User{}
		}
		if _, ok := outusers[cu.F_administrator]; !ok {
			userids = append(userids, cu.F_administrator)
			outusers[cu.F_administrator] = &t_user.User{}
		}
	}

	//Step 4. get users detail
	if users, err = t_user.GetUsersByUserids(c.Mysql(), userids); err == nil {
		for i, u := range users {
			outusers[u.F_user_id] = &users[i]
		}
	}
	for _, c := range corps {
		st := t_corp_corporation.Status(c.F_status)
		adminName := "未找到"
		if _, ok := outusers[c.F_administrator]; !ok {
			log.Fatalf("Can not found output administrator user by id: [%d]", c.F_administrator)
			adminName = outusers[c.F_administrator].F_name
		}
		oitem := &outputDataItem{
			CorpId:              c.F_id,
			CorpNo:              "20190101",
			CorpName:            c.F_name,
			AdminId:             c.F_administrator,
			AdminName:           adminName,
			CorpType:            c.F_type,
			CorpTypeStr:         t_corp_corporation.TypeToTypeStr(c.F_type),
			CorpStatus:          c.F_status,
			CorpStatusStr:       st.String(),
			CorpEnableStatusStr: st.EnableString(),
			CorpVerifyStatusStr: st.VerifyString(),
			CorpVerifiedTime:    0,
		}
		output.Data = append(output.Data, oitem)
	}

	//Step 5. set success
	return c.RESULT(output)
}
