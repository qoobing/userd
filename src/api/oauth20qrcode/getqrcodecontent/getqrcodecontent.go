/***********************************************************************
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php.
//******
// Filename:
// Description:
// Author:
// CreateTime:
/***********************************************************************/
package getqrcodecontent

import (
	"github.com/labstack/echo"
	. "github.com/qoobing/userd/src/config"
	. "github.com/qoobing/userd/src/const"
	. "qoobing.com/utillib.golang/api"
	"qoobing.com/utillib.golang/gls"
	"qoobing.com/utillib.golang/log"
	"qoobing.com/utillib.golang/xyz"
)

type Input struct {
	Scene  string `json:"scene"    validate:"required,max=16"`
	Params string `json:"params"`
}

type Output struct {
	Eno      int    `json:"eno"`
	Err      string `json:"err"`
	Content  string `json:"content"`  //兵分两路之第一路： 二维码展示，传递参数给扫码后的页面
	TempCode string `json:"tempcode"` //兵分两路之第二路： 调用者查询用户操作结果所需参数
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

	//Step 3. generate code and save to redis
	var (
		tempcode     = "Q" + xyz.GetRandomCharString(18)
		content, err = genQrContent(tempcode, input.Scene, input.Params)
	)
	if err != nil {
		return c.RESULT_ERROR(ERR_GENERATE_QRCODE_ERROR, err.Error())
	}

	//Step 4. set output
	output.Eno = 0
	output.Err = "success"
	output.Content = content
	output.TempCode = tempcode

	return c.RESULT(output)
}

func genQrContent(tempcode, scene, params string) (content string, err error) {
	content = "https://qoobing.com/uc/" + tempcode
	err = saveQRCodeContent(tempcode)
	return content, err
}

func saveQRCodeContent(tempcode string) (err error) {
	//Step 1. generate codekey and content
	var (
		tempcodekey   = "QRCODE_TEMPCODE-" + tempcode
		tempcodevalue = "WAIT_COMMIT"
		expiretime    = Config().QrCodeTimeout
	)
	if expiretime <= 0 {
		expiretime = 300
	}

	//Step 2. save to redis
	rds := gls.GetGlsValueNotNil("redis").(*RedisConn)
	_, err = rds.Do("SETEX", tempcodekey, expiretime, tempcodevalue)
	if err != nil {
		log.Debugf("failed saved [%s], value:[%s]", tempcodekey, tempcodevalue)
		return err
	}
	log.Debugf("success saved [%s], value:[%s]", tempcodekey, tempcodevalue)

	return nil
}
