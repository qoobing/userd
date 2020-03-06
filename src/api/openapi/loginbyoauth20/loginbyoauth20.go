/***********************************************************************
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php.
//******
// Filename:
// Description:
// Author:
// CreateTime:
/***********************************************************************/
package loginbyoauth20

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/labstack/echo"
	"io"
	"net/http"
	"net/url"
	"github.com/qoobing/userd/src/api/oauth20qrcode/setqrcodeuserinfo"
	"github.com/qoobing/userd/src/api/openapi/login"
	"github.com/qoobing/userd/src/api/rbac/basic/checkPermission"
	. "github.com/qoobing/userd/src/config"
	. "github.com/qoobing/userd/src/const"
	"github.com/qoobing/userd/src/model"
	"github.com/qoobing/userd/src/model/t_login_scene"
	"github.com/qoobing/userd/src/model/t_user"
	. "qoobing.com/utillib.golang/api"
	"qoobing.com/utillib.golang/log"
	. "qoobing.com/utillib.golang/rpc"
	"strconv"
	"time"
)

type Input struct {
	Type     int    `form:"type"     json:"type"     validate:"required,min=1,max=4"` //3钉钉
	Code     string `form:"code"     json:"code"     validate:"required,min=1,max=64"`
	Scene    string `form:"scene"    json:"scene"    validate:"required,min=4,max=16"`
	Location string `form:"location" json:"location" validate:"omitempty,min=1,max=1024"`
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
		user   t_user.User
	)
	if err := c.BindInput(&input); err != nil {
		return c.RESULT_PARAMETER_ERROR(err.Error())
	}

	if input.Type == AUTH20ID_LOGIN_TYPE_WX {
		//Step 3. get auth 2.0 server access token by code
		access_token, openid, err := getWxAccessTokenByCode(input.Code)
		if err != nil {
			return c.RESULT_ERROR(ERR_LOGIN_FAILED, "login failed: get wx access token failed")
		}

		//Step 4. find user by id
		user, err = t_user.FindUserByAuth20id(c.Mysql(), openid, AUTH20ID_TYPE_WX_OPENID)
		if err != nil && err.Error() == USER_NOT_EXIST {
			log.Debugf("user not exist, continue create new user")
			//Step 4.1 get wx userinfo
			user, err = getWxUserinfo(access_token, openid)
			if err != nil {
				return c.RESULT_ERROR(ERR_LOGIN_FAILED, "login failed: get wx userinfo failed")
			}
			//Step 4.2 register new user to database
			err = t_user.CreateUser(c.Mysql(), &user)
			if err != nil {
				return c.RESULT_ERROR(ERR_CREATE_USER_ERROR, err.Error())
			}
		} else if err != nil {
			log.Debugf("FindUserByAuth20id error %s", err.Error())
			return c.RESULT_ERROR(ERR_LOGIN_FAILED, "find user error")
		}

	} else if input.Type == AUTH20ID_LOGIN_TYPE_DD && input.Scene == "pamuserd" { //钉钉登录
		//Step 3. get auth 2.0 server access token by code
		var err error
		duser, err := getDingUserinfoForPamuserd(input.Code)
		if err != nil {
			return c.RESULT_ERROR(ERR_LOGIN_FAILED, "login failed: get ding userinfo failed,"+err.Error())
		}
		//Step 4. find user by id
		user, err = t_user.FindUserByAuth20id(c.Mysql(), duser.F_exid_dd_unionid, AUTH20ID_TYPE_DING_UNIONID)
		if err != nil && err.Error() == USER_NOT_EXIST {
			log.Debugf("user not exist")
			return c.RESULT_ERROR(ERR_LOGIN_FAILED, "user not exist")
		} else if err != nil {
			log.Debugf("FindUserByAuth20id error %s", err.Error())
			return c.RESULT_ERROR(ERR_LOGIN_FAILED, "find user error")
		}
	} else if input.Type == AUTH20ID_LOGIN_TYPE_DD { //钉钉登录
		//Step 3. get auth 2.0 server access token by code
		var err error
		duser, err := getDingUserinfo(input.Code)
		if err != nil {
			return c.RESULT_ERROR(ERR_LOGIN_FAILED, "login failed: get ding userinfo failed,"+err.Error())
		}
		//Step 4. find user by id
		user, err = t_user.FindUserByAuth20id(c.Mysql(), duser.F_exid_dd_openid, AUTH20ID_TYPE_DING_OPENID)
		if err != nil && err.Error() == USER_NOT_EXIST {
			log.Debugf("user not exist, continue create new user")
			err = t_user.CreateUser(c.Mysql(), &duser)
			if err != nil {
				return c.RESULT_ERROR(ERR_CREATE_USER_ERROR, err.Error())
			}
			user = duser
		} else if err != nil {
			log.Debugf("FindUserByAuth20id error %s", err.Error())
			return c.RESULT_ERROR(ERR_LOGIN_FAILED, "find user error")
		}
	} else if input.Type == AUTH20ID_LOGIN_TYPE_UW { //酉告文武登录
		//Step 3. get auth 2.0 server access token by code
		var err error
		user, err = setqrcodeuserinfo.GetQRCodeResult(input.Code)
		if err != nil {
			return c.RESULT_ERROR(ERR_LOGIN_FAILED, "login failed: get userinfo failed:"+err.Error())
		}
	}

	//Step 5. find use scene
	scene, err := t_login_scene.FindLoginScene(c.Mysql(), input.Scene)
	if err != nil {
		log.Debugf("FindLoginScene error %s", err.Error())
		return c.RESULT_ERROR(ERR_LOGIN_SCENE_INVALID, "find user_scene error:"+err.Error())
	}
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

