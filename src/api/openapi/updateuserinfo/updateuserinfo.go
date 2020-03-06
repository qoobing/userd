/***********************************************************************
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php.
//******
// Filename:
// Description:
// Author:
// CreateTime:
/***********************************************************************/
package updateuserinfo

import (
	"github.com/labstack/echo"
	"github.com/qoobing/userd/src/api/openapi/getuserinfo"
	. "github.com/qoobing/userd/src/const"
	"github.com/qoobing/userd/src/model"
	"github.com/qoobing/userd/src/model/t_user"
	. "qoobing.com/utillib.golang/api"
	"qoobing.com/utillib.golang/log"
	"strings"
	"time"
)

type Input struct {
	UATK  string            `json:"UATK"       validate:"omitempty,min=4"` //User Access ToKen
	SUIVM getuserinfo.SUIVM `json:"SUIVM"      validate:"omitempty,min=4"` //Standard User Information Value Map
}

type Output struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func Main(cc echo.Context) error {
	//Step 1. init apicontext
	c := cc.(ApiContext)
	c.SetExAttribute(ATTRIBUTE_OUTPUT_FORMAT_CODE, "yes")
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

	//Step 3. find user by input UATK
	user, err := model.GetUserAccessTokenData(c.Redis(), input.UATK)
	if err != nil {
		log.Debugf(err.Error())
		return c.RESULT_ERROR(ERR_NOT_LOGIN, "not login")
	}

	output.Code = 0
	output.Msg = "success"
	//Step 4 set update fields
	var updatefields = map[string]interface{}{}
	if input.SUIVM.Name != nil {
		updatefields["F_name"] = *input.SUIVM.Name
	}
	if input.SUIVM.Nickname != nil {
		updatefields["F_nickname"] = *input.SUIVM.Nickname
	}
	if input.SUIVM.Avatar != nil {
		updatefields["F_avatar"] = *input.SUIVM.Avatar
	}
	if input.SUIVM.Address != nil {
		if addrarr := strings.Split(*input.SUIVM.Address, "|"); len(addrarr) != 4 {
			return c.RESULT_ERROR(ERR_INNER_ERROR, "Address error， must be 4 levels like '省|市|区|详细地址'")
		}
		updatefields["F_address"] = *input.SUIVM.Address
	}
	if input.SUIVM.Birthday != nil {
		birthday := time.Unix(*input.SUIVM.Birthday, 0).Format("2006-01-02 15:04:05")
		updatefields["F_birthday"] = birthday
	}
	if input.SUIVM.CreateTime != nil {
		log.Warningf("Can not modify 'createtime'")
	}
	if input.SUIVM.QQ != nil {
		updatefields["F_exid_qq"] = *input.SUIVM.QQ
	}
	if input.SUIVM.WX != nil {
		updatefields["F_exid_wx"] = *input.SUIVM.WX
	}
	if input.SUIVM.DD != nil {
		updatefields["F_exid_dd"] = *input.SUIVM.DD
	}
	log.PrintPreety("updatefields:", updatefields)
	if len(updatefields) == 0 {
		log.Debugf("updatefields empty, return directly")
		return c.RESULT(output)
	}

	//Step 5. update user
	tx := c.Mysql().Begin()
	txsuccess := false
	defer func() {
		if !txsuccess {
			tx.Rollback()
		}
	}()
	updateduser := t_user.User{}
	if err := tx.Model(&updateduser).Where("F_user_id = ?", user.Userid).Update(updatefields).Error; err != nil {
		log.Fatalf("update user error:%s", err.Error())
		return c.RESULT_ERROR(ERR_UPDATE_ERROR, "update user error")
	}
	updateduser, err = t_user.GetUserByUserid(tx, user.Userid)
	if err != nil {
		log.Fatalf(err.Error())
		return c.RESULT_ERROR(ERR_UPDATE_ERROR, "update user error(get user by id failed)")
	}
	err = model.UpdateUserAccessTokenData(c.Redis(), updateduser, input.UATK)
	if err != nil {
		log.Warningf("update REDIS uatk failed, please let user redo Login!!!!!!")
	}

	//Step 7. set output
	log.Debugf("done")
	return c.RESULT(output)
}
