/***********************************************************************
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php.
//******
// Filename:
// Description:
// Author:
// CreateTime:
/***********************************************************************/
package getcaptchaimage

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo"
	"github.com/lifei6671/gocaptcha"
	"net/http"
	. "github.com/qoobing/userd/src/config"
	. "github.com/qoobing/userd/src/const"
	"github.com/qoobing/userd/src/types"
	. "qoobing.com/utillib.golang/api"
	"qoobing.com/utillib.golang/gls"
	"qoobing.com/utillib.golang/log"
	"qoobing.com/utillib.golang/xyz"
	"strings"
	"time"
)

type Input struct {
	Sign string `json:"sign" query:"sign" validate:"required,len=32"`
}

type Output struct {
	Eno        int    `json:"eno"`
	Err        string `json:"err"`
	Vcodekey   string `json:"vcodekey"`
	Vcodeimage string `json:"vcodeimage" view:"logignore"`
}

func Main(cc echo.Context) error {
	//Step 1. init apicontext
	c := cc.(ApiContext)
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

	//Step 3. generate verify code by estimated name type [mobile or email]
	vcodekey, vcodeimage, err := GenerateCaptchaImage(input.Sign)
	if err != nil {
		log.Debugf(err.Error())
		return c.RESULT_ERROR(ERR_GETVCODE_ERROR, "get verify code failed")
	}

	//Step 4. set output
	output.Eno = 0
	output.Err = "success"
	output.Vcodekey = vcodekey
	output.Vcodeimage = vcodeimage

	return c.RESULT(output)
}

func GenerateCaptchaImage(sign string) (vcodekey, vcodeimage string, err error) {
	const (
		CAPTCHA_WIDTH  = 150
		CAPTCHA_HEIGHT = 50
	)
	//初始化一个验证码对象
	captchaImage := gocaptcha.NewCaptchaImage(CAPTCHA_WIDTH, CAPTCHA_HEIGHT, gocaptcha.RandLightColor())
	if captchaImage == nil {
		log.Debugf("get captcha error, NewCaptchaImage return null")
		return vcodekey, vcodeimage, fmt.Errorf("get captcha error, NewCaptchaImage return null")
	}
	//生成验证token
	var captcha2save = types.CaptchaInfo{
		Sign:     sign,
		Vcodekey: xyz.GetRandomCharString(20),
		Vcode:    gocaptcha.RandText(4),
	}
	vcodekey = captcha2save.Vcodekey

	//画随机噪点
	captchaImage.DrawNoise(gocaptcha.CaptchaComplexLower)
	//画随机文字噪点
	captchaImage.DrawTextNoise(gocaptcha.CaptchaComplexLower)
	//画验证码文字，可以预先保持到Session种或其他储存容器种
	captchaImage.DrawText(captcha2save.Vcode)
	//画边框
	//captchaImage.DrawBorder(gocaptcha.ColorToRGB(0x17A7A7A))
	//captchaImage.DrawSineLine()
	//captchaImage.DrawHollowLine()

	//Step 2. save vcodekey and vcode to redis
	var (
		dataHead = "data:image/png;base64,"
		value, _ = json.Marshal(captcha2save)
		rds      = gls.GetGlsValueNotNil("redis").(*RedisConn)
		c        = gls.GetGlsValueNotNil("context").(ApiContext)
		rdskey   = "CAPTCHA-" + vcodekey
		expire   = Config().CaptchaTimeout
	)
	if expire < 30 {
		expire = 180
	}
	_, err = rds.Do("SETEX", rdskey, expire, value)
	if err != nil {
		log.Debugf("failed saved [%s], value:[%s]", rdskey, value)
		return vcodekey, vcodeimage, err
	}
	log.Debugf("success saved [%s], value:[%s]", rdskey, value)

	//Step 3. set cookie
	cookie := new(http.Cookie)
	cookie.Name = COOKIE_NAME_CAPTCHA
	cookie.Domain = COOKIE_DOMAIN_USERINFO
	cookie.Path = COOKIE_DEFAULT_PATH
	cookie.Value = vcodekey
	cookie.Expires = time.Now().Add(time.Duration(expire) * time.Second)
	c.SetCookie(cookie)

	//Step 3. 将验证码保存到输出流，可以是文件或HTTP流等
	buf := bytes.NewBuffer(nil)
	captchaImage.SaveImage(buf, gocaptcha.ImageFormatPng)
	//base64编码 必须编码实际内容长度，否则无法正常显示图片
	lendst := base64.StdEncoding.EncodedLen(len(buf.Bytes()))
	imagesrc := make([]byte, lendst)
	base64.StdEncoding.Encode(imagesrc, buf.Bytes())
	vcodeimage = dataHead + string(imagesrc)

	return vcodekey, vcodeimage, nil
}

func VerifyCaptcha(vcodekey string, inputvcode string, sign string) bool {
	var (
		rds = gls.GetGlsValueNotNil("redis").(*RedisConn)
		c   = gls.GetGlsValueNotNil("context").(ApiContext)
	)
	if vcodekey == "" {
		cname := COOKIE_NAME_CAPTCHA
		if hc, err := c.Cookie(cname); err != nil {
			log.Debugf("VerifyCaptcha failed: get cookie '" + cname + "' error:" + err.Error())
			return false
		} else {
			vcodekey = hc.Value
		}
	}
	rdskey := "CAPTCHA-" + vcodekey
	jsondata, err := rds.Do("GET", rdskey)
	if err != nil {
		log.Debugf("verify email vcode failed: GET " + rdskey + " error:" + err.Error())
		return false
	} else if jsondata == nil {
		log.Debugf("verify email vcode failed: GET " + rdskey + " return nil[not found]")
		return false
	}

	var vc types.CaptchaInfo
	err = json.Unmarshal(jsondata.([]byte), &vc)
	if err != nil {
		log.Debugf("verify email vcode failed:  unmarshal error:" + err.Error())
		return false
	} else if strings.ToUpper(vc.Vcode) != strings.ToUpper(inputvcode) {
		log.Debugf("verify vcode failed: input vcode not match saved vcode")
		return false
	}
	return true
}

func init() {
	err := gocaptcha.ReadFonts("fonts", ".ttf")
	if err != nil {
		log.Fatalf("read fonts for captcha failed, error:%s", err.Error())
		return
	}
}
