/***********************************************************************
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php.
//******
// Filename:
// Description:
// Author:
// CreateTime:
/***********************************************************************/
package getverifycode

import (
	"encoding/json"
	"errors"
	"github.com/labstack/echo"
	. "github.com/qoobing/userd/src/config"
	. "github.com/qoobing/userd/src/const"
	"github.com/qoobing/userd/src/model"
	"github.com/qoobing/userd/src/model/t_user"
	. "qoobing.com/utillib.golang/api"
	"qoobing.com/utillib.golang/gls"
	"qoobing.com/utillib.golang/log"
	"qoobing.com/utillib.golang/rpc"
	. "qoobing.com/utillib.golang/xyz"

	"github.com/qoobing/userd/src/api/openapi/getcaptchaimage"
	"github.com/qoobing/userd/src/types"
	"regexp"
	"strconv"
	"strings"
	"yqtc.com/ubox.golib/xyz"
)

type Input struct {
	UATK       string `json:"UATK"        validate:"omitempty,min=4"` //User Access ToKen
	Type       int    `json:"type"        validate:"required,min=1,max=8"`
	Address    string `json:"address"     validate:"required,min=4,max=256"`
	CaptchaKey string `json:"captchakey"`
	Captcha    string `json:"captcha"`
}

type InputUser struct {
}

type Output struct {
	Eno      int    `json:"eno"`
	Err      string `json:"err"`
	Vcodekey string `json:"vcodekey"`
}

type OutputV2 struct {
	Code     int    `json:"code"`
	Msg      string `json:"msg"`
	Vcodekey string `json:"vcodekey"`
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
	c.Save()

	//Step 2. parameters initial
	var (
		input  Input
		output Output
	)
	if err := c.BindInput(&input); err != nil {
		return c.RESULT_PARAMETER_ERROR(err.Error())
	}
	//TODO: check need temp token or not
	if input.Captcha == "" {
		log.Debugf("DONOT need to verify captcha")
	} else if !getcaptchaimage.VerifyCaptcha(input.CaptchaKey, input.Captcha, "NOT_USED") {
		log.Debugf("VerifyCaptcha failed")
		return c.RESULT_ERROR(ERR_GETVCODE_CAPTCHA_VERIFY_FAILED, "captcha verify failed")
	}
	if input.Type == VCODEADDRESSTYPE_SAFE_MOBILE_PASSWORD || input.Type == VCODEADDRESSTYPE_SAFE_EMAIL_PASSWORD {
		if input.UATK == "" {
			return c.RESULT_ERROR(ERR_NOT_LOGIN, "not login when get vcode by safe_mobile/safe_email")
		}
		u, err := model.GetUserAccessTokenData(c.Redis(), input.UATK)
		if err != nil {
			log.Debugf(err.Error())
			return c.RESULT_ERROR(ERR_NOT_LOGIN, "not login when get vcode by safe_mobile/safe_email")
		}
		if user, e := t_user.GetUserByUserid(c.Mysql(), u.Userid); e != nil {
			return c.RESULT_ERROR(ERR_INNER_ERROR, "find user error")
		} else if input.Type == VCODEADDRESSTYPE_SAFE_MOBILE_PASSWORD {
			orgAddress := strings.Replace(input.Address, "*", ".*", -1)

			if reg, err := regexp.Compile(orgAddress); err != nil {
				return c.RESULT_PARAMETER_ERROR("invalid input address:" + input.Address)
			} else if reg.MatchString(user.F_mobile) {
				input.Address = user.F_mobile
			} else {
				return c.RESULT_ERROR(ERR_GETVCODE_ERROR, "input address not match")
			}
		} else if input.Type == VCODEADDRESSTYPE_SAFE_EMAIL_PASSWORD {
			orgAddress := strings.Replace(input.Address, "*", ".*", -1)

			if reg, err := regexp.Compile(orgAddress); err != nil {
				return c.RESULT_PARAMETER_ERROR("invalid input address:" + input.Address)
			} else if reg.MatchString(user.F_email) {
				input.Address = user.F_email
			} else {
				return c.RESULT_ERROR(ERR_GETVCODE_ERROR, "input address not match")
			}
		} else {
			log.Fatalf("UNREACHABLE CODE")
		}
	}

	//Step 3. generate verify code by estimated name type [mobile or email]
	var vcodekey string
	var err error
	var intype = model.EstimateNametype(input.Address)
	switch input.Type {
	case VCODEADDRESSTYPE_MOBILE_LOGIN,
		VCODEADDRESSTYPE_MOBILE_REGISTER,
		VCODEADDRESSTYPE_MOBILE_PASSWORD,
		VCODEADDRESSTYPE_SAFE_MOBILE_PASSWORD:
		ASSERT(intype == NAMETYPE_MOBILE, "address invalid: not mobile address")
		vcodekey, err = generateMobileVcode(input.Address, input.Type)
	case VCODEADDRESSTYPE_EMAIL_LOGIN,
		VCODEADDRESSTYPE_EMAIL_REGISTER,
		VCODEADDRESSTYPE_EMAIL_PASSWORD,
		VCODEADDRESSTYPE_SAFE_EMAIL_PASSWORD:
		ASSERT(intype == NAMETYPE_EMAIL, "address invalid: not email address")
		vcodekey, err = generateEmailVcode(input.Address, input.Type)
	default:
		panic("generate verify code: unknown name-type")
	}
	if err != nil {
		log.Debugf(err.Error())
		return c.RESULT_ERROR(ERR_GETVCODE_ERROR, "get verify code failed")
	}

	//Step 4. set output
	output.Eno = 0
	output.Err = "success"
	output.Vcodekey = vcodekey

	if outputversion == "2" {
		outputv2 := OutputV2{
			Code:     output.Eno,
			Msg:      output.Err,
			Vcodekey: output.Vcodekey,
		}
		return c.RESULT(outputv2)
	} else {
		return c.RESULT(output)
	}
}

