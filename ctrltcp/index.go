package ctrltcp

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
)

/**
https://github.com/baresip/baresip/blob/b87324556a4c5fa9c2ad90c5293067268a31a75f/modules/ctrl_tcp/ctrl_tcp.c
连接 baresip ctrl_tcp 模块
*/

type write chan string
type read chan string

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

//ConnectInfo 连接信息
type ConnectInfo struct {
	conn     net.Conn
	writeMsg write
	readMsg  read
	events   chan EventInfo
}

var info *ConnectInfo
var once sync.Once

//GetConn get conn
func GetConn() *ConnectInfo {
	once.Do(func() {
		info = &ConnectInfo{
			writeMsg: make(write),
			readMsg:  make(read),
			events:   make(chan EventInfo, 1000),
		}
		connect()
	})

	return info
}

//connect 连接服务
func connect() {
	var err error
	info.conn, err = net.Dial("tcp", "127.0.0.1:4444")
	if err != nil {
		fmt.Printf("connect failed, err : %v\n", err.Error())
		return
	}
	// go info.writeData()
	go info.readData()
	// test
	go func() {
		for {
			data := <-info.events
			fmt.Printf("data %+v\n", data)
		}
	}()
}

func (info *ConnectInfo) writeData() {
	select {
	case msg := <-info.writeMsg:
		info.conn.Write([]byte(msg))
	}
}

//249:{"event":true,"type":"CALL_INCOMING","class":"call","accountaor":"sip:100@192.168.11.150","direction":"incoming","peeruri":"sip:101@192.168.11.150","peerdisplayname":"101","id":"687ec2b9-9427-4c15-b30e-dba29b99fdca","param":"sip:101@192.168.11.150"},
// 协议格式 长度:数据,
func packetSlitFunc(data []byte, atEOF bool) (advance int, token []byte, err error) {
	// 检查 atEOF 参数
	if !atEOF && len(data) > 1 {
		var l int64
		// 读出 数据包中 实际数据 的长度(大小为 : 号前数据)
		index := strings.IndexByte(string(data), ':')
		if index < 2 {
			return
		}
		l, _ = strconv.ParseInt(string(data[0:index]), 10, 64)
		pl := int(l) + index + 1 // : 占了一位
		if pl <= len(data) {
			return pl, data[index+1 : pl], nil
		}
	}
	return
}

func (info *ConnectInfo) readData() {
	defer info.conn.Close()
	result := bytes.NewBuffer(nil)
	var buf [65542]byte // 标识数据包长度
	for {
		n, err := info.conn.Read(buf[0:])
		result.Write(buf[0:n])
		if err != nil {
			if err == io.EOF {
				continue
			} else {
				fmt.Println("read err:", err)
				break
			}
		} else {
			scanner := bufio.NewScanner(result)
			scanner.Split(packetSlitFunc)
			for scanner.Scan() {
				var eventInfo EventInfo
				json.Unmarshal(scanner.Bytes()[:], &eventInfo)
				info.events <- eventInfo
			}
		}
		result.Reset()
	}
}
