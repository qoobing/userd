/***********************************************************************
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php.
//******
// Filename:
// Description:
// Author:
// CreateTime:
/***********************************************************************/
package createnewuser

import (
	"github.com/labstack/echo"
	"net/http"
	"github.com/qoobing/userd/src/api/openapi/login"
	. "github.com/qoobing/userd/src/const"
	"github.com/qoobing/userd/src/model"
	"github.com/qoobing/userd/src/model/t_login_scene"
	"github.com/qoobing/userd/src/model/t_user"
	. "qoobing.com/utillib.golang/api"
	"qoobing.com/utillib.golang/log"
	"strings"
	"time"
)

type Input struct {
	Scene    string `form:"scene"      json:"scene"      validate:"required,min=4,max=16"`
	Location string `form:"location"   json:"location"   validate:"omitempty,min=1,max=1024"`
	Name     string `form:"name"       json:"name"       validate:"required,min=4,max=64"`
	Nametype int    `form:"nametype"   json:"nametype"   validate:"omitempty,min=1,max=4"`
	Pass     string `form:"pass"       json:"pass"       validate:"omitempty,min=8,max=64"` //新增参数，替换理解不便的多参数结构
	Passtype int    `form:"passtype"   json:"passtype"   validate:"omitempty,min=1,max=4"`  //新增参数，替换理解不便的多参数结构
	Mobile   string `form:"mobile"     json:"mobile"     validate:"omitempty,min=8"`
	Email    string `form:"email"      json:"email"      validate:"omitempty,email"`
	Password string `form:"password"   json:"password"   validate:"omitempty,len=64"`
	Nickname string `form:"nickname"   json:"nickname"   validate:"omitempty,min=2"`
	Avatar   string `form:"avatar"     json:"avatar"     validate:"omitempty,url"`
}

type Output struct {
	Eno      int    `json:"eno"`
	Err      string `json:"err"`
	Location string `json:"location"`
}

func Main(cc echo.Context) error {
	//Step 1. init apicontext
	c := cc.(ApiContext)
	defer c.PANIC_RECOVER()
	c.Redis()
	c.Mysql()

	//Step 2. parameters initial
	var (
		input      Input
		output     Output
		vcodekey   string
		smsvcode   string
		emailvcode string
	)
	if err := c.BindInput(&input); err != nil {
		return c.RESULT_PARAMETER_ERROR(err.Error())
	}

	//Step 3. find user by input name
	user, err := t_user.FindUser(c.Mysql(), input.Name, NAMETYPE_ANY)
	if err != nil && err.Error() == USER_NOT_EXIST {
		log.Debugf("user not exist, continue create new user")
	} else if err != nil {
		return c.RESULT_ERROR(ERR_CREATE_USER_ERROR, err.Error())
	} else {
		return c.RESULT_ERROR(ERR_USER_ALREADY_EXIST, "user already exist")
	}
	//Step 3.1 verify user: initial type
	var verifytype = VERIFYTYPE_PASSWORD
	if input.Passtype != 0 {
		verifytype = input.Passtype
		switch verifytype {
		case VERIFYTYPE_PASSWORD:
			input.Password = input.Pass
		case VERIFYTYPE_MOBILE, VERIFYTYPE_EMAIL:
			arr := strings.Split(input.Pass, "#^_^#")
			if len(arr) != 2 {
				return c.RESULT_PARAMETER_ERROR("pass format error")
			}
			vcodekey = arr[0]
			if verifytype == VERIFYTYPE_MOBILE {
				smsvcode = arr[1]
			} else {
				emailvcode = arr[1]
			}
		default:

		}
	} else {
		return c.RESULT_PARAMETER_ERROR("Passtype Invalid")
	}

	//Step 3.2 verify user by input password or verify code
	c.Redis()
	switch verifytype {
	case VERIFYTYPE_MOBILE:
		if t_user.VerifyMobileVcode(input.Name, vcodekey, smsvcode, VCODEADDRESSTYPE_MOBILE_REGISTER) != true {
			return c.RESULT_ERROR(ERR_VERIFYCODE_INVLID, "register failed: verify code invalid")
		}
	case VERIFYTYPE_EMAIL:
		if t_user.VerifyEmailVcode(input.Name, vcodekey, emailvcode, VCODEADDRESSTYPE_EMAIL_REGISTER) != true {
			return c.RESULT_ERROR(ERR_VERIFYCODE_INVLID, "register failed: verify code invalid")
		}
	default:
		return c.RESULT_ERROR(ERR_CREATE_USER_ERROR, "register failed: unknown verify type")
	}

	//Step 4. create new user
	user.F_name = input.Name
	user.F_mobile = input.Mobile
	user.F_email = input.Email
	user.F_nickname = input.Nickname
	user.F_avatar = input.Avatar
	user.F_sec_password = input.Password
	if err := t_user.CreateUser(c.Mysql(), &user); err != nil {
		return c.RESULT_ERROR(ERR_CREATE_USER_ERROR, err.Error())
	}

	//Step 5. find use scene
	scene, err := t_login_scene.FindLoginScene(c.Mysql(), input.Scene)
	if err != nil {
		return c.RESULT_ERROR(ERR_LOGIN_SCENE_INVALID, err.Error())
	}

	//Step 6. write reids cache
	cookievalue, err := model.SetUserAccessTokenData(c.Redis(), user, scene)
	if err != nil {
		return c.RESULT_ERROR(ERR_LOGIN_FAILED, err.Error())
	}

	//Step 7. set cookie
	cookie := new(http.Cookie)
	cookie.Name = COOKIE_NAME_USERINFO
	cookie.Domain = COOKIE_DOMAIN_USERINFO
	cookie.Path = COOKIE_DEFAULT_PATH
	cookie.Value = cookievalue
	cookie.Expires = time.Now().Add(24 * 7 * time.Hour)
	c.SetCookie(cookie)

	//Step 8. set redirect location
	if reurl, err := login.RecreateRedirectUrl(scene, input.Location, user, cookie.Value); err != nil {
		log.Warningf("create redirect_url error: %s", err.Error())
		return c.RESULT_ERROR(ERR_LOGIN_FAILED, "redirect error")
	} else {
		output.Location = reurl
	}

	//Step 9. set success
	output.Eno = 0
	output.Err = "success"

	return c.RESULT(output)
}
