package logs

import (
	"fmt"
	"github.com/bwgame666/common/libs"
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"strings"
	"time"
)

// NatsHook 定义Nats钩子结构体
type NatsHook struct {
	conn *nats.Conn
}

// NewNatsHook 创建Nats钩子
func NewNatsHook(servers []string, username, password string) (*NatsHook, error) {
	var url string
	if username != "" {
		url = fmt.Sprintf("nats://%s:%s@%s", username, password, servers[0])
	} else {
		url = fmt.Sprintf("nats://%s", servers[0])
	}

	conn, err := nats.Connect(url)
	if err != nil {
		return nil, err
	}
	return &NatsHook{
		conn: conn,
	}, nil
}

// Levels 实现logrus.Hook接口的Levels方法
func (h *NatsHook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.ErrorLevel,
		logrus.WarnLevel,
		logrus.InfoLevel,
	}
}

// Fire 实现logrus.Hook接口的Fire方法
func (h *NatsHook) Fire(entry *logrus.Entry) error {
	level := entry.Data["@level"]
	if level == "I" {
		return nil
	}

	telegram := entry.Data["@telegram"]
	if telegram == true {
		logMessage := fmt.Sprintf("%s %s [%s] %s.%s():%d %s",
			entry.Time.Format(defaultTimestampFormat),
			entry.Data["@level"],
			entry.Data["@logName"],
			entry.Data["@fileName"],
			entry.Data["@funcName"],
			entry.Data["@line"],
			entry.Message)
		_ = h.conn.Publish("log-telegram", []byte(logMessage))
	}

	index := entry.Data["@logName"].(string)
	if index == "" {
		index = "default"
	}
	index = strings.Split(index, ".")[0]

	ts := time.Now()
	data := map[string]interface{}{
		"_index":     fmt.Sprintf("%s_%04d%02d", index, ts.Year(), ts.Month()),
		"@level":     entry.Data["@level"],
		"@pos":       fmt.Sprintf("%s.%s():%d", entry.Data["@fileName"], entry.Data["@funcName"], entry.Data["@line"]),
		"@content":   entry.Message,
		"@timestamp": entry.Time.Format(defaultTimestampFormat),
	}
	payload, _ := libs.JsonMarshal(data)
	return h.conn.Publish("log-zinc", payload)
}

// Close 关闭Nats连接
func (h *NatsHook) Close() error {
	h.conn.Close()
	return nil
}
