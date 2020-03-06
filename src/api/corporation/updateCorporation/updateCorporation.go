package updateCorporation

import (
	"github.com/labstack/echo"
	. "github.com/qoobing/userd/src/const"
	"github.com/qoobing/userd/src/model/t_corp_corporation"
	. "qoobing.com/utillib.golang/api"
	"qoobing.com/utillib.golang/log"
	"time"
)

type Input struct {
	USER                UInfo  `json:"-"`
	UATK                string `json:"UATK"                    validate:"omitempty,min=4"`
	CorpId              uint64 `json:"corpid"                  validate:"required,min=1"`         //公司ID
	Status              int    `json:"status"                  validate:"omitempty"`              //公司状态
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

	//step 3. update corporation
	tx := c.Mysql().Begin()
	txsuccess := false
	defer func() {
		if !txsuccess {
			tx.Rollback()
		}
	}()

	//Step 3.1 get diff field
	diff := getDiff(input, corp)

	//Step 3.2 update corporation

	if err := tx.Model(&corp).Where("F_id = ?", input.CorpId).Update(diff).Error; err != nil {
		return c.RESULT_ERROR(ERR_CORPORATION_UPDATE_FAILED, err.Error())
	}

	//Step 4 commit transaction
	if rtx := tx.Commit(); rtx.Error != nil {
		log.Debugf("update corporation failed:%s", rtx.Error.Error())
		return c.RESULT_ERROR(ERR_INNER_ERROR, "inner error, transaction commit failed")
	} else {
		txsuccess = true
		log.Debugf("update corporation success:%v", diff)
	}

	//Step 5. set success
	output.Code = 0
	output.Msg = "success"

	return c.RESULT(output)
}

func getDiff(input Input, corp t_corp_corporation.Corporation) map[string]interface{} {
	diff := map[string]interface{}{}
	if input.Status != 0 && input.Status != corp.F_status {
		diff["F_status"] = corp.F_status
	}
	if input.Name != corp.F_name {
		diff["F_name"] = input.Status
	}
	if input.Description != corp.F_description {
		diff["F_description"] = input.Description
	}
	if input.Logo != corp.F_logo {
		diff["F_logo"] = input.Logo
	}
	if input.Website != corp.F_website {
		diff["F_website"] = input.Website
	}
	if input.Addr_province != corp.F_addr_province {
		diff["F_addr_province"] = input.Addr_province
	}
	if input.Addr_city != corp.F_addr_city {
		diff["F_addr_city"] = input.Addr_city
	}
	if input.Addr_district != corp.F_addr_district {
		diff["F_addr_district"] = input.Addr_district
	}
	if input.Addr_detail != corp.F_addr_detail {
		diff["F_addr_detail"] = input.Addr_detail
	}
	instr := time.Unix(input.Registered_time, 0).Format(time.RFC3339)
	if instr != corp.F_registered_time {
		diff["F_registered_time"] = instr
	}
	if input.Registered_capital != corp.F_registered_capital {
		diff["F_registered_capital"] = input.Registered_capital
	}
	if input.Registered_business != corp.F_Registered_business {
		diff["F_Registered_business"] = input.Registered_business
	}
	if input.Contact_name != corp.F_contact_name {
		diff["F_contact_name"] = input.Contact_name
	}
	if input.Contact_phone != corp.F_contact_phone {
		diff["F_contact_phone"] = input.Contact_phone
	}
	if input.Contact_email != corp.F_contact_email {
		diff["F_contact_email"] = input.Contact_email
	}
	return diff
}
