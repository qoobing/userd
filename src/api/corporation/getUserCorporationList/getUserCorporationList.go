package getUserCorporationList

import (
	"github.com/labstack/echo"
	"github.com/qoobing/userd/src/model/t_corp_corporation"
	"github.com/qoobing/userd/src/model/t_corp_corporation_user"
	. "qoobing.com/utillib.golang/api"
	"yqtc.com/log"
)

type Input struct {
	USER UInfo  `json:"-"`
	UATK string `json:"UATK"         validate:"omitempty,min=4"`
}

type Output struct {
	Code int               `json:"code"`
	Msg  string            `json:"msg"`
	Data []*outputDataItem `json:"data"`
}

type outputDataItem struct {
	CorpId              uint64 `json:"corpid"`              //企业ID
	CorpNo              uint64 `json:"corpNo"`              //企业编号
	CorpName            string `json:"corpName"`            //企业名称
	AdminId             string `json:"adminId"`             //管理员ID
	AdminName           string `json:"adminName"`           //管理员名称(手机号)
	CorpType            int    `json:"corpType"`            //企业类型
	CorpTypeStr         string `json:"corpTypeStr"`         //企业类型字符串
	CorpStatus          int    `json:"corpStatus"`          //企业状态
	CorpStatusStr       string `json:"corpStatusStr"`       //企业状态解释-字符串
	CorpEnableStatusStr string `json:"corpEnableStatusStr"` //企业状态解释-可用状态字符串（启用|禁用）
	CorpVerifyStatusStr string `json:"corpVerifyStatusStr"` //企业状态解释-审核状态字符串（审核中|审核通过）
	CorpVerifiedTime    string `json:"corpVerifiedTime"`    //企业审核认证通过时间
	IsDefaultCorp       bool   `json:"isDefaultCorp"`       //是否是用户默认企业
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

	//step 3. get user corporations
	var (
		usercorps = []t_corp_corporation_user.CorporationUser{}
		err       error
	)
	if usercorps, err = t_corp_corporation_user.FindUserCorporations(c.Mysql(), input.USER.Userid); err != nil {
		return c.RESULT_ERROR(ERR_INNER_ERROR, err.Error())
	}
	corpids := []uint64{}
	for _, uc := range usercorps {
		corpids = append(corpids, uc.F_corporation_id)
	}
	corps := []t_corp_corporation.Corporation{}
	if corps, err = t_corp_corporation.FindCorporationsByIds(c.Mysql(), corpids); err != nil {
		return c.RESULT_ERROR(ERR_INNER_ERROR, err.Error())
	}
	corpmap := map[uint64]*t_corp_corporation.Corporation{}
	for _, c := range corps {
		corpmap[c.F_id] = &t_corp_corporation.Corporation{
			F_name: c.F_name,
		}
	}
	log.PrintPreety("corpmap", corpmap)
	log.PrintPreety("usercorps", usercorps)
	for _, uc := range usercorps {
		output.Data = append(output.Data, &outputDataItem{
			CorpId:   uc.F_corporation_id,
			CorpName: corpmap[uc.F_corporation_id].F_name,
		})
	}

	//Step 5. set success
	output.Code = 0
	output.Msg = "success"

	return c.RESULT(output)
}
