package utils

import (
	"errors"
	"fmt"
	"regexp"
	"time"

	log "github.com/sirupsen/logrus"
)

// ParseAccountaor 解析分机号,注册地址 sip:1004@192.168.11.242 =》 1004，192.168.11.242
func ParseAccountaor(aor string) (exten, host string, error error) {
	if aor == "" {
		return "", "", errors.New("aor is empty")
	}
	log.Debugf("parse accountaor %s", aor)
	re := regexp.MustCompile(`sip:(\d+)@(\S+)`)
	matched := re.FindAllStringSubmatch(aor, -1)
	for _, match := range matched {
		return match[1], match[2], nil
	}
	return "", "", errors.New("aor format error")
}

//ParseRegInfo 解析注册信息, OK, ERR 彩色显示
// 	--- User Agents (2) ---
// 	> sip:1004@192.168.11.242                    OK
// 	  sip:1005@192.168.11.242                    ERR
// `
func ParseRegInfo(data string) (result map[string]NewEventInfo) {
	log.Infof("ParseRegInfo input %s", data)
	result = make(map[string]NewEventInfo)
	re := regexp.MustCompile(`sip:(\d+)@(\S+)\s+\S+(OK|ERR)`)
	matched := re.FindAllStringSubmatch(data, -1)
	for _, match := range matched {
		key := fmt.Sprintf("%s-%s", match[1], match[2])
		log.Debugf("exten %s, status is: %s\n", key, match[3])

		eventType := "REGISTER_OK"
		regStatus := "ok"
		if match[3] == "ERR" {
			regStatus = "fail"
			eventType = "REGISTER_FAIL"
		}

		result[key] = NewEventInfo{
			Host:      match[2],
			Exten:     match[1],
			RegStatus: regStatus,
			TimeStamp: time.Now().Format("2006-01-02 15:04:05"),
			EventInfo: EventInfo{
				Event:      true,
				Class:      "init register",
				Accountaor: fmt.Sprintf("sip:%s@%s", match[1], match[2]),
				Type:       eventType,
				Param:      "cmd response reginfo",
			},
		}

	}
	log.Infof("ParseRegInfo %+v", result)
	return
}

//ParseActiveCall 解析激活通话
func ParseActiveCall(data string) {

}
