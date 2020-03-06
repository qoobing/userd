/***********************************************************************
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php.
//******
// Filename:
// Description:
// Author:
// CreateTime:
/***********************************************************************/
package getuserinfo

import (
	"github.com/labstack/echo"
	. "github.com/qoobing/userd/src/const"
	"github.com/qoobing/userd/src/model"
	. "qoobing.com/utillib.golang/api"
	"qoobing.com/utillib.golang/log"
	"qoobing.com/utillib.golang/xyz"
)

type Input struct {
	UATK string `json:"UATK" validate:"omitempty,min=4"`
}

type Output struct {
	Eno        int    `json:"eno"`
	Err        string `json:"err"`
	Name       string `json:"name"`
	Nickname   string `json:"nickname"`
	Userid     uint64 `json:"userid"`
	Mobile     string `json:"mobile"`
	Email      string `json:"email"`
	Avatar     string `json:"avatar"`
	Loginstate int    `json:"loginstate"`
}

type OutputV2 struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Name       string `json:"name"`
		Nickname   string `json:"nickname"`
		Userid     uint64 `json:"userid"`
		Mobile     string `json:"mobile"`
		Email      string `json:"email"`
		Avatar     string `json:"avatar"`
		Loginstate int    `json:"loginstate"`
	} `json:"data"`
}

func MainV1(cc echo.Context) error {
	c := cc.(ApiContext)
	return Main(c, "1")
}
func MainV2(cc echo.Context) error {
	c := cc.(ApiContext)
	c.SetExAttribute(ATTRIBUTE_OUTPUT_FORMAT_CODE, "yes")
	return Main(c, "2")
}

func Main(c ApiContext, ver string) error {
	//Step 1. init apicontext
	defer c.PANIC_RECOVER()
	c.Redis()
	c.Mysql()

	//Step 2. parameters initial
	var (
		input Input
	)
	if err := c.BindInput(&input); err != nil {
		return c.RESULT_PARAMETER_ERROR(err.Error())
	}

	//Step 3. get user info
	user, err := model.GetUserAccessTokenData(c.Redis(), input.UATK)
	if err != nil {
		log.Debugf(err.Error())
		return c.RESULT_ERROR(ERR_NOT_LOGIN, "not login")
	}

	if ver == "1" {
		var output Output
		//Step 4. set output
		output.Eno = 0
		output.Err = "success"
		output.Name = user.Name
		output.Nickname = user.Nickname
		output.Userid = user.Userid
		output.Avatar = user.Avatar
		output.Loginstate = user.Loginstate

		//Step 5. temp.  //TODO: xxxxxxx
		nametype := xyz.EstimateNametype(output.Name)
		switch nametype {
		case "MOBILE":
			output.Mobile = output.Name
			output.Email = ""
		case "EMAIL":
			output.Mobile = ""
			output.Email = output.Name
		case "NAME":
			output.Mobile = ""
			output.Email = ""
		}
		return c.RESULT(output)
	} else {
		var output OutputV2
		//Step 4. set output
		output.Code = 0
		output.Msg = "success"
		output.Data.Name = user.Name
		output.Data.Nickname = user.Nickname
		output.Data.Userid = user.Userid
		output.Data.Avatar = user.Avatar
		output.Data.Loginstate = user.Loginstate

		//Step 5. temp.  //TODO: xxxxxxx
		nametype := xyz.EstimateNametype(output.Data.Name)
		switch nametype {
		case "MOBILE":
			output.Data.Mobile = output.Data.Name
			output.Data.Email = ""
		case "EMAIL":
			output.Data.Mobile = ""
			output.Data.Email = output.Data.Name
		case "NAME":
			output.Data.Mobile = ""
			output.Data.Email = ""
		}
		return c.RESULT(output)
	}
}
