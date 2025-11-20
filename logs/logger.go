package logs

import (
	"context"
	"fmt"
	commonModel "github.com/bwgame666/common/model"
	"github.com/fsnotify/fsnotify"
	"github.com/pelletier/go-toml"
	log "github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

type (
	CustomFormatter struct {
	}

	LoggerManager struct {
		loggers map[string]*Logger
		conf    *LogConf
		logImpl *log.Logger
		env     string
	}

	Logger struct {
		name  string
		impl  *log.Logger
		level Level
	}

	Level int

	// Appender 修改：Appender接口增加Close方法
	Appender interface {
		Write(p []byte) (n int, err error)
		Close() error
	}

	// ConsoleAppender 新增：ConsoleAppender结构体
	ConsoleAppender struct{}

	// FileAppender 新增：FileAppender结构体
	FileAppender struct {
		fileWriter *lumberjack.Logger
	}
)

const (
	OffLevel Level = iota
	LogLevel
	ErrorLevel
	WarnLevel
	InfoLevel
	DebugLevel

	EnvProd = "prod"
	EnvUat  = "uat"
	EnvDev  = "dev"

	defaultTimestampFormat = time.RFC3339
)

var (
	loggerManager     *LoggerManager = nil
	loggerManagerOnce sync.Once
	loggerLevelMap    = map[string]Level{
		"off":   OffLevel,
		"log":   LogLevel,
		"error": ErrorLevel,
		"warn":  WarnLevel,
		"info":  InfoLevel,
		"debug": DebugLevel,
	}

	loggerLevelStringMap = map[Level]string{
		OffLevel:   "OFF",
		LogLevel:   "L",
		ErrorLevel: "E",
		WarnLevel:  "W",
		InfoLevel:  "I",
		DebugLevel: "D",
	}
)

func (f *CustomFormatter) Format(entry *log.Entry) ([]byte, error) {
	pid := os.Getpid()

	logMessage := fmt.Sprintf("%s %s [%s:%d] %s.%s():%d %s\n",
		entry.Time.Format(defaultTimestampFormat),
		entry.Data["@level"],
		entry.Data["@logName"],
		pid,
		entry.Data["@fileName"],
		entry.Data["@funcName"],
		entry.Data["@line"],
		entry.Message)

	return []byte(logMessage), nil
}

func GetLogger(name string) *Logger {
	if loggerManager == nil {
		loggerManagerOnce.Do(func() {
			_ = NewLogManager(nil, "dev")
		})
	}

	logger, ok := loggerManager.loggers[name]
	if !ok {
		defaultLevel := DebugLevel
		switch strings.ToLower(loggerManager.env) {
		case EnvProd:
			defaultLevel = WarnLevel
		case EnvUat:
			defaultLevel = InfoLevel
		case EnvDev:
			defaultLevel = DebugLevel
		default:
			defaultLevel = DebugLevel
		}

		level := loggerManager.GetLoggerLevel(name, defaultLevel)

		if loggerManager.conf.FileAppender {
			host, err := os.Hostname()
			if err != nil {
				host = "unknown"
			}
			filePath := filepath.Dir(loggerManager.conf.File.FilePath)
			filePath = path.Join(filePath, host, fmt.Sprintf("%s_%s.log", name, time.Now().Format("0102T15")))

			fileAppender := &FileAppender{
				fileWriter: &lumberjack.Logger{
					Filename:   filePath,
					MaxSize:    loggerManager.conf.File.MaxSize,
					MaxBackups: loggerManager.conf.File.MaxBackups,
					MaxAge:     loggerManager.conf.File.MaxAge,
					Compress:   loggerManager.conf.File.Compress,
				},
			}
			multiWriters := io.MultiWriter(loggerManager.logImpl.Out, fileAppender)

			impl := log.New()
			impl.SetLevel(log.TraceLevel)
			impl.SetFormatter(&CustomFormatter{})
			impl.SetOutput(multiWriters)
			impl.Hooks = loggerManager.logImpl.Hooks

			logger = &Logger{
				name:  name,
				impl:  impl,
				level: level,
			}
		} else {
			logger = &Logger{
				name:  name,
				impl:  loggerManager.logImpl,
				level: level,
			}
		}

		loggerManager.loggers[name] = logger
	}
	return logger
}

func NewLogManager(etcdClient *commonModel.EtcdClient, env string) (err error) {
	conf := &LogConf{}

	// 尝试从ETCD中读取配置
	if etcdClient != nil {
		err = etcdClient.ParseTomlStruct("/global/log.toml", conf)
		fmt.Println("Logger 尝试读取ETCD配置: /global/log.toml Error:", err)
		if err == nil {
			go watchEtcdConfigUpdated(etcdClient, "/global/log.toml")
		}
	}
	// 尝试从本地读取配置
	if err != nil || etcdClient == nil {
		// 获取当前工作目录
		dir, _ := os.Getwd()

		// 拼接 log.toml 文件路径
		logFilePath := path.Join(dir, "log.toml")

		file, err2 := os.ReadFile(logFilePath)
		if err2 == nil {
			err2 = toml.Unmarshal(file, conf)
		}
		err = err2
		fmt.Println("Logger 尝试读取本地配置文件:", logFilePath, " Error:", err)
		if err == nil {
			go watchFileConfigUpdated(logFilePath)
		}
	}
	// 如果读取配置失败，使用默认配置
	if err != nil {
		fmt.Println("Logger 使用默认配置")
		conf = &LogConf{
			ConsoleAppender: true,
			FileAppender:    false,
			NatsHook:        false,

			Level: make(map[string]string),
		}
	}

	l := log.New()
	l.SetLevel(log.TraceLevel)

	var writers []io.Writer

	// 使用ConsoleAppender模式
	if conf.ConsoleAppender {
		consoleAppender := &ConsoleAppender{}
		writers = append(writers, consoleAppender)
	}

	if conf.NatsHook {
		natsHook, err := NewNatsHook(conf.Nats.Servers, conf.Nats.Username, conf.Nats.Password)
		if err != nil {
			fmt.Println("Failed to create Nats hook:", err)
		} else {
			l.AddHook(natsHook)
		}
	}

	multiWriter := io.MultiWriter(writers...)
	l.SetOutput(multiWriter)
	l.SetFormatter(&CustomFormatter{})

	loggerManager = &LoggerManager{logImpl: l, loggers: make(map[string]*Logger), conf: conf, env: env}
	return nil
}

func watchEtcdConfigUpdated(etcdClient *commonModel.EtcdClient, key string) {
	watchChan := etcdClient.GetClient().Watch(context.Background(), key)
	for watchResp := range watchChan {
		for _, ev := range watchResp.Events {
			switch ev.Type {
			case clientv3.EventTypePut:
				conf := &LogConf{}
				err := toml.Unmarshal(ev.Kv.Value, conf)
				if err == nil {
					for loggerName, levelName := range conf.Level {
						if level, ok := loggerLevelMap[strings.ToLower(levelName)]; ok {
							if logger, ok := loggerManager.loggers[loggerName]; ok {
								logger.SetLevel(level)
							}
						}
					}
				}
			case clientv3.EventTypeDelete:
				// Do nothing
			}
		}
	}
}

// 监听文件变化
func watchFileConfigUpdated(filePath string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	err = watcher.Add(filePath)
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				fmt.Printf("Config file changed: %s\n", event.Name)
				file, err2 := os.ReadFile(filePath)
				if err2 == nil {
					conf := &LogConf{}
					err2 = toml.Unmarshal(file, conf)
					if err2 == nil {
						for loggerName, levelName := range conf.Level {
							if level, ok := loggerLevelMap[strings.ToLower(levelName)]; ok {
								if logger, ok := loggerManager.loggers[loggerName]; ok {
									logger.SetLevel(level)
								}
							}
						}
					}
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("Error:", err)
		}
	}
}