func getWxAccessTokenByCode(code string) (access_token string, openid string, err error) {
	urlstr := Config().Auth20Login.WxAccessTokenUrl + "?"
	input := map[string]string{
		"appid":      Config().Auth20Login.WxAppId,
		"secret":     Config().Auth20Login.WxAppSecret,
		"code":       code,
		"grant_type": "authorization_code",
	}
	for k, v := range input {
		urlstr += k + "=" + v + "&"
	}
	output := struct {
		Errcode       int    `json:"errcode"`       //错误码
		Errmsg        string `json:"errmsg"`        //错误信息
		Access_token  string `json:"access_token"`  //接口调用凭证
		Expires_in    int    `json:"expires_in"`    //access_token接口调用凭证超时时间，单位（秒）
		Refresh_token string `json:"refresh_token"` //用户刷新access_token
		Openid        string `json:"openid"`        //授权用户唯一标识
		Scope         string `json:"scope"`         //用户授权的作用域，使用逗号（,）分隔
	}{}
	err = HttpGet(urlstr, &output)
	if err != nil {
		return "", "", err
	} else if output.Errcode != 0 {
		log.Debugf("wx return: errcode:%d, errmsg:%s", output.Errcode, output.Errmsg)
		return "", "", errors.New("wx return:" + output.Errmsg)
	}
	return output.Access_token, output.Openid, nil
}

func getWxUserinfo(access_token string, openid string) (user t_user.User, err error) {
	urlstr := Config().Auth20Login.WxUserinfoUrl
	urlstr += "?access_token=" + access_token
	urlstr += "&openid=" + openid

	output := struct {
		Errcode    int    `json:"errcode"`    //错误码
		Errmsg     string `json:"errmsg"`     //错误信息
		Openid     string `json:"openid"`     //普通用户的标识，对当前开发者帐号唯一
		Nickname   string `json:"nickname"`   //普通用户昵称
		Sex        int    `json:"sex"`        //普通用户性别，1为男性，2为女性
		Frovince   string `json:"province"`   //普通用户个人资料填写的省份
		City       string `json:"city"`       //普通用户个人资料填写的城市
		Country    string `json:"country"`    //国家，如中国为CN
		Headimgurl string `json:"headimgurl"` //用户头像，最后一个数值代表正方形头像大小（有0、46、64、96、132数值可选，0代表640*640正方形头像），用户没有头像时该项为空
		Unionid    string `json:"unionid"`    //用户统一标识。针对一个微信开放平台帐号下的应用，同一用户的unionid是唯一的。
	}{}
	err = HttpGet(urlstr, &output)
	if err != nil {
		return user, err
	}
	user = t_user.User{
		F_avatar:          output.Headimgurl,
		F_nickname:        output.Nickname,
		F_name:            output.Nickname,
		F_exid_wx_openid:  output.Openid,
		F_exid_wx_unionid: output.Unionid,
	}
	return user, nil
}

