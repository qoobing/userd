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
	"encoding/json"
	"github.com/labstack/echo"
	"github.com/qoobing/userd/src/common"
	. "github.com/qoobing/userd/src/const"
	"github.com/qoobing/userd/src/model"
	"github.com/qoobing/userd/src/model/t_user"
	"github.com/qoobing/userd/src/model/t_user_common_keyvalue"
	. "qoobing.com/utillib.golang/api"
	"qoobing.com/utillib.golang/log"
	"reflect"
	"strings"
)

type Input struct {
	UATK  string `json:"UATK"            validate:"omitempty,min=4"` //User Access ToKen
	SUIKA []SUIK `json:"SUIKA,omitempty" validate:"omitempty"`       //Standard User Information Key Array
}

type Output struct {
	Eno        int    `json:"eno"`
	Err        string `json:"err"`
	Name       string `json:"name"`            //名称
	Nickname   string `json:"nickname"`        //昵称
	Avatar     string `json:"avatar"`          //图像
	Loginstate int    `json:"loginstate"`      //登录等级状态
	SUIVM      *SUIVM `json:"SUIVM,omitempty"` //Standard User Information Value Map
}

type OutputV2 struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Name       string `json:"name"`            //名称
		Nickname   string `json:"nickname"`        //昵称
		Avatar     string `json:"avatar"`          //图像
		Loginstate int    `json:"loginstate"`      //登录等级状态
		SUIVM      *SUIVM `json:"SUIVM,omitempty"` //Standard User Information Key Array
	} `json:"data"`
}

type SUIK string

type SUIVM struct {
	Address    *string `json:"address,omitempty"    validate:"omitempty,min=4"` //地址
	Birthday   *int64  `json:"birthday,omitempty"   validate:"omitempty,min=4"` //生日
	CreateTime *int64  `json:"createtime,omitempty" validate:"omitempty,min=4"` //注册时间
	Mobile     *string `json:"mobile,omitempty"     validate:"omitempty,min=4"` //手机号
	MobileSafe *string `json:"mobilesafe,omitempty" validate:"omitempty,min=4"` //手机号
	Email      *string `json:"email,omitempty"      validate:"omitempty,min=4"` //电子邮箱
	EmailSafe  *string `json:"emailsafe,omitempty"  validate:"omitempty,min=4"` //电子邮箱
	QQ         *string `json:"qq,omitempty"         validate:"omitempty,min=4"` //QQ号
	WX         *string `json:"wx,omitempty"         validate:"omitempty,min=4"` //微信号
	DD         *string `json:"dd,omitempty"         validate:"omitempty,min=4"` //钉钉号
	Name       *string `json:"name,omitempty"       validate:"omitempty,min=4"` //名称
	Nickname   *string `json:"nickname,omitempty"   validate:"omitempty,min=4"` //昵称
	Avatar     *string `json:"avatar,omitempty"     validate:"omitempty,min=4"` //图像
	data       map[string]interface{}
}

func (suivm *SUIVM) MarshalJSON() ([]byte, error) {
	retmap := map[string]interface{}{}
	for k, v := range suivm.data {
		retmap[k] = v
	}

	rvalue := reflect.ValueOf(*suivm)
	rtype := reflect.TypeOf(*suivm)

	for i, n := 0, rvalue.NumField(); i < n; i++ {
		f := rvalue.Field(i)
		t := rtype.Field(i)
		if !f.IsNil() && t.Name != "data" {
			retmap[t.Name] = f.Interface()
		}
	}

	return json.Marshal(retmap)
}

const (
	SUIK_USER_DEFINED_KEY_PREFIX = "SUIK_"
	SUIK_ADDRESS                 = "address"
	SUIK_BIRTHDAY                = "birthday"
	SUIK_CREATETIME              = "createtime"
	SUIK_MOBILE                  = "mobile"
	SUIK_EMAIL                   = "email"
	SUIK_MOBILE_SAFE             = "mobilesafe"
	SUIK_EMAIL_SAFE              = "emailsafe"
	SUIK_QQ                      = "qq"
	SUIK_WX                      = "wx"
	SUIK_DD                      = "dd"
)

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
	//user, err := model.GetLoginInfo(c, c.Redis())
	user, err := model.GetUserAccessTokenData(c.Redis(), input.UATK)
	if err != nil {
		log.Debugf(err.Error())
		return c.RESULT_ERROR(ERR_NOT_LOGIN, "not login")
	}

	//Step 4. set output
	if ver == "1" {
		var output Output
		output.Eno = 0
		output.Err = "success"
		output.Name = user.Name
		output.Nickname = user.Nickname
		output.Avatar = user.Avatar
		output.Loginstate = user.Loginstate
		output.SUIVM = GetSUIVM(c, user.Userid, input.SUIKA)
		return c.RESULT(output)
	} else {
		var output OutputV2
		output.Code = 0
		output.Msg = "success"
		output.Data.Name = user.Name
		output.Data.Nickname = user.Nickname
		output.Data.Avatar = user.Avatar
		output.Data.Loginstate = user.Loginstate
		output.Data.SUIVM = GetSUIVM(c, user.Userid, input.SUIKA)
		return c.RESULT(output)
	}
}

func GetSUIVM(c ApiContext, userid uint64, suika []SUIK) (extensions *SUIVM) {
	user, err := t_user.GetUserByUserid(c.Mysql(), userid)
	if err != nil {
		return extensions
	}
	user_defined_keys := []string{}
	for _, k := range suika {
		if extensions == nil {
			extensions = &SUIVM{data: map[string]interface{}{}}
		}
		switch k {
		case SUIK_ADDRESS:
			address := user.F_address
			extensions.Address = &address
		case SUIK_BIRTHDAY:
			tampstamp := common.GetTimeStamp(user.F_birthday)
			extensions.Birthday = &tampstamp
		case SUIK_CREATETIME:
			tampstamp := common.GetTimeStamp(user.F_create_time)
			extensions.CreateTime = &tampstamp
		case SUIK_MOBILE:
			str := user.F_mobile
			extensions.Mobile = &str
		case SUIK_EMAIL:
			str := user.F_email
			extensions.Email = &str
		case SUIK_MOBILE_SAFE:
			str := common.FormatToSafeValue("MOBILE", user.F_mobile)
			extensions.MobileSafe = &str
		case SUIK_EMAIL_SAFE:
			str := common.FormatToSafeValue("EMAIL", user.F_email)
			extensions.EmailSafe = &str
		case SUIK_QQ:
			str := user.F_exid_qq
			extensions.QQ = &str
		case SUIK_WX:
			str := user.F_exid_wx
			extensions.WX = &str
		case SUIK_DD:
			str := user.F_exid_dd
			extensions.WX = &str
		default:
			if strings.HasPrefix(string(k), SUIK_USER_DEFINED_KEY_PREFIX) {
				user_defined_keys = append(user_defined_keys, string(k))
			}
		}
	}

	if len(user_defined_keys) > 0 {
		user_defined_values, _ := t_user_common_keyvalue.GetMulti(c.Mysql(), userid, user_defined_keys)
		for _, k := range user_defined_keys {
			if v, ok := user_defined_values[k]; ok {
				extensions.data[k] = v
			} else {
				extensions.data[k] = ""
			}
		}
	}

	return extensions
}