func (l *LoggerManager) GetLoggerLevel(loggerName string, defaultLevel Level) Level {
	if levelName, ok := l.conf.Level[loggerName]; ok {
		if level, ok := loggerLevelMap[strings.ToLower(levelName)]; ok {
			return level
		}
	}
	return defaultLevel
}

// 新增：ConsoleAppender的Write方法实现
func (c *ConsoleAppender) Write(p []byte) (n int, err error) {
	return os.Stdout.Write(p)
}

// Close 修改：ConsoleAppender的Close方法实现
func (c *ConsoleAppender) Close() error {
	// 控制台输出无需关闭，直接返回nil
	return nil
}

// 新增：FileAppender的Write方法实现
func (f *FileAppender) Write(p []byte) (n int, err error) {
	return f.fileWriter.Write(p)
}

// Close 修改：FileAppender的Close方法实现
func (f *FileAppender) Close() error {
	// 关闭文件资源
	return f.fileWriter.Close()
}

func (l *Logger) SetLevel(level Level) {
	l.level = level
}

func (l *Logger) IsLevelEnabled(level Level) bool {
	return l.level >= level
}

func (l *Logger) IsDebugEnabled() bool {
	return l.IsLevelEnabled(DebugLevel)
}

func (l *Logger) IsInfoEnabled() bool {
	return l.IsLevelEnabled(InfoLevel)
}

