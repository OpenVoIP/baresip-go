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

//ConnectInfo 连接信息
type ConnectInfo struct {
	conn     net.Conn
	writeMsg chan string             // 发送到 baresip 的消息
	events   chan utils.NewEventInfo // 记录从 baresip 获取的事件, 命令响应
	stop     chan bool               // 标记连接是否断开
}

var ctx = context.Background()

var connectInfoInstacne *ConnectInfo

//InitConn get conn
func InitConn() *ConnectInfo {
	connectInfoInstacne = &ConnectInfo{
		writeMsg: make(chan string, 100),
		events:   make(chan utils.NewEventInfo, 1000),
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
		msg := <-info.writeMsg
		_, err := info.conn.Write([]byte(fmt.Sprintf("%d:%s,", len(msg), msg)))
		if err != nil {
			log.Errorf("tcp connect write error", err)
			info.stop <- true
			return
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
				var eventInfo utils.NewEventInfo
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
					fallthrough
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
				// FIXME 在通话过程中也会出现 RegStatus  状态，此操作导致状态覆盖
				// 若注册成功且Status 为空，设置为  idle 状态
				// if event.RegStatus == "ok" && event.Status == "" {
				// 	event.Status = "idle"
				// }

			}

			// TODO 由于 baresip 与 asterisk 交互存在多余事件 "param":"401 Unauthorized"
			if event.Param == "401 Unauthorized" {
				continue
			}

			// 分机通话状态 写入 redis
			if event.Event && event.Status != "" {
				eventJSON, _ := json.Marshal(event)
				event.Exten, event.Host, _ = utils.ParseAccountaor(event.Accountaor)
				if err := RedisInstance.Set(ctx, fmt.Sprintf("baresip-call-status-%s-%s", event.Exten, event.Host), eventJSON, 0).Err(); err != nil {
					log.Errorf("redis write %+v", err)
				}
			}

			// 分机注册状态
			if event.RegStatus != "" {
				regStatus := map[string]string{"status": event.RegStatus, "cause": event.Param}
				statusJSON, _ := json.Marshal(regStatus)
				event.Exten, event.Host, _ = utils.ParseAccountaor(event.Accountaor)
				if err := RedisInstance.Set(ctx, fmt.Sprintf("baresip-reg-status-%s-%s", event.Exten, event.Host), statusJSON, 0).Err(); err != nil {
					log.Errorf("redis write %+v", err)
				}
			}

			// 查询分机注册状态响应, 将 response 转为 reg event
			if strings.Contains(event.Data, "User Agents") {
				for _, value := range utils.ParseRegInfo(event.Data) {
					regStatus := map[string]string{"status": value.RegStatus, "cause": "init query"}
					statusJSON, _ := json.Marshal(regStatus)
					if err := RedisInstance.Set(ctx, fmt.Sprintf("baresip-reg-status-%s-%s", value.Exten, value.Host), statusJSON, 0).Err(); err != nil {
						log.Errorf("redis write %+v", err)
					}

					// 如果分机状态 ok, 设置通话状态默认为 idle
					callStatus := map[string]string{"cause": "init query"}
					if value.RegStatus == "ok" {
						callStatus["regstatus"] = "ok"
						callStatus["status"] = "idle"
					} else {
						callStatus["regstatus"] = value.RegStatus
						callStatus["status"] = ""
					}
					callStatusJSON, _ := json.Marshal(callStatus)
					if err := RedisInstance.Set(ctx, fmt.Sprintf("baresip-call-status-%s-%s", value.Exten, value.Host), callStatusJSON, 0).Err(); err != nil {
						log.Errorf("redis write %+v", err)
					}

					PublishEvent(value)
				}
			}

			// 当前通话列表
			// if strings.Contains(event.Data, "Active calls") {
			// 	utils.ParseActiveCall(event.Data)
			// }

			// 发布订阅
			PublishEvent(event)

			// 日志调试
			// log.Debugf("event %+v", event)
			// log.Infof("data %s", event.Data)
		case <-info.stop:
			log.Info("exit loop event handle")
			return
		}
	}
}
