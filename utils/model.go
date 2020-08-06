package utils

//ResponseInfo 命令响应信息
// "response":true,"ok":true, "data":"\n--- Active calls (1) ---\n> [line 1]  0:00:00   INCOMING             sip:1003@192.168.11.242\n\n"
type ResponseInfo struct {
	Response bool `json:"response"`
	// OK       bool   `json:"ok"`
	Data string `json:"data"`
}

//EventInfo baresip 事件信息
type EventInfo struct {
	Event           bool   `json:"event"`
	Type            string `json:"type"`
	Class           string `json:"class"`
	Accountaor      string `json:"accountaor"`
	Direction       string `json:"direction"`
	Peeruri         string `json:"peeruri"`
	Peerdisplayname string `json:"peerdisplayname"`
	ID              string `json:"id"`
	Param           string `json:"param"`
}

//NewEventInfo 重定义新事件
type NewEventInfo struct {
	RegStatus string `json:"regstatus"` //注册状态-是否注册成功
	Status    string `json:"status"`    //从 type 获取分机当前状态-是否通话
	Exten     string `json:"exten"`     // 注册分机号
	Host      string `json:"host"`      // 注册地址
	TimeStamp string `json:"timestamp"`

	EventInfo
	ResponseInfo
}
