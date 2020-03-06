package common

import (
	"math/rand"
	"qoobing.com/utillib.golang/log"
	"strings"
	"time"
)

var random = rand.New(rand.NewSource(time.Now().UnixNano()))

func GenerateId(typ string) uint64 {
	switch typ {
	case "USER_ID":
		return uint64(1000000000 + random.Int31n(1000000000))
	case "ROLE_ID":
		return uint64(100000000 + random.Int31n(100000000))
	default:
		return uint64(1000000000 + random.Int31n(1000000000))
	}
}

func GetTimeStamp(dbtimestr string) int64 {
	timestamp, err := time.Parse(time.RFC3339, dbtimestr)
	if err != nil {
		log.Warningf("parse database time string '%s' failed:%s", dbtimestr, err.Error())
		return 0
	}
	return timestamp.Unix()
}

func FormatToSafeValue(mode, value string) string {
	switch mode {
	case "MOBILE":
		if len(value) < 7 {
			return value
		} else {
			return value[0:3] + "****" + value[len(value)-4:]
		}
	case "EMAIL":
		values := strings.Split(value, "@")
		name := values[0]
		if len(values) != 2 {
			return value
		} else if len(name) < 3 {
			return name + "**" + "@" + values[1]
		} else {
			return name[0:2] + "**" + name[len(name)-1:] + "@" + values[1]
		}
	}
	return value
}
