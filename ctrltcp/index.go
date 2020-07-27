package ctrltcp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"time"

	log "github.com/sirupsen/logrus"
)

/**
https://github.com/baresip/baresip/blob/b87324556a4c5fa9c2ad90c5293067268a31a75f/modules/ctrl_tcp/ctrl_tcp.c
连接 baresip ctrl_tcp 模块
*/

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
	Status    string `json:"status"` //从 type 获取当前状态
	TimeStamp string `json:"timestamp"`

	EventInfo
}

//ConnectInfo 连接信息
type ConnectInfo struct {
	conn     net.Conn
	writeMsg chan string    // 发送到 baresip 的消息
	events   chan EventInfo // 记录从 baresip 获取的事件
	stop     chan bool      // 标记连接是否断开
}

var ctx = context.Background()

var connectInfoInstacne *ConnectInfo

//InitConn get conn
func InitConn() *ConnectInfo {
	connectInfoInstacne = &ConnectInfo{
		writeMsg: make(chan string, 100),
		events:   make(chan EventInfo, 1000),
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
		result.Write(buf[0:n])
		if err != nil {
			log.Errorf("read err: %+v", err)
			info.stop <- true
			break
		} else {
			scanner := bufio.NewScanner(result)
			scanner.Split(packetSlitFunc)
			for scanner.Scan() {
				var eventInfo EventInfo
				json.Unmarshal(scanner.Bytes()[:], &eventInfo)
				// log.Debugf("socket read %+v", eventInfo)
				info.events <- eventInfo
			}
		}
		result.Reset()
	}
}

//eventHandle 事件处理
func (info *ConnectInfo) eventHandle() {
	log.Info("for event handle")
	for {
		select {
		case data := <-info.events:
			newEvent := NewEventInfo{EventInfo: data, TimeStamp: time.Now().Format("2006-01-02 15:04:05")}
			switch data.Type {
			case "CALL_PROGRESS":
				newEvent.Status = "ringing"
			case "CALL_INCOMING":
				newEvent.Status = "ring"
			case "CALL_ESTABLISHED":
				newEvent.Status = "answer"
			case "CALL_CLOSED":
				newEvent.Status = "hangup"
			}

			// 发布订阅
			RedisInstance.Publish(ctx, "session-channel", newEvent)

			// 写入 redis
			err := RedisInstance.Set(ctx, "baresip-status", data.Type, 0).Err()
			if err != nil {
				log.Errorf("redis write %+v", err)
			}

			// 日志调试
			log.Debugf("event %+v", newEvent)
		case <-info.stop:
			log.Info("exit for event handle")
			return
		}
	}
}
