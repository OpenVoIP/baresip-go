package ctrltcp

import (
	"strconv"
	"strings"
)

// 解析 baresip 返回消息
//249:{"event":true,"type":"CALL_INCOMING","class":"call","accountaor":"sip:100@192.168.11.150","direction":"incoming","peeruri":"sip:101@192.168.11.150","peerdisplayname":"101","id":"687ec2b9-9427-4c15-b30e-dba29b99fdca","param":"sip:101@192.168.11.150"},
// 协议格式 长度:数据,
func packetSlitFunc(data []byte, atEOF bool) (advance int, token []byte, err error) {
	// log.Infof("packetSlitFunc %s", data)
	// 检查 atEOF 参数
	if !atEOF && len(data) > 1 {
		var count int
		// 读出 数据包中 实际数据 的长度(大小为 : 号前数据)
		index := strings.IndexByte(string(data), ':')
		if index < 2 {
			return
		}
		// 获取数据长度
		count, _ = strconv.Atoi(string(data[0:index]))

		// 获取结束位置
		endIndex := count + index + 1 // : 占了一位
		if endIndex <= len(data) {
			//need +1, skip ','
			return endIndex + 1, data[index+1 : endIndex], nil
		}
	}
	return
}
