/***********************************************************************
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php.
//******
// Filename:
// Description:
// Author:
// CreateTime:
/***********************************************************************/
package login

import (
	"encoding/json"
	"fmt"
	"github.com/labstack/echo"
	"net/http"
	"net/url"
	"github.com/qoobing/userd/src/api/rbac/basic/checkPermission"
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
	Scene      string `form:"scene"      json:"scene"      validate:"required,min=4,max=16"`
	Location   string `form:"location"   json:"location"   validate:"omitempty,min=1,max=1024"`
	Name       string `form:"name"       json:"name"       validate:"required,min=4,max=64"`
	Nametype   int    `form:"nametype"   json:"nametype"   validate:"omitempty,min=1,max=4"`
	Pass       string `form:"pass"       json:"pass"       validate:"omitempty,min=8,max=64"`  //新增参数，替换理解不便的多参数结构
	Passtype   int    `form:"passtype"   json:"passtype"   validate:"omitempty,min=1,max=4"`   //新增参数，替换理解不便的多参数结构
	Vcodekey   string `form:"vcodekey"   json:"vcodekey"   validate:"omitempty,min=10,max=80"` //准备废弃，由Pass&Passtype结构代替
	Password   string `form:"password"   json:"password"   validate:"omitempty,len=64"`        //准备废弃，由Pass&Passtype结构代替
	Smsvcode   string `form:"smsvcode"   json:"smsvcode"   validate:"omitempty,len=6"`         //准备废弃，由Pass&Passtype结构代替
	Emailvcode string `form:"emailvcode" json:"emailvcode" validate:"omitempty,len=6"`         //准备废弃，由Pass&Passtype结构代替
	Uatkcookie string `form:"uatkcookie" json:"uatkcookie" validate:"omitempty,min=10"`        //准备废弃，由Pass&Passtype结构代替
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
		input  Input
		output Output
	)
	if err := c.BindInput(&input); err != nil {
		return c.RESULT_PARAMETER_ERROR(err.Error())
	}
	if input.Pass == "" && input.Passtype == 0 &&
		input.Vcodekey == "" && input.Smsvcode == "" && input.Emailvcode == "" &&
		input.Password == "" && input.Uatkcookie == "" {
		return c.RESULT_PARAMETER_ERROR("you should input pass & passtype")
	}

	//Step 3.1 find scene
	scene, err := t_login_scene.FindLoginScene(c.Mysql(), input.Scene)
	if err != nil {
		return c.RESULT_ERROR(ERR_LOGIN_SCENE_INVALID, err.Error())
	}
	//Step 3.2 find user by input name
	user, err := t_user.FindUser(c.Mysql(), input.Name, input.Nametype)
	if err != nil && err.Error() == USER_NOT_EXIST {
		nametype := input.Nametype
		if nametype == NAMETYPE_ANY {
			nametype = model.EstimateNametype(input.Name)
		}
		autocreate := scene.GetAttribute(t_login_scene.ATTRIBUTE_NAME_AUTO_REGISTER_MOBILE_LOGIN, true).(bool)
		if autocreate && nametype == NAMETYPE_MOBILE {
			//Step 4.2 register new user to database
			newuser := t_user.User{
				F_name:   input.Name,
				F_mobile: input.Name,
			}
			err = t_user.CreateUser(c.Mysql(), &newuser)
			if err != nil {
				return c.RESULT_ERROR(ERR_CREATE_USER_ERROR, err.Error())
			}
			user, err = t_user.FindUser(c.Mysql(), input.Name, input.Nametype)
			if err != nil {
				log.Fatalf("Create a new user, but can not find it!!!!!!")
				return c.RESULT_ERROR(ERR_LOGIN_FAILED, err.Error())
			}
		} else {
			return c.RESULT_ERROR(ERR_USER_NOT_EXIST, err.Error())
		}
	} else if err != nil {
		return c.RESULT_ERROR(ERR_LOGIN_FAILED, err.Error())
	}

	//Step 4 verify user: initial type
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
			input.Password = input.Pass
			input.Vcodekey = arr[0]
			if verifytype == VERIFYTYPE_MOBILE {
				input.Smsvcode = arr[1]
			} else {
				input.Emailvcode = arr[1]
			}
		case VERIFYTYPE_COOKIE:
			input.Uatkcookie = input.Pass
		}
	} else {
		if input.Password != "" {
			verifytype = VERIFYTYPE_PASSWORD
		} else if input.Smsvcode != "" {
			verifytype = VERIFYTYPE_MOBILE
		} else if input.Emailvcode != "" {
			verifytype = VERIFYTYPE_EMAIL
		} else if input.Uatkcookie != "" {
			verifytype = VERIFYTYPE_COOKIE
		}
	}

	//Step 4.1 verify user: verify by input password or verify code
	switch verifytype {
	case VERIFYTYPE_PASSWORD:
		if t_user.VerifyUserPassword(user, input.Password) != true {
			return c.RESULT_ERROR(ERR_PASSWORD_INVLID, "login failed: username or password invalid")
		}
	case VERIFYTYPE_MOBILE:
		if t_user.VerifyMobileVcode(input.Name, input.Vcodekey, input.Smsvcode, VCODEADDRESSTYPE_MOBILE_LOGIN) != true {
			return c.RESULT_ERROR(ERR_VERIFYCODE_INVLID, "login failed: verify code invalid")
		}
	case VERIFYTYPE_EMAIL:
		if t_user.VerifyEmailVcode(input.Name, input.Vcodekey, input.Emailvcode, VCODEADDRESSTYPE_EMAIL_LOGIN) != true {
			return c.RESULT_ERROR(ERR_VERIFYCODE_INVLID, "login failed: verify code invalid")
		}
	case VERIFYTYPE_COOKIE:
		strfn := log.TRACE_INTO("get uatk data")
		defer log.TRACE_EXIT(strfn, "get uatk data")
		if t_user.VerifyUatkCookie(user, input.Uatkcookie) != true {
			return c.RESULT_ERROR(ERR_VERIFYCODE_INVLID, "login failed: uatk invalid")
		}
	default:
		return c.RESULT_ERROR(ERR_VERIFYCODE_INVLID, "login failed: invalid parameter")
	}

	//Step 5. find user scene
	if ok, err := checkPermission.CheckPrivilege(scene.F_privilege, user, nil); err != nil {
		log.Fatalf("CheckScenePrivilege error %s", err.Error())
		return c.RESULT_ERROR(ERR_SCENE_PERMISSION_DENIED, "CheckScenePrivilege error:"+err.Error())
	} else if !ok {
		log.Debugf("CheckScenePrivilege return false, permission deny")
		return c.RESULT_ERROR(ERR_LOGIN_FAILED, "permission deny")
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
	if reurl, err := RecreateRedirectUrl(scene, input.Location, user, cookie.Value); err != nil {
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

/**
 * 重构重定向地址
 * scene:     场景
 * location:  重定向地址或重定向地址中希望回传的参数
 * user:      用户实例
 * uatk:      用户授权token
 */
func RecreateRedirectUrl(scene t_login_scene.LoginScene, location string, user t_user.User, uatk string) (reurl string, err error) {
	//// 数据库设置的重定向向参数
	orgurl := scene.F_redirect_url
	u, err := url.Parse(orgurl)
	if err != nil {
		panic("redirect url:'" + orgurl + "' parse error[" + err.Error() + "]")
	}
	reurl = u.Scheme + "://" + u.Host + u.Path
	fregment := u.Fragment
	paramsvalues := url.Values{}

	//// 用户传入的重定向参数
	if strings.HasPrefix(location, "http://") || strings.HasPrefix(location, "https://") {
		//传入参数中有http或者https前缀，
		// 则认为其是完整的重定向地址，按照参数传入地址重定向
		uu, err := url.Parse(location)
		if err != nil {
			panic("redirect url:'" + orgurl + "' parse error[" + err.Error() + "]")
		}
		if strings.HasPrefix(uu.Host, "localhost") {
			log.Warningf("is localhost, DO'NOT verify redirect setting in t_scene table")
			reurl = uu.Scheme + "://" + uu.Host + uu.Path
			fregment = uu.Fragment
		} else if uu.Host != u.Host || !strings.HasPrefix(uu.Path, u.Path) {
			return "", fmt.Errorf("redirect error: input location[%s] is not match configured[%s]", location, reurl)
		}
		paramsvalues = uu.Query()
	} else {
		//传入参数中无http或者https前缀， 则认为此时传入的是querystring形式的附加参数，
		// 拼接上数据库中配置的重定向地址进行重定向
		if paramsvalues, err = url.ParseQuery(location); err != nil {
			return "", fmt.Errorf("redirect error: input location[%s] is not url and ParseQuery failed", location)
		}
	}

	m := u.Query()
	var needuatk = true
	for key, value := range paramsvalues {
		switch value[0] {
		case "${code}":
			m.Add(key, uatk)
			needuatk = false
		case "${uatk}":
			m.Add(key, uatk)
			needuatk = false
		case "${pamuser}":
			if pamuser, err := checkPermission.EnvGetPamUserInfo(user.F_user_id); err != nil {
				panic("get pam user failed")
			} else if pamuserjsonstr, err := json.Marshal(pamuser); err != nil {
				panic("json Marshal pam user failed:" + err.Error())
			} else {
				m.Add(key, string(pamuserjsonstr))
			}
		default:
			m.Add(key, value[0])
		}
	}

	if needuatk {
		m.Add("uatk", uatk)
	}

	keys := make([]string, 0, len(m))
	for k, v := range m {
		keys = append(keys, k+"="+v[0])
	}
	params := strings.Join(keys, "&")
	if fregment != "" {
		return reurl + "?" + params + "#" + fregment, nil
	}
	return reurl + "?" + params, nil
}
