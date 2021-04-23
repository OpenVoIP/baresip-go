package ctrltcp

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/OpenVoIP/baresip-go/utils"
	redis "github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"
)

// 创建 socket 连接，单例模式
var once sync.Once

//RedisInstance redis实例
var RedisInstance *redis.Client

//ControlInfo 控制指令信息
type ControlInfo struct {
	CMD  string `json:"cmd" form:"cmd"`
	Data string `json:"data" form:"data"`
}

//ConnectRedis 连接redis
func ConnectRedis() {
	once.Do(func() {
		RedisInstance = redis.NewClient(&redis.Options{
			Addr:     "localhost:6379",
			Password: "", // no password set
			DB:       0,  // use default DB
		})
	})

}

//HandControlAction 处理 redis队列 控制指令, 每 200 ms 查询一次
func HandControlAction() {
	ticker := time.NewTicker(200 * time.Millisecond)
	for {
		<-ticker.C
		ctx := context.Background()
		result := RedisInstance.RPop(ctx, "sipau-channel")
		value := result.Val()
		// log.Infof("%s get %s", result.Name(), value)
		if value != "" {
			info := &ControlInfo{}
			json.Unmarshal([]byte(value), info)
			runCMD(info)
		}
	}
}

//runCMD 像 baresip 执行命令
func runCMD(info *ControlInfo) {
	data, _ := json.Marshal(map[string]string{"command": info.CMD, "params": info.Data})
	cmd := string(data)
	log.Infof("send cmd %s to baresip  %s", info.CMD, info.Data)
	if connectInfoInstacne != nil && connectInfoInstacne.writeMsg != nil {
		connectInfoInstacne.writeMsg <- cmd
	}
}

// PublishEvent 发布订阅
func PublishEvent(event utils.NewEventInfo) {
	publishJSON, err := json.Marshal(event)
	if err != nil {
		log.Errorf("publish json marshal error %+v", err)
		return
	}
	err = RedisInstance.Publish(ctx, "session-channel", publishJSON).Err()
	if err != nil {
		log.Errorf("publish error %+v", err)
	}
}
