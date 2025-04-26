package logs

type LogConf struct {
	ConsoleAppender bool `toml:"console_appender"`

	FileAppender bool        `toml:"file_appender"`
	File         LogFileConf `toml:"file"`

	NatsHook bool     `toml:"nats_hook"`
	Nats     NatsConf `toml:"nats"`

	Level map[string]string `toml:"level"`
}

type LogFileConf struct {
	FilePath   string `toml:"file_path"`   // 日志文件路径
	MaxSize    int    `toml:"max_size"`    // 日志文件最大大小（MB）
	MaxBackups int    `toml:"max_backups"` // 最多保留的旧日志文件数量
	MaxAge     int    `toml:"max_age"`     // 保留旧日志文件的最大天数
	Compress   bool   `toml:"compress"`    // 是否压缩
}

type NatsConf struct {
	Servers  []string `toml:"servers"`
	Username string   `toml:"username"`
	Password string   `toml:"password"`
}