func getDingUserinfo(code string) (user t_user.User, err error) {
	urlstr := Config().Auth20Login.DingUserinfoUrl + "?"

	timeStamp := time.Now().UnixNano() / 1e6
	timeStr := strconv.Itoa(int(timeStamp))
	log.Debugf("timestamp is:%d", timeStamp)
	log.Debugf("DingAppSecret is:%s", Config().Auth20Login.DingAppSecret)
	h := hmac.New(sha256.New, []byte(Config().Auth20Login.DingAppSecret)) //[]byte(timeStr)
	io.WriteString(h, timeStr)
	log.Debugf("after hmac sha256 is:%s", hex.EncodeToString(h.Sum(nil)))
	encodeStr := base64.StdEncoding.EncodeToString(h.Sum(nil))
	log.Debugf("after base64 encode is:%s", encodeStr)
	sign := url.QueryEscape(encodeStr)
	log.Debugf("signature is:%s", sign)
	//sign := encodeStr

	param := map[string]string{
		"signature": sign,
		"timestamp": timeStr,
		"accessKey": Config().Auth20Login.DingAppId,
	}
	for k, v := range param {
		urlstr += k + "=" + v + "&"
	}

	input := struct {
		Tmp_auth_code string `json:"tmp_auth_code"`
	}{}
	input.Tmp_auth_code = code

	output := struct {
		Errcode  int    `json:"errcode"` //错误码
		Errmsg   string `json:"errmsg"`  //错误信息
		UserInfo struct {
			Dingid   string `json:"dingId"`
			Nickname string `json:"nick"`    //普通用户昵称
			Openid   string `json:"openid"`  //普通用户的标识，对当前开发者帐号唯一
			Unionid  string `json:"unionid"` //用户统一标识。针对一个微信开放平台帐号下的应用，同一用户的unionid是唯一的。
		} `json:"user_info"`
	}{}

	err = HttpPost(urlstr, input, &output)
	if err != nil {
		return user, err
	}
	log.Debugf("out is:%+v", output)
	user = t_user.User{
		F_nickname:        output.UserInfo.Nickname,
		F_name:            output.UserInfo.Nickname,
		F_exid_dd_openid:  output.UserInfo.Openid,
		F_exid_dd_unionid: output.UserInfo.Unionid,
	}

	if output.Errcode != 0 {
		err = errors.New(output.Errmsg)
	}
	return user, err
}

func getDingUserinfoForPamuserd(code string) (user t_user.User, err error) {

	output := struct {
		Errcode      int    `json:"errcode"`      //错误码
		Errmsg       string `json:"errmsg"`       //错误信息
		Access_token string `json:"access_token"` //
	}{}
	requrl := fmt.Sprintf(
		"https://oapi.dingtalk.com/gettoken?appkey=%s&appsecret=%s",
		Config().Auth20Login.DingPamAppKey, Config().Auth20Login.DingPamAppSecret)
	err = HttpGet(requrl, &output)
	if err != nil {
		return user, err
	} else if output.Errcode != 0 {
		return user, fmt.Errorf("step1 errcode:%d, errmsg:%s", output.Errcode, output.Errmsg)
	}
	log.Debugf("step1 out1 is:%+v", output)

	output2 := struct {
		Errcode int    `json:"errcode"` //错误码
		Errmsg  string `json:"errmsg"`  //错误信息
		Userid  string `json:"userid"`  //
	}{}
	requrl = fmt.Sprintf("https://oapi.dingtalk.com/user/getuserinfo?access_token=%s&code=%s",
		output.Access_token, code)
	err = HttpGet(requrl, &output2)
	if err != nil {
		return user, err
	} else if output2.Errcode != 0 {
		return user, fmt.Errorf("step2 errcode:%d, errmsg:%s", output2.Errcode, output2.Errmsg)
	}
	log.Debugf("step2 out2 is:%+v", output2)

	output3 := struct {
		Errcode int    `json:"errcode"` //错误码
		Errmsg  string `json:"errmsg"`  //错误信息
		Userid  string `json:"userid"`  //
		Unionid string `json:"unionid"` //
	}{}
	requrl = fmt.Sprintf("https://oapi.dingtalk.com/user/get?access_token=%s&userid=%s",
		output.Access_token, output2.Userid)
	err = HttpGet(requrl, &output3)
	if err != nil {
		return user, err
	} else if output3.Errcode != 0 {
		return user, fmt.Errorf("step3 errcode:%d, errmsg:%s", output3.Errcode, output3.Errmsg)
	}
	log.Debugf("step3 out3 is:%+v", output3)

	user = t_user.User{
		F_exid_dd_unionid: output3.Unionid,
	}

	return user, err
}