func generateMobileVcode(phonenum string, vcodetype int) (vcodekey string, err error) {
	//Step 1. generate vcodekey and vcode
	var verifycode = types.VerifyCodeInfo{
		Name:      phonenum,
		Nametype:  NAMETYPE_MOBILE,
		VcodeType: vcodetype,
		Vcodekey:  GetRandomString(32),
		Vcode:     xyz.IF(phonenum == "18888888888", string("888888"), GetRandomString(6)).(string),
	}
	vcodekey = verifycode.Vcodekey
	value, _ := json.Marshal(verifycode)

	//Step 2. save vcodekey and vcode to redis
	rds := gls.GetGlsValueNotNil("redis").(*RedisConn)
	rdskey := "USER_MOBILE_VCODE-" + t_user.GetVcodeTypeName(vcodetype) + "-" + vcodekey
	_, err = rds.Do("SETEX", rdskey, Config().VCodeTimeout, value)
	if err != nil {
		log.Debugf("failed saved [%s], value:[%s]", rdskey, value)
		return verifycode.Vcodekey, err
	}
	log.Debugf("success saved [%s], value:[%s]", rdskey, value)

	//Step 3.1 send sms : initial input
	var (
		ret      = rpc.RpcResult{}
		eno      = 0
		smsinput = map[string]string{
			"PhoneNumber":   phonenum,
			"TemplateCode":  Config().Communicator.SmsTemplateCodeRegister,
			"TemplateParam": "{\"code\":\"" + verifycode.Vcode + "\"}",
		}
	)
	svrname := Config().Communicator.ServiceName
	cmdpath := Config().Communicator.CmdSendSms

	//Step 3.2 send sms : send request & handle result
	if phonenum == "18888888888" {
		log.Warningf("FAKE phonenum, trade as success. vcode=-----[%s]------", verifycode.Vcode)
	} else if strings.HasPrefix(phonenum, "19595") {
		log.Warningf("FAKE phonenum[%s], trade as success. vcode=-----[%s]------", phonenum, verifycode.Vcode)
	} else if ret, err = rpc.RpcSjsCall(svrname, "POST", cmdpath, smsinput); err != nil {
		return "", errors.New("send sms error: rpc call error = " + err.Error())
	} else if eno, err = ret.Get("eno").Int(); err != nil {
		return "", errors.New("send sms error: get 'Eno' error = " + err.Error())
	} else if eno != 0 {
		return "", errors.New("send sms error:" + strconv.Itoa(eno))
	}

	return vcodekey, nil
}

func generateEmailVcode(emailaddr string, vcodetype int) (vcodekey string, err error) {
	//Step 1. generate vcodekey and vcode
	var verifycode = types.VerifyCodeInfo{
		Name:      emailaddr,
		Nametype:  NAMETYPE_EMAIL,
		VcodeType: vcodetype,
		Vcodekey:  GetRandomString(32),
		Vcode:     xyz.IF(emailaddr == "8888@qq.com", string("888888"), GetRandomString(6)).(string),
	}
	vcodekey = verifycode.Vcodekey
	value, _ := json.Marshal(verifycode)

	//Step 2. save vcodekey and vcode to redis
	rds := gls.GetGlsValueNotNil("redis").(*RedisConn)
	rdskey := "USER_EMAIL_VCODE-" + t_user.GetVcodeTypeName(vcodetype) + "-" + vcodekey
	_, err = rds.Do("SETEX", rdskey, Config().VCodeTimeout, value)
	if err != nil {
		log.Debugf("failed saved [%s], value:[%s]", rdskey, value)
		return verifycode.Vcodekey, err
	}
	log.Debugf("success saved [%s], value:[%s]", rdskey, value)

	//Step 3.1 send sms : initial input
	var (
		ret      = rpc.RpcResult{}
		eno      = 0
		smsinput = map[string]string{
			"EmailAddr":     emailaddr,
			"TemplateCode":  Config().Communicator.EmailTemplateCodeRegister,
			"TemplateParam": "{\"code\":\"" + verifycode.Vcode + "\"}",
		}
	)
	svrname := Config().Communicator.ServiceName
	cmdpath := Config().Communicator.CmdSendEmail

	//Step 3.2 send sms : send request & handle result
	if emailaddr == "8888@qq.com" {
		log.Warningf("FAKE emailaddress, trade as success. vcode=-----[%s]------", verifycode.Vcode)
	} else if strings.HasSuffix(emailaddr, "@888.com") {
		log.Warningf("FAKE emailaddress[%s], trade as success. vcode=-----[%s]------", emailaddr, verifycode.Vcode)
	} else if ret, err = rpc.RpcSjsCall(svrname, "POST", cmdpath, smsinput); err != nil {
		return "", errors.New("send email error: rpc call error = " + err.Error())
	} else if eno, err = ret.Get("eno").Int(); err != nil {
		return "", errors.New("send email error: get 'Eno' error = " + err.Error())
	} else if eno != 0 {
		return "", errors.New("send email error:" + strconv.Itoa(eno))
	}

	return vcodekey, nil
}
