/***********************************************************************
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php.
//******
// Filename:
// Description:
// Author:
// CreateTime:
/***********************************************************************/
package updateuserpassword

import (
	"github.com/labstack/echo"
	. "github.com/qoobing/userd/src/const"
	"github.com/qoobing/userd/src/model/t_user"
	. "qoobing.com/utillib.golang/api"
	"qoobing.com/utillib.golang/log"
	"strings"
)

type Input struct {
	Name        string `form:"name"        json:"name"        validate:"required,min=4,max=64"`
	Nametype    int    `form:"nametype"    json:"nametype"    validate:"omitempty,min=1,max=4"`
	Pass        string `form:"pass"        json:"pass"        validate:"omitempty,min=8,max=64"`  //新增参数，替换理解不便的多参数结构
	Passtype    int    `form:"passtype"    json:"passtype"    validate:"omitempty,min=1,max=4"`   //新增参数，替换理解不便的多参数结构
	Update      string `form:"update"      json:"update"      validate:"omitempty,min=8,max=64"`  //新增参数，替换理解不便的多参数结构
	Updatefield int    `form:"updatefield" json:"updatefield" validate:"omitempty,min=1,max=4"`   //新增参数，替换理解不便的多参数结构
	Vcodekey    string `form:"vcodekey"    json:"vcodekey"    validate:"omitempty,min=10,max=80"` //准备废弃，由Pass&Passtype结构代替
	Smsvcode    string `form:"smsvcode"    json:"smsvcode"    validate:"omitempty,len=6"`         //准备废弃，由Pass&Passtype结构代替
	Emailvcode  string `form:"emailvcode"  json:"emailvcode"  validate:"omitempty,len=6"`         //准备废弃，由Pass&Passtype结构代替
	Loginvcode  string `form:"loginvcode"  json:"loginvcode"  validate:"omitempty,len=6"`         //准备废弃，由Pass&Passtype结构代替
	Newpassword string `form:"newpassword" json:"newpassword" validate:"omitempty,min=4,max=80"`  //准备废弃，由Update&Updatetype结构代替
}

