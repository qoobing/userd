/***********************************************************************
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php.
//******
// Filename:
// Description:
// Author:
// CreateTime:
/***********************************************************************/
package setUserInfoSelfDefined

import (
	"github.com/labstack/echo"
	"github.com/qoobing/userd/src/model/t_user_common_keyvalue"
	. "qoobing.com/utillib.golang/api"
)

type Input struct {
	USER  UInfo  `json:"-"`
	UATK  string `json:"UATK"            validate:"omitempty,min=4"` //User Access ToKen
	Key   string `json:"key"`                                        //键
	Value string `json:"value"`                                      //值
}

type Output struct {
	Eno   int    `json:"eno"`
	Err   string `json:"err"`
	Key   string `json:"key"`   //键
	Value string `json:"value"` //值
}

type OutputV2 struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
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

	//Step 3. set
	t_user_common_keyvalue.Set(c.Mysql(), input.USER.Userid, input.Key, input.Value)

	//Step 4. set output
	if ver == "1" {
		var output Output
		output.Eno = 0
		output.Err = "success"
		return c.RESULT(output)
	} else {
		var output OutputV2
		output.Code = 0
		output.Msg = "success"
		return c.RESULT(output)
	}
}
