/***********************************************************************
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php.
//******
// Filename:
// Description:
// Author:
// CreateTime:
/***********************************************************************/
package getuserinfoaccesstoken

import (
	"github.com/labstack/echo"
	. "github.com/qoobing/userd/src/const"
	"github.com/qoobing/userd/src/model"
	. "qoobing.com/utillib.golang/api"
	"qoobing.com/utillib.golang/log"
)

type Input struct {
	Code     string `json:"code"       form:"code"      validate:"required,min=4,max=128"`
	ClientId string `json:"client_id"  form:"client_id" validate:"required,min=4,max=64"`
}

type Output struct {
	Eno           int    `json:"eno"`
	Err           string `json:"err"`
	Access_token  string `json:"access_token"`  //接口调用凭证
	Expires_in    int64  `json:"expires_in"`    //access_token接口调用凭证超时时间，单位（秒）
	Refresh_token string `json:"refresh_token"` //用户刷新access_token
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
	if err := c.BindInput(&input); err != nil {
		return c.RESULT_PARAMETER_ERROR(err.Error())
	}

	//Step 3. get ttl info
	ttl, err := model.GetUserAccessTokenTTL(c.Redis(), input.Code)
	if err != nil {
		log.Debugf(err.Error())
		return c.RESULT_ERROR(ERR_NOT_LOGIN, "not login")
	}

	//Step 4. set output
	output.Eno = 0
	output.Err = "success"
	output.Access_token = input.Code
	output.Expires_in = ttl
	output.Refresh_token = input.Code

	return c.RESULT(output)
}