type Output struct {
	Eno int    `json:"eno"`
	Err string `json:"err"`
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

func Main(c ApiContext, outputversion string) error {
	//Step 1. init apicontext
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
	if input.Pass == "" && input.Passtype == 0 &&
		input.Vcodekey == "" && input.Smsvcode == "" && input.Emailvcode == "" &&
		input.Loginvcode == "" {
		return c.RESULT_PARAMETER_ERROR("you should input pass & passtype")
	}

	//Step 3. find user by input name
	user, err := t_user.FindUser(c.Mysql(), input.Name, input.Nametype)
	if err != nil && err.Error() == USER_NOT_EXIST {
		return c.RESULT_ERROR(ERR_USER_NOT_EXIST, err.Error())
	} else if err != nil {
		return c.RESULT_ERROR(ERR_FIND_USER_ERROR, "find user error")
	}
	log.Debugf("found user: userid=%d, username=%s", user.F_user_id, user.F_name)

	//Step 4 verify user: initial type
	var verifytype = VERIFYTYPE_PASSWORD
	var vcodeaddresstype = VCODEADDRESSTYPE_MOBILE_LOGIN
	if input.Passtype != 0 {
		verifytype = input.Passtype
		switch verifytype {
		case VERIFYTYPE_PASSWORD:

		case VERIFYTYPE_MOBILE, VERIFYTYPE_EMAIL:
			arr := strings.Split(input.Pass, "#^_^#")
			if len(arr) != 2 {
				return c.RESULT_PARAMETER_ERROR("pass format error")
			}
			input.Vcodekey = arr[0]
			if verifytype == VERIFYTYPE_MOBILE {
				vcodeaddresstype = VCODEADDRESSTYPE_SAFE_MOBILE_PASSWORD
				input.Smsvcode = arr[1]
			} else {
				vcodeaddresstype = VCODEADDRESSTYPE_SAFE_EMAIL_PASSWORD
				input.Emailvcode = arr[1]
			}
		}
	} else {
		return c.RESULT_ERROR(ERR_OUT_OF_DATE_API, "out of date usage, see more: 'https://github.com/qoobing/userd'")
	}

	//Step 4.1 verify user: verify by input password or verify code
	switch verifytype {
	case VERIFYTYPE_PASSWORD:
		if t_user.VerifyUserPassword(user, input.Pass) != true {
			return c.RESULT_ERROR(ERR_PASSWORD_INVLID, "update failed: username or password invalid")
		}
	case VERIFYTYPE_MOBILE:
		if t_user.VerifyMobileVcode(input.Name, input.Vcodekey, input.Smsvcode, vcodeaddresstype) != true {
			return c.RESULT_ERROR(ERR_VERIFYCODE_INVLID, "update failed: verify code invalid")
		}
	case VERIFYTYPE_EMAIL:
		if t_user.VerifyEmailVcode(input.Name, input.Vcodekey, input.Emailvcode, vcodeaddresstype) != true {
			return c.RESULT_ERROR(ERR_VERIFYCODE_INVLID, "update failed: verify code invalid")
		}
	default:
		return c.RESULT_ERROR(ERR_VERIFYCODE_INVLID, "update failed: invalid parameter")
	}

	//Step 5. update
	switch input.Updatefield {
	case UPDATE_PASSWORD_FIELD_CHECK_PERMISSION:
		output.Eno = 0
		output.Err = "success"
		log.Debugf("done check permission")
		if outputversion == "2" {
			outputv2 := OutputV2{
				Code: output.Eno,
				Msg:  output.Err,
			}
			return c.RESULT(outputv2)
		} else {
			return c.RESULT(output)
		}
	case UPDATE_PASSWORD_FIELD_PASSWORD:
		//Step 5.1 update user password
		user.F_sec_password = input.Update
		err = t_user.UpdateUserPassword(c.Mysql(), user)
		if err != nil {
			return c.RESULT_ERROR(ERR_UPDATE_ERROR, "update password failed")
		}
	case UPDATE_PASSWORD_FIELD_MOBILE:
		//Step 5.2 update user mobile
		arr := strings.Split(input.Update, "#^_^#")
		if len(arr) != 3 {
			return c.RESULT_PARAMETER_ERROR("update value format error")
		}
		var (
			newmobile      = arr[0]
			newmobilevkey  = arr[1]
			newmobilevcode = arr[2]
		)
		if _, e := t_user.FindUser(c.Mysql(), newmobile, NAMETYPE_MOBILE); e != nil && e != t_user.ERROR_USER_NOT_EXIST {
			return c.RESULT_ERROR(ERR_UPDATE_ERROR, "find user by mobile '"+input.Update+"' error")
		} else if e == nil {
			return c.RESULT_ERROR(ERR_UPDATE_MOBILE_TO_ALREADY_USED, "can not bind to already used mobile:"+input.Update)
		}
		if t_user.VerifyMobileVcode(newmobile, newmobilevkey, newmobilevcode, VCODEADDRESSTYPE_MOBILE_PASSWORD) != true {
			return c.RESULT_ERROR(ERR_UPDATE_ERROR, "update failed: verify code invalid")
		}
		user.F_mobile = newmobile
		err = t_user.UpdateUserMobile(c.Mysql(), user)
		if err != nil {
			return c.RESULT_ERROR(ERR_UPDATE_ERROR, "update mobile failed")
		}
	case UPDATE_PASSWORD_FIELD_MOBILE_UNBIND:
		user.F_mobile = ""
		err = t_user.UpdateUserMobile(c.Mysql(), user)
		if err != nil {
			return c.RESULT_ERROR(ERR_UPDATE_ERROR, "update mobile failed")
		}
	case UPDATE_PASSWORD_FIELD_EMAIL:
		//Step 5.3 update user email
		arr := strings.Split(input.Update, "#^_^#")
		if len(arr) != 3 {
			return c.RESULT_PARAMETER_ERROR("update value format error")
		}
		var (
			newemail      = arr[0]
			newemailvkey  = arr[1]
			newemailvcode = arr[2]
		)
		if _, e := t_user.FindUser(c.Mysql(), newemail, NAMETYPE_EMAIL); e != nil && e != t_user.ERROR_USER_NOT_EXIST {
			return c.RESULT_ERROR(ERR_UPDATE_ERROR, "find user by mobile '"+input.Update+"' error")
		} else if e == nil {
			return c.RESULT_ERROR(ERR_UPDATE_EMAIL_TO_ALREADY_USED, "can not bind to already used email:"+input.Update)
		}
		if t_user.VerifyEmailVcode(newemail, newemailvkey, newemailvcode, VCODEADDRESSTYPE_EMAIL_PASSWORD) != true {
			return c.RESULT_ERROR(ERR_UPDATE_ERROR, "update failed: verify code invalid")
		}
		user.F_email = newemail
		err = t_user.UpdateUserEmail(c.Mysql(), user)
		if err != nil {
			return c.RESULT_ERROR(ERR_UPDATE_ERROR, "update email failed")
		}
	case UPDATE_PASSWORD_FIELD_EMAIL_UNBIND:
		user.F_email = ""
		err = t_user.UpdateUserEmail(c.Mysql(), user)
		if err != nil {
			return c.RESULT_ERROR(ERR_UPDATE_ERROR, "update email failed")
		}
	default:
		return c.RESULT_PARAMETER_ERROR("updatefield invalid")
	}

	//Step 7. set output
	output.Eno = 0
	output.Err = "success"
	if outputversion == "2" {
		outputv2 := OutputV2{
			Code: output.Eno,
			Msg:  output.Err,
		}
		return c.RESULT(outputv2)
	} else {
		return c.RESULT(output)
	}
}
