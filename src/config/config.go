/***********************************************************************
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php.
//******
// Filename:
// Description:
// Author:
// CreateTime:
/***********************************************************************/
package config

import (
	"github.com/pelletier/go-toml"
	"io/ioutil"
	"qoobing.com/utillib.golang/api"
	"qoobing.com/utillib.golang/log"
	"regexp"
	"sync"
)

type appConfig struct {
	apicontext.ApiConfigBase
	Server            string
	Port              string
	TraceLog          string
	ErrorLog          string
	SpringBoot        SpringBoot
	MobileReg         string
	MobileCompiledReg *regexp.Regexp
	EmailReg          string
	EmailCompiledReg  *regexp.Regexp
	Communicator      Communicator
	UatkTimeout       int
	VCodeTimeout      int
	QrCodeTimeout     int
	CaptchaTimeout    int
	Auth20Login       Auth20Login
}

type Communicator struct {
	ServiceName               string
	CmdSendSms                string
	CmdSendEmail              string
	SmsTemplateCodeRegister   string
	SmsTemplateCodeLogin      string
	EmailTemplateCodeRegister string
	EmailTemplateCodeLogin    string
}

type SpringBoot struct {
	Eureka SpringBootEureka
}

type SpringBootEureka struct {
	EurekaAddr string
	Name       string
	ServiceIp  string
	Port       string
	SecurePort string
}

type Auth20Login struct {
	WxAppId          string
	WxAppSecret      string
	WxAccessTokenUrl string
	WxUserinfoUrl    string
	DingAppId        string
	DingAppSecret    string
	DingUserinfoUrl  string
	DingMysqlConf    string
	PamMysqlConf     string
	DingPamAppKey    string
	DingPamAppSecret string
}

/***
type InneruserService struct{
	DingMysqlConf    string
}
***/

var (
	cfg  appConfig
	once sync.Once
)

func Config() *appConfig {
	once.Do(func() {
		defer log.PrintPreety("config:", &cfg)
		doc, err := ioutil.ReadFile("./conf/userd.conf")
		if err != nil {
			panic("initial config, read config file error:" + err.Error())
		}
		if err := toml.Unmarshal(doc, &cfg); err != nil {
			panic("initial config, unmarshal config file error:" + err.Error())
		}
		if err := toml.Unmarshal(doc, &cfg.ApiConfigBase); err != nil {
			panic("initial config, unmarshal config file error:" + err.Error())
		}
		cfg.MobileCompiledReg, err = regexp.Compile(cfg.MobileReg)
		if err != nil {
			panic("initial config, compile MobileReg error:" + err.Error())
		}
		cfg.EmailCompiledReg, err = regexp.Compile(cfg.EmailReg)
		if err != nil {
			panic("initial config, compile EmailReg error:" + err.Error())
		}
	})
	return &cfg
}
