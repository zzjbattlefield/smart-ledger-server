package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config 应用配置
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	AI       AIConfig       `mapstructure:"ai"`
	Log      LogConfig      `mapstructure:"log"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port         int           `mapstructure:"port"`
	Mode         string        `mapstructure:"mode"` // debug, release, test
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	Username        string        `mapstructure:"username"`
	Password        string        `mapstructure:"password"`
	DBName          string        `mapstructure:"dbname"`
	Charset         string        `mapstructure:"charset"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

// DSN 返回数据库连接字符串
func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local&multiStatements=true",
		c.Username,
		c.Password,
		c.Host,
		c.Port,
		c.DBName,
		c.Charset,
	)
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// Addr 返回Redis地址
func (c *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret     string        `mapstructure:"secret"`
	Issuer     string        `mapstructure:"issuer"`
	ExpireTime time.Duration `mapstructure:"expire_time"`
}

// AIConfig AI服务配置
type AIConfig struct {
	APIKey       string      `mapstructure:"api_key"`        // API 密钥
	BaseURL      string      `mapstructure:"base_url"`       // 基础 URL（可选）
	Model        string      `mapstructure:"model"`          // 模型名称
	MaxImageSize int64       `mapstructure:"max_image_size"` // 最大图片大小(字节)
	Batch        BatchConfig `mapstructure:"batch"`          // 批量处理配置
}

// BatchConfig 批量处理配置
type BatchConfig struct {
	MaxImages   int `mapstructure:"max_images"`   // 单次请求最大图片数量
	WorkerCount int `mapstructure:"worker_count"` // Worker并发数
	RPM         int `mapstructure:"rpm"`          // 每分钟最大AI调用次数
	TaskTimeout int `mapstructure:"task_timeout"` // 单个任务超时时间(秒)
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string          `mapstructure:"level"`       // debug, info, warn, error
	Format     string          `mapstructure:"format"`      // json, console
	OutputPath string          `mapstructure:"output_path"` // stdout, file path
	Rotation   LogRotateConfig `mapstructure:"rotation"`    // 日志轮转配置
}

// LogRotateConfig 日志轮转配置
type LogRotateConfig struct {
	MaxSize    int  `mapstructure:"max_size"`    // 单个日志文件最大大小(MB)
	MaxBackups int  `mapstructure:"max_backups"` // 保留的旧日志文件数量
	MaxAge     int  `mapstructure:"max_age"`     // 保留的日志文件最大天数
	Compress   bool `mapstructure:"compress"`    // 是否压缩旧日志文件
}

var globalConfig *Config

// Load 加载配置文件
func Load(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	// 支持环境变量覆盖
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 设置默认值
	setDefaults(&cfg)

	globalConfig = &cfg
	return &cfg, nil
}

// Get 获取全局配置
func Get() *Config {
	return globalConfig
}

// setDefaults 设置默认值
func setDefaults(cfg *Config) {
	// Server defaults
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.Server.Mode == "" {
		cfg.Server.Mode = "debug"
	}
	if cfg.Server.ReadTimeout == 0 {
		cfg.Server.ReadTimeout = 10 * time.Second
	}
	if cfg.Server.WriteTimeout == 0 {
		cfg.Server.WriteTimeout = 10 * time.Second
	}

	// Database defaults
	if cfg.Database.Charset == "" {
		cfg.Database.Charset = "utf8mb4"
	}
	if cfg.Database.MaxIdleConns == 0 {
		cfg.Database.MaxIdleConns = 10
	}
	if cfg.Database.MaxOpenConns == 0 {
		cfg.Database.MaxOpenConns = 100
	}
	if cfg.Database.ConnMaxLifetime == 0 {
		cfg.Database.ConnMaxLifetime = time.Hour
	}

	// JWT defaults
	if cfg.JWT.Issuer == "" {
		cfg.JWT.Issuer = "smart-ledger"
	}
	if cfg.JWT.ExpireTime == 0 {
		cfg.JWT.ExpireTime = 24 * time.Hour
	}

	// AI defaults
	if cfg.AI.Model == "" {
		cfg.AI.Model = "gpt-4o"
	}
	if cfg.AI.MaxImageSize == 0 {
		cfg.AI.MaxImageSize = 10 * 1024 * 1024 // 10MB
	}
	// AI batch defaults
	if cfg.AI.Batch.MaxImages == 0 {
		cfg.AI.Batch.MaxImages = 20
	}
	if cfg.AI.Batch.WorkerCount == 0 {
		cfg.AI.Batch.WorkerCount = 1
	}
	if cfg.AI.Batch.RPM == 0 {
		cfg.AI.Batch.RPM = 60
	}
	if cfg.AI.Batch.TaskTimeout == 0 {
		cfg.AI.Batch.TaskTimeout = 60
	}

	// Log defaults
	if cfg.Log.Level == "" {
		cfg.Log.Level = "info"
	}
	if cfg.Log.Format == "" {
		cfg.Log.Format = "json"
	}
	if cfg.Log.OutputPath == "" {
		cfg.Log.OutputPath = "stdout"
	}
	// Log rotation defaults
	if cfg.Log.Rotation.MaxSize == 0 {
		cfg.Log.Rotation.MaxSize = 100 // 100MB
	}
	if cfg.Log.Rotation.MaxBackups == 0 {
		cfg.Log.Rotation.MaxBackups = 10
	}
	if cfg.Log.Rotation.MaxAge == 0 {
		cfg.Log.Rotation.MaxAge = 30 // 30天
	}
}
