package utils

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
)

// ParseAccountaor 解析分机号,注册地址 sip:1004@192.168.11.242 =》 1004，192.168.11.242
func ParseAccountaor(aor string) (exten, host string, error error) {
	if aor == "" {
		return "", "", errors.New("aor is empty")
	}
	result := strings.Split(aor, ":")[1]
	if result == "" {
		return "", "", errors.New("aor format error " + aor)
	}
	data := strings.Split(result, "@")
	return data[0], data[1], nil
}

//ParseRegInfo 解析注册信息, OK, ERR 彩色显示
// 	--- User Agents (2) ---
// 	> sip:1004@192.168.11.242                    OK
// 	  sip:1005@192.168.11.242                    ERR
// `
func ParseRegInfo(data string) (result map[string]string) {
	result = make(map[string]string)
	re := regexp.MustCompile(`sip:(\d+)@(\S+)\s+\S+(OK|ERR)`)
	matched := re.FindAllStringSubmatch(data, -1)
	for _, match := range matched {
		key := fmt.Sprintf("%s-%s", match[1], match[2])
		log.Debugf("exten %s, status is: %s\n", key, match[3])

		result[key] = "ok"
		if match[3] == "ERR" {
			result[key] = "fail"
		}
	}
	return
}

//ParseActiveCall 解析激活通话
func ParseActiveCall(data string) {

}
