/***********************************************************************
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php.
//******
// Filename:
// Description:
// Author:
// CreateTime:
/***********************************************************************/
package setqrcodeuserinfo

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	. "github.com/qoobing/userd/src/config"
	. "github.com/qoobing/userd/src/const"
	"github.com/qoobing/userd/src/model"
	"github.com/qoobing/userd/src/model/t_user"
	"github.com/qoobing/userd/src/types"
	. "qoobing.com/utillib.golang/api"
	"qoobing.com/utillib.golang/gls"
	"qoobing.com/utillib.golang/log"
	"qoobing.com/utillib.golang/xyz"
	"strconv"
)

type Input struct {
	UATK     string `json:"UATK" validate:"omitempty,min=4"`
	TempCode string `json:"tempcode" validate:"required,min=4,max=64"`
}

type Output struct {
	Eno int    `json:"eno"`
	Err string `json:"err"`
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

	//Step 3. get user info
	user, err := model.GetUserAccessTokenData(c.Redis(), input.UATK)
	if err != nil {
		log.Debugf(err.Error())
		return c.RESULT_ERROR(ERR_NOT_LOGIN, "not login")
	}

	//Step 4. save redis
	err = saveQRCodeResult(input.TempCode, user)
	if err != nil {
		log.Debugf(err.Error())
		return c.RESULT_ERROR(ERR_SAVE_QRCODE_ERROR, "save qrcode failed:"+err.Error())
	}

	//Step 4. set output
	output.Eno = 0
	output.Err = "success"

	return c.RESULT(output)
}

func saveQRCodeResult(tempcode string, user types.Userinfo) (err error) {
	//Step 1. generate codekey and content
	var (
		code                         = xyz.GetRandomString(20)
		codekey                      = "QRCODE_CODE-" + code
		codevalue                    = user.Userid
		tempcodekey                  = "QRCODE_TEMPCODE-" + tempcode
		tempcodevalue                = code
		tempcodevalueold interface{} = nil
		expiretime                   = Config().QrCodeTimeout
	)
	if expiretime <= 0 {
		expiretime = 300
	}

	//Step 2. save to redis
	rds := gls.GetGlsValueNotNil("redis").(*RedisConn)
	//
	// Step 2.1 save tempcode to redis
	tempcodevalueold, err = rds.Do("GETSET", tempcodekey, tempcodevalue)
	if err != nil {
		log.Debugf("failed saved [%s], value:[%s]", tempcodekey, tempcodevalue)
		return err
	} else if tempcodevalueold == nil {
		log.Debugf("invalid or expired tempcode(old valie of key [%s] is nil)", tempcodekey)
		return fmt.Errorf("invalid or expired tempcode")
	}
	log.Debugf("success saved [%s], value:[%s]", tempcodekey, tempcodevalue)
	//
	// Step 2.2 save code to redis
	_, err = rds.Do("SETEX", codekey, expiretime, codevalue)
	if err != nil {
		log.Debugf("failed saved [%s], value:[%s]", codekey, codevalue)
		return err
	}
	log.Debugf("success saved [%s], value:[%s]", codekey, codevalue)

	return nil
}

func GetQRCodeResult(code string) (user t_user.User, err error) {
	var (
		codekey               = "QRCODE_CODE-" + code
		codevalue interface{} = nil
	)

	rds := gls.GetGlsValueNotNil("redis").(*RedisConn)
	codevalue, err = rds.Do("GET", codekey)
	if err != nil {
		log.Debugf("failed GET [%s]", codekey)
		return user, err
	} else if codevalue == nil {
		log.Debugf("invalid or expired code [%s]", codekey)
		return user, fmt.Errorf("invalid or expired code")
	}
	log.Debugf("success GET [%s], value is:[%v]", codekey, codevalue)

	db := gls.GetGlsValueNotNil("mysql").(*gorm.DB)
	userid, err := strconv.ParseUint(string(codevalue.([]byte)), 10, 64)
	if err != nil {
		log.Debugf("failed parse userid by [%s]", string(codevalue.([]byte)))
		return user, err
	}

	user, err = t_user.GetUserByUserid(db, userid)
	if err != nil {
		log.Debugf("failed GetUserByUserid(%d)", userid)
	} else {
		log.Debugf("success GetUserByUserid(%d), user:[%+v]", userid, user)
	}

	return user, err
}