func (l *Logger) IsWarnEnabled() bool {
	return l.IsLevelEnabled(WarnLevel)
}

func (l *Logger) IsErrorEnabled() bool {
	return l.IsLevelEnabled(ErrorLevel)
}

func (l *Logger) IsLogEnabled() bool {
	return l.IsLevelEnabled(LogLevel)
}

func (l *Logger) Debug(args ...interface{}) {
	if l.level < DebugLevel {
		return
	}
	message := ""
	if len(args) == 0 {
		return
	}
	format, ok := args[0].(string)
	if !ok {
		message = fmt.Sprint(args...)
	} else {
		message = fmt.Sprintf(format, args[1:]...)
	}

	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		file = "unknown"
		line = 0
	}
	fileName := path.Base(file)
	fileName = strings.TrimSuffix(fileName, path.Ext(fileName))
	fullFuncName := runtime.FuncForPC(pc).Name()
	funcName := fullFuncName[strings.LastIndex(fullFuncName, ".")+1:]

	l.impl.WithFields(log.Fields{
		"@logName":  l.name,
		"@fileName": fileName,
		"@funcName": funcName,
		"@line":     line,
		"@level":    loggerLevelStringMap[DebugLevel],
		"@telegram": false,
	}).Debug(message)
}

func (l *Logger) Info(args ...interface{}) {
	if l.level < InfoLevel {
		return
	}
	message := ""
	if len(args) == 0 {
		return
	}
	format, ok := args[0].(string)
	if !ok {
		message = fmt.Sprint(args...)
	} else {
		message = fmt.Sprintf(format, args[1:]...)
	}
	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		file = "unknown"
		line = 0
	}
	fileName := path.Base(file)
	fileName = strings.TrimSuffix(fileName, path.Ext(fileName))
	fullFuncName := runtime.FuncForPC(pc).Name()
	funcName := fullFuncName[strings.LastIndex(fullFuncName, ".")+1:]

	l.impl.WithFields(log.Fields{
		"@logName":  l.name,
		"@fileName": fileName,
		"@funcName": funcName,
		"@line":     line,
		"@level":    loggerLevelStringMap[InfoLevel],
		"@telegram": false,
	}).Info(message)
}

func (l *Logger) Warn(args ...interface{}) {
	if l.level < WarnLevel {
		return
	}
	message := ""
	if len(args) == 0 {
		return
	}
	format, ok := args[0].(string)
	if !ok {
		message = fmt.Sprint(args...)
	} else {
		message = fmt.Sprintf(format, args[1:]...)
	}
	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		file = "unknown"
		line = 0
	}
	fileName := path.Base(file)
	fileName = strings.TrimSuffix(fileName, path.Ext(fileName))
	fullFuncName := runtime.FuncForPC(pc).Name()
	funcName := fullFuncName[strings.LastIndex(fullFuncName, ".")+1:]

	l.impl.WithFields(log.Fields{
		"@logName":  l.name,
		"@fileName": fileName,
		"@funcName": funcName,
		"@line":     line,
		"@level":    loggerLevelStringMap[WarnLevel],
		"@telegram": false,
	}).Warn(message)
}

