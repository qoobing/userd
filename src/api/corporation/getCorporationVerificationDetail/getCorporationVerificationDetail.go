package getCorporationVerificationDetail

import (
	"github.com/labstack/echo"
	. "github.com/qoobing/userd/src/const"
	"github.com/qoobing/userd/src/model/t_corp_corporation"
	. "qoobing.com/utillib.golang/api"
	"time"
)

type Input struct {
	USER   UInfo  `json:"-"`
	UATK   string `json:"UATK"              validate:"omitempty,min=4"`
	CorpId uint64 `json:"corpid"            validate:"required,min=1"` //企业ID
}

type Output struct {
	Code int        `json:"code"`
	Msg  string     `json:"msg"`
	Data outputData `json:"data"`
}

type outputData struct {
	CorpId              uint64 `json:"corpid"`
	Name                string `json:"name"                    validate:"required,min=1"`         //公司名称
	Description         string `json:"description"             validate:"required,min=1"`         //公司简介
	Logo                string `json:"logo"                    validate:"omitempty,min=1"`        //公司图标
	Website             string `json:"website"                 validate:"omitempty,min=1"`        //权限参数逻辑表达式
	Addr_province       string `json:"addr_province"           validate:"required,min=1,max=64"`  //公司地址-省
	Addr_city           string `json:"addr_city"               validate:"required,min=1,max=64"`  //公司地址-市
	Addr_district       string `json:"addr_district"           validate:"omitempty,min=1,max=64"` //公司地址-区
	Addr_detail         string `json:"addr_detail"             validate:"required,min=5,max=128"` //公司地址-详细地址
	Registered_time     int64  `json:"registered_time"         validate:"required,min=0"`         //注册年份
	Registered_capital  int64  `json:"registered_capital"      validate:"required,min=0"`         //注册资本(分)
	Registered_business string `json:"registered_business"     validate:"omitempty,min=0"`        //主营业务
	Contact_name        string `json:"contact_name"            validate:"required,min=1"`         //联系人
	Contact_phone       string `json:"contact_phone"           validate:"required,min=1"`         //联系人电话
	Contact_email       string `json:"contact_email"           validate:"omitempty,min=1"`        //联系人邮箱
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
	corp, err := t_corp_corporation.FindCorporationById(c.Mysql(), input.CorpId)
	if err != nil {
		return c.RESULT_ERROR(ERR_CORPORATION_NOT_EXIST, "企业不存在")
	}

	//step 4. corporation detail
	registered_time, _ := time.Parse(time.RFC3339, corp.F_registered_time)
	output.Data = outputData{
		CorpId:              corp.F_id,
		Name:                corp.F_name,
		Description:         corp.F_description,
		Logo:                corp.F_logo,
		Website:             corp.F_website,
		Addr_province:       corp.F_addr_province,
		Addr_city:           corp.F_addr_city,
		Addr_district:       corp.F_addr_district,
		Addr_detail:         corp.F_addr_detail,
		Registered_time:     registered_time.Unix(),
		Registered_capital:  corp.F_registered_capital,
		Registered_business: corp.F_Registered_business,
		Contact_name:        corp.F_contact_name,
		Contact_phone:       corp.F_contact_phone,
		Contact_email:       corp.F_contact_email,
	}

	//Step 5. set success
	output.Code = 0
	output.Msg = "success"
	output.Data.CorpId = corp.F_id

	return c.RESULT(output)
}
