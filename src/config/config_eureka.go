// +build springboot

package config

import (
	"github.com/labstack/echo"
	"qoobing.com/utillib.golang/eureka"
	"qoobing.com/utillib.golang/log"
	"qoobing.com/utillib.golang/xyz"
)

func InitEureka() {
	e := Config().SpringBoot.Eureka
	if e.Name != "" {
		var (
			name    = e.Name
			port    = xyz.If(e.Port == "", "8080", e.Port).(string)
			sport   = xyz.If(e.SecurePort == "", "8443", e.Port).(string)
			options = map[string]string{
				"ipAddress": e.ServiceIp,
			}
		)
		go func() {
			if e.EurekaAddr != "" {
				log.Debugf("eureka supported, RegisterAt('%s','%s',%s,%s)", e.EurekaAddr, name, port, sport)
				eureka.RegisterAt(e.EurekaAddr, name, port, sport, options)
			} else {
				log.Debugf("eureka supported, Register('%s',%s,%s)", name, port, sport)
				eureka.Register(name, port, sport, options)
			}
		}()
	} else {
		log.Debugf("eureka not configured")
	}
}

func Health(cc echo.Context) error {
	return nil
}