func (l *Logger) WarnT(args ...interface{}) {
	if l.level < WarnLevel {
		return
	}
	message := ""
	if len(args) == 0 {
		return
	}
	format, ok := args[0].(string)
	if !ok {
		message = fmt.Sprint(args...)
	} else {
		message = fmt.Sprintf(format, args[1:]...)
	}
	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		file = "unknown"
		line = 0
	}
	fileName := path.Base(file)
	fileName = strings.TrimSuffix(fileName, path.Ext(fileName))
	fullFuncName := runtime.FuncForPC(pc).Name()
	funcName := fullFuncName[strings.LastIndex(fullFuncName, ".")+1:]

	l.impl.WithFields(log.Fields{
		"@logName":  l.name,
		"@fileName": fileName,
		"@funcName": funcName,
		"@line":     line,
		"@level":    loggerLevelStringMap[WarnLevel],
		"@telegram": true,
	}).Warn(message)
}

func (l *Logger) Error(args ...interface{}) {
	if l.level < ErrorLevel {
		return
	}
	message := ""
	if len(args) == 0 {
		return
	}
	format, ok := args[0].(string)
	if !ok {
		message = fmt.Sprint(args...)
	} else {
		message = fmt.Sprintf(format, args[1:]...)
	}
	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		file = "unknown"
		line = 0
	}
	fileName := path.Base(file)
	fileName = strings.TrimSuffix(fileName, path.Ext(fileName))
	fullFuncName := runtime.FuncForPC(pc).Name()
	funcName := fullFuncName[strings.LastIndex(fullFuncName, ".")+1:]

	l.impl.WithFields(log.Fields{
		"@logName":  l.name,
		"@fileName": fileName,
		"@funcName": funcName,
		"@line":     line,
		"@level":    loggerLevelStringMap[ErrorLevel],
		"@telegram": false,
	}).Error(message)
}

func (l *Logger) ErrorT(args ...interface{}) {
	if l.level < ErrorLevel {
		return
	}
	message := ""
	if len(args) == 0 {
		return
	}
	format, ok := args[0].(string)
	if !ok {
		message = fmt.Sprint(args...)
	} else {
		message = fmt.Sprintf(format, args[1:]...)
	}
	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		file = "unknown"
		line = 0
	}
	fileName := path.Base(file)
	fileName = strings.TrimSuffix(fileName, path.Ext(fileName))
	fullFuncName := runtime.FuncForPC(pc).Name()
	funcName := fullFuncName[strings.LastIndex(fullFuncName, ".")+1:]

	l.impl.WithFields(log.Fields{
		"@logName":  l.name,
		"@fileName": fileName,
		"@funcName": funcName,
		"@line":     line,
		"@level":    loggerLevelStringMap[ErrorLevel],
		"@telegram": true,
	}).Error(message)
}

func (l *Logger) Log(args ...interface{}) {
	if l.level < LogLevel {
		return
	}
	message := ""
	if len(args) == 0 {
		return
	}
	format, ok := args[0].(string)
	if !ok {
		message = fmt.Sprint(args...)
	} else {
		message = fmt.Sprintf(format, args[1:]...)
	}
	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		file = "unknown"
		line = 0
	}
	fileName := path.Base(file)
	fileName = strings.TrimSuffix(fileName, path.Ext(fileName))
	fullFuncName := runtime.FuncForPC(pc).Name()
	funcName := fullFuncName[strings.LastIndex(fullFuncName, ".")+1:]

	l.impl.WithFields(log.Fields{
		"@logName":  l.name,
		"@fileName": fileName,
		"@funcName": funcName,
		"@line":     line,
		"@level":    loggerLevelStringMap[LogLevel],
		"@telegram": false,
	}).Info(message)
}

func (l *Logger) LogT(args ...interface{}) {
	if l.level < LogLevel {
		return
	}
	message := ""
	if len(args) == 0 {
		return
	}
	format, ok := args[0].(string)
	if !ok {
		message = fmt.Sprint(args...)
	} else {
		message = fmt.Sprintf(format, args[1:]...)
	}
	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		file = "unknown"
		line = 0
	}
	fileName := path.Base(file)
	fileName = strings.TrimSuffix(fileName, path.Ext(fileName))
	fullFuncName := runtime.FuncForPC(pc).Name()
	funcName := fullFuncName[strings.LastIndex(fullFuncName, ".")+1:]

	l.impl.WithFields(log.Fields{
		"@logName":  l.name,
		"@fileName": fileName,
		"@funcName": funcName,
		"@line":     line,
		"@level":    loggerLevelStringMap[LogLevel],
		"@telegram": true,
	}).Info(message)
}
