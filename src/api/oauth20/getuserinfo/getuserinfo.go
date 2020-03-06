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
	"fmt"
	"github.com/labstack/echo"
	. "github.com/qoobing/userd/src/const"
	"github.com/qoobing/userd/src/model"
	. "qoobing.com/utillib.golang/api"
	"qoobing.com/utillib.golang/log"
	"qoobing.com/utillib.golang/xyz"
	"strings"
)

type Input struct {
	UATK string `json:"uatk"   validate:"required,min=4,max=128"`
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

func Main(cc echo.Context) error {
	//Step 1. init apicontext
	c := cc.(ApiContext)
	defer c.PANIC_RECOVER()
	c.Redis()
	c.Mysql()

	//Step 2. parameters initial
	var (
		input  Input
		output Output
	)
	if err := thisApiBind(c, &input); err != nil {
		return c.RESULT_PARAMETER_ERROR(err.Error())
	}

	//Step 3. get user info
	user, err := model.GetUserAccessTokenData(c.Redis(), input.UATK)
	if err != nil {
		log.Debugf(err.Error())
		return c.RESULT_ERROR(ERR_NOT_LOGIN, "not login")
	}

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
}

func thisApiBind(c ApiContext, input *Input) error {
	baererauth := c.Request().Header.Get("authorization")
	if strings.HasPrefix(baererauth, "Bearer") {
		input.UATK = strings.Split(baererauth, " ")[1]
		return nil
	}
	return fmt.Errorf("Have No Header: [Authorization: Bearer xxxxxxxxxxxxxxxx]")
}

func CheckOAuth20Sign() {

}
