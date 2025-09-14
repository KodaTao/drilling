package config

import (
	"log"
	"strings"

	"github.com/spf13/viper"
)

// Config 应用配置结构
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	SSH      SSHConfig      `mapstructure:"ssh"`
	Logging  LoggingConfig  `mapstructure:"logging"`
	Security SecurityConfig `mapstructure:"security"`
	Debug    bool           `mapstructure:"debug"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port string `mapstructure:"port"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Path string `mapstructure:"path"`
}

// SSHConfig SSH配置
type SSHConfig struct {
	Timeout        string `mapstructure:"timeout"`
	Keepalive      string `mapstructure:"keepalive"`
	MaxConnections int    `mapstructure:"max_connections"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level string `mapstructure:"level"`
	File  string `mapstructure:"file"`
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	EncryptKey string `mapstructure:"encrypt_key"`
}

// Load 加载配置
func Load() *Config {
	// 设置默认配置
	setDefaults()

	// 设置配置文件名和路径
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./configs")

	// 支持环境变量
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("No config file found, using defaults and environment variables")
		} else {
			log.Printf("Error reading config file: %v", err)
		}
	} else {
		log.Printf("Using config file: %s", viper.ConfigFileUsed())
	}

	// 解析配置到结构体
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("Unable to decode config: %v", err)
	}

	return &config
}

// setDefaults 设置默认配置值
func setDefaults() {
	// 服务器默认配置
	viper.SetDefault("server.host", "127.0.0.1")
	viper.SetDefault("server.port", "8080")

	// 数据库默认配置
	viper.SetDefault("database.path", "./drilling.db")

	// SSH默认配置
	viper.SetDefault("ssh.timeout", "30s")
	viper.SetDefault("ssh.keepalive", "10s")
	viper.SetDefault("ssh.max_connections", 100)

	// 日志默认配置
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.file", "./drilling.log")

	// 安全配置
	viper.SetDefault("security.encrypt_key", "default-encryption-key-change-in-production")

	// 调试模式
	viper.SetDefault("debug", false)
}