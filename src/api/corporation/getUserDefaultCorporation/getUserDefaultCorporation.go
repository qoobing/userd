package getUserDefaultCorporation

import (
	"github.com/labstack/echo"
	"github.com/qoobing/userd/src/api/corporation/createCorporation"
	. "github.com/qoobing/userd/src/const"
	"github.com/qoobing/userd/src/model/t_corp_corporation"
	"github.com/qoobing/userd/src/model/t_corp_corporation_user"
	"github.com/qoobing/userd/src/model/t_user_common_keyvalue"
	. "qoobing.com/utillib.golang/api"
	"strconv"
)

type Input struct {
	UserId uint64 `json:"userid" query:"userid"  validate:"required,min=1"`
}

type Output struct {
	Code int        `json:"code"`
	Msg  string     `json:"msg"`
	Data outputData `json:"data"`
}

type outputData struct {
	CorpId              uint64 `json:"corpid"`              //企业ID
	CorpName            string `json:"corpName"`            //企业名称
	CorpStatus          int    `json:"corpStatus"`          //企业状态
	CorpStatusStr       string `json:"corpStatusStr"`       //企业状态解释-字符串
	CorpEnableStatusStr string `json:"corpEnableStatusStr"` //企业状态解释-可用状态字符串（启用|禁用）
	CorpVerifyStatusStr string `json:"corpVerifyStatusStr"` //企业状态解释-审核状态字符串（审核中|审核通过）
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
	if usercorps, err = t_corp_corporation_user.FindUserCorporations(c.Mysql(), input.UserId); err != nil {
		return c.RESULT_ERROR(ERR_INNER_ERROR, err.Error())
	}
	if len(usercorps) == 0 {
		return c.RESULT_ERROR(ERR_USER_HAVE_NO_CORPORATION, "user have no valid corporation")
	}

	corpids := []uint64{}
	corpidstr, _ := t_user_common_keyvalue.Get(c.Mysql(), input.UserId, createCorporation.SUIK_LAST_SELECTED_CORPID)
	corpidselected, _ := strconv.ParseUint(corpidstr, 10, 64)
	for _, uc := range usercorps {
		if corpidselected == uc.F_corporation_id {
			corpids = append(corpids, uc.F_corporation_id)
			break
		}
	}
	if len(corpids) == 0 {
		corpids = append(corpids, usercorps[0].F_corporation_id)
		corpidstr = strconv.FormatUint(usercorps[0].F_corporation_id, 10)
		t_user_common_keyvalue.Set(c.Mysql(), input.UserId, createCorporation.SUIK_LAST_SELECTED_CORPID, corpidstr)
	}

	corps := []t_corp_corporation.Corporation{}
	if corps, err = t_corp_corporation.FindCorporationsByIds(c.Mysql(), corpids); err != nil {
		return c.RESULT_ERROR(ERR_INNER_ERROR, err.Error())
	}
	if len(corps) != 1 {
		return c.RESULT_ERROR(ERR_INNER_ERROR, "UNREACHABLE CODE, Find no one or more than one corporation")
	}

	corp := corps[0]
	st := t_corp_corporation.Status(corp.F_status)

	//Step 5. set success
	output.Code = 0
	output.Msg = "success"
	output.Data.CorpId = corp.F_id
	output.Data.CorpName = corp.F_name
	output.Data.CorpStatus = corp.F_status
	output.Data.CorpStatusStr = st.String()
	output.Data.CorpEnableStatusStr = st.EnableString()
	output.Data.CorpVerifyStatusStr = st.VerifyString()

	return c.RESULT(output)
}
