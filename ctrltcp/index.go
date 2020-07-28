package ctrltcp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/OpenVoIP/baresip-go/utils"
	log "github.com/sirupsen/logrus"
)

/**
https://github.com/baresip/baresip/blob/b87324556a4c5fa9c2ad90c5293067268a31a75f/modules/ctrl_tcp/ctrl_tcp.c
连接 baresip ctrl_tcp 模块
*/

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

//ConnectInfo 连接信息
type ConnectInfo struct {
	conn     net.Conn
	writeMsg chan string       // 发送到 baresip 的消息
	events   chan NewEventInfo // 记录从 baresip 获取的事件, 命令响应
	stop     chan bool         // 标记连接是否断开
}

var ctx = context.Background()

var connectInfoInstacne *ConnectInfo

//InitConn get conn
func InitConn() *ConnectInfo {
	connectInfoInstacne = &ConnectInfo{
		writeMsg: make(chan string, 100),
		events:   make(chan NewEventInfo, 1000),
		stop:     make(chan bool),
	}
	return connectInfoInstacne
}

//Connect 连接服务
func (info *ConnectInfo) Connect(server string, port int) {
	var err error
	address := fmt.Sprintf("%s:%d", server, port)
	info.conn, err = net.DialTimeout("tcp", address, 5*time.Second)
	if err != nil {
		log.Errorf("connect baresip %s failed, err : %+v \n", address, err.Error())
		return
	}
	log.Infof("connect basesip %s success", address)

	err = info.conn.(*net.TCPConn).SetKeepAlive(true)
	if err != nil {
		log.Infof("SetKeepAlive failed, err : %+v \n", err.Error())
		return
	}

	go info.writeData()
	go info.readData()
	info.eventHandle()
}

func (info *ConnectInfo) writeData() {
	defer info.conn.Close()

	for {
		select {
		case msg := <-info.writeMsg:
			_, err := info.conn.Write([]byte(fmt.Sprintf("%d:%s,", len(msg), msg)))
			if err != nil {
				log.Errorf("tcp connect write error", err)
				info.stop <- true
				return
			}
		}
	}

}

func (info *ConnectInfo) readData() {
	defer info.conn.Close()

	result := bytes.NewBuffer(nil)
	var buf [65542]byte // 标识数据包长度
	for {
		n, err := info.conn.Read(buf[0:])
		if err != nil {
			log.Errorf("read err: %+v", err)
			info.stop <- true
			break
		} else {
			msg := string(buf[0:n])
			// log.Debugf("socket raw %+v", msg)

			result.Write([]byte(msg))
			scanner := bufio.NewScanner(result)
			scanner.Split(packetSlitFunc)
			for scanner.Scan() {
				res := scanner.Bytes()[:]
				// log.Debugf("socket raw res %d %+v", len(res), string(res))
				if len(res) == 0 {
					continue
				}
				var eventInfo NewEventInfo
				err := json.Unmarshal(res, &eventInfo)
				if err != nil {
					log.Errorf("json unmarshal error %+v", err)
				}
				eventInfo.TimeStamp = time.Now().Format("2006-01-02 15:04:05")
				// log.Debugf("socket event %+v", eventInfo)
				info.events <- eventInfo
			}
		}
		result.Reset()
	}
}

//eventHandle 事件处理
func (info *ConnectInfo) eventHandle() {
	log.Info("loop event/response handle")

	// 初始化处理
	// reginfo
	runCMD(&ControlInfo{CMD: "reginfo"})
	// listcalls
	runCMD(&ControlInfo{CMD: "listcalls"})
	for {
		select {
		case event := <-info.events:
			if event.Event {
				switch event.Type {
				case "CALL_PROGRESS":
				case "CALL_RINGING":
					event.Status = "ringing"
				case "CALL_INCOMING":
					event.Status = "ring"
				case "CALL_ESTABLISHED":
					event.Status = "answer"
				case "CALL_CLOSED":
					event.Status = "idle"
				case "REGISTER_OK":
					event.RegStatus = "ok"
				case "REGISTER_FAIL":
					event.RegStatus = "fail"
				}

				exten, host, err := utils.ParseAccountaor(event.Accountaor)
				if err != nil {
					log.Error(err)
				}
				event.Exten = exten
				event.Host = host
			}

			// 发布订阅
			RedisInstance.Publish(ctx, "session-channel", event)

			// 写入 redis
			// 分机通话状态
			if event.Status != "" {
				err := RedisInstance.Set(ctx, fmt.Sprintf("baresip-call-status-%s", event.Exten), event.Status, 0).Err()
				if err != nil {
					log.Errorf("redis write %+v", err)
				}
			}

			// 分机注册状态
			if event.RegStatus != "" {
				err := RedisInstance.Set(ctx, fmt.Sprintf("baresip-reg-status-%s", event.Exten), event.RegStatus, 0).Err()
				if err != nil {
					log.Errorf("redis write %+v", err)
				}
			}

			// 查询分机注册状态响应
			if strings.Contains(event.Data, "User Agents") {
				for key, value := range utils.ParseRegInfo(event.Data) {
					err := RedisInstance.Set(ctx, fmt.Sprintf("baresip-reg-status-%s", key), value, 0).Err()
					if err != nil {
						log.Errorf("redis write %+v", err)
					}
				}
			}

			if strings.Contains(event.Data, "Active calls") {
				utils.ParseActiveCall(event.Data)
			}

			// 日志调试
			// log.Debugf("event %+v", event)
			// log.Infof("data %s", event.Data)
		case <-info.stop:
			log.Info("exit loop event handle")
			return
		}
	}
}
