package logs

import (
	"fmt"
	commonModel "github.com/bwgame666/common/model"
	"github.com/pelletier/go-toml"
	log "github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"os"
	"path"
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
		}

		level := loggerManager.GetLoggerLevel(name, defaultLevel)

		logger = &Logger{
			name:  name,
			impl:  loggerManager.logImpl,
			level: level,
		}
		loggerManager.loggers[name] = logger
	}
	return logger
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

func NewLogManager(etcdClient *commonModel.EtcdClient, env string) (err error) {
	conf := &LogConf{}

	// 尝试从ETCD中读取配置
	if etcdClient != nil {
		err = etcdClient.ParseTomlStruct("/global/log.toml", conf)
		fmt.Println("Logger 尝试读取ETCD配置: /global/log.toml Error:", err)
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
	}
	// 如果读取配置失败，使用默认配置
	if err != nil {
		fmt.Println("Logger 使用默认配置")
		conf = &LogConf{
			ConsoleAppender: true,
			FileAppender:    true,
			File: LogFileConf{
				FilePath:   "./logs/app.log",
				MaxSize:    500,
				MaxBackups: 9,
				MaxAge:     30,
				Compress:   true,
			},
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

	// 使用FileAppender模式
	if conf.FileAppender {
		fileAppender := &FileAppender{
			fileWriter: &lumberjack.Logger{
				Filename:   conf.File.FilePath,
				MaxSize:    conf.File.MaxSize,
				MaxBackups: conf.File.MaxBackups,
				MaxAge:     conf.File.MaxAge,
				Compress:   conf.File.Compress,
			},
		}
		writers = append(writers, fileAppender)
	}

	multiWriter := io.MultiWriter(writers...)
	l.SetOutput(multiWriter)
	l.SetFormatter(&CustomFormatter{})

	loggerManager = &LoggerManager{logImpl: l, loggers: make(map[string]*Logger), conf: conf, env: env}
	return nil
}

func (l *LoggerManager) GetLoggerLevel(loggerName string, defaultLevel Level) Level {
	if levelName, ok := l.conf.Level[strings.ToLower(loggerName)]; ok {
		if level, ok := loggerLevelMap[strings.ToLower(levelName)]; ok {
			return level
		}
	}
	return defaultLevel
}

func (l *Logger) SetLevel(level Level) {
	l.level = level
}

func (l *Logger) Debug(args ...interface{}) {
	if l.level < DebugLevel {
		return
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
	}).Debug(args)
}

func (l *Logger) Info(args ...interface{}) {
	if l.level < InfoLevel {
		return
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
	}).Info(args)
}

func (l *Logger) Warn(args ...interface{}) {
	if l.level < WarnLevel {
		return
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
	}).Warn(args)
}

func (l *Logger) Error(args ...interface{}) {
	if l.level < ErrorLevel {
		return
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
	}).Error(args)
}

func (l *Logger) Log(args ...interface{}) {
	if l.level < LogLevel {
		return
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
	}).Info(args)
}
