/***********************************************************************
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php.
//******
// Filename:
// Description:
// Author:
// CreateTime:
/***********************************************************************/
package getqrcoderesultcode

import (
	"github.com/labstack/echo"
	. "github.com/qoobing/userd/src/const"
	. "qoobing.com/utillib.golang/api"
	"qoobing.com/utillib.golang/gls"
	"qoobing.com/utillib.golang/log"

	"fmt"
	"reflect"
)

type Input struct {
	TempCode string `json:"tempcode" validate:"required,min=4,max=64"`
}

type Output struct {
	Eno   int    `json:"eno"`
	Err   string `json:"err"`
	Code  string `json:"code"`
	State string `json:"state"`
}

func Main(cc echo.Context) error {
	//Step 1. init apicontext
	c := cc.(ApiContext)
	defer c.PANIC_RECOVER()
	c.Redis()

	//Step 2. parameters initial
	var (
		input  Input
		output Output
	)
	if err := c.BindInput(&input); err != nil {
		return c.RESULT_PARAMETER_ERROR(err.Error())
	}

	//Step 3. generate verify code by estimated name type [mobile or email]
	var (
		code, err = checkQRCodeContent(input.TempCode)
	)
	if err == nil {
		log.Debugf("success full get code:[%s]", code)
	} else if err.Error() == "ERR_TEMPCODE_WAIT_COMMIT" {
		return c.RESULT_ERROR(ERR_TEMPCODE_WAIT_COMMIT, "get verify code failed")
	} else if err.Error() == "ERR_TEMPCODE_INVALID_OR_EXPIRED" {
		return c.RESULT_ERROR(ERR_TEMPCODE_INVALID_OR_EXPIRED, "tempcode invalid or expired")
	} else if err.Error() == "ERR_TEMPCODE_INVALID" {
		return c.RESULT_ERROR(ERR_TEMPCODE_INVALID, "tempcode invalid")
	} else {
		log.Debugf(err.Error())
		return c.RESULT_ERROR(ERR_CHECK_TEMPCODE_FAILED, "check qrcode failed")
	}

	//Step 4. set output
	output.Eno = 0
	output.Err = "success"
	output.Code = code

	return c.RESULT(output)
}

const CODEVALE_SEPERATOR = "-=+=-"

func checkQRCodeContent(tempcode string) (code string, err error) {
	//Step 1. generate codekey and content
	var (
		tempcodekey               = "QRCODE_TEMPCODE-" + tempcode
		tempcodevalue interface{} = nil
	)

	//Step 2. save vcodekey and vcode to redis
	rds := gls.GetGlsValueNotNil("redis").(*RedisConn)
	tempcodevalue, err = rds.Do("GET", tempcodekey)
	if err != nil {
		log.Debugf("failed get [%s]", tempcodekey)
		return "", err
	} else if tempcodevalue == nil {
		log.Debugf("tempcode is invalid or expired, key [%s] is nil", tempcodekey)
		return "", fmt.Errorf("ERR_TEMPCODE_INVALID_OR_EXPIRED")
	} else if codebyte, ok := tempcodevalue.([]byte); !ok {
		log.Warningf("tempcodevalue [value:%v] is NOT string, is %s", tempcodevalue, reflect.TypeOf(tempcodevalue).String())
		return "", fmt.Errorf("ERR_TEMPCODE_INVALID")
	} else if string(codebyte) == "WAIT_COMMIT" {
		log.Debugf("WAIT_COMMIT: value of key [%s] is [%s]", tempcodekey, tempcodevalue)
		return "", fmt.Errorf("ERR_TEMPCODE_WAIT_COMMIT")
	} else {
		code = string(codebyte)
	}

	log.Debugf("success get [%s], value is:[%s]", tempcodekey, tempcodevalue)
	return code, nil
}
