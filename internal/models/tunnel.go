package models

import (
	"time"

	"gorm.io/gorm"
)

// Tunnel 隧道模型
type Tunnel struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	HostID        uint           `json:"host_id" gorm:"not null" binding:"required"`
	Name          string         `json:"name" gorm:"not null" binding:"required"`
	Type          string         `json:"type" gorm:"not null" binding:"required,oneof=local_forward remote_forward dynamic"`
	LocalAddress  string         `json:"local_address" gorm:"default:127.0.0.1"`
	LocalPort     int            `json:"local_port" gorm:"not null" binding:"required"`
	RemoteAddress string         `json:"remote_address"`
	RemotePort    int            `json:"remote_port"`
	Description   string         `json:"description"`
	Status        string         `json:"status" gorm:"default:inactive"`
	AutoStart     bool           `json:"auto_start" gorm:"default:false"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联的主机
	Host *Host `json:"host,omitempty" gorm:"foreignKey:HostID"`

	// 关联的连接日志
	ConnectionLogs []ConnectionLog `json:"connection_logs,omitempty" gorm:"foreignKey:TunnelID"`
}

// TableName 指定表名
func (Tunnel) TableName() string {
	return "tunnels"
}

// ConnectionLog 连接日志模型
type ConnectionLog struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	TunnelID  uint      `json:"tunnel_id" gorm:"not null"`
	EventType string    `json:"event_type" gorm:"not null"` // connect, disconnect, error
	Message   string    `json:"message" gorm:"type:text"`
	Timestamp time.Time `json:"timestamp" gorm:"default:CURRENT_TIMESTAMP"`

	// 关联的隧道
	Tunnel *Tunnel `json:"tunnel,omitempty" gorm:"foreignKey:TunnelID"`
}

// TableName 指定表名
func (ConnectionLog) TableName() string {
	return "connection_logs"
}

// TunnelType 隧道类型常量
const (
	TunnelTypeLocalForward  = "local_forward"  // 本地端口转发（远程服务映射到本地）
	TunnelTypeRemoteForward = "remote_forward" // 远程端口转发（本地服务映射到远程）
	TunnelTypeDynamic       = "dynamic"        // 动态端口转发（SOCKS5代理）
)

// TunnelStatus 隧道状态常量
const (
	TunnelStatusActive   = "active"
	TunnelStatusInactive = "inactive"
	TunnelStatusError    = "error"
)

// LogEventType 日志事件类型常量
const (
	LogEventConnect    = "connect"
	LogEventDisconnect = "disconnect"
	LogEventError      = "error"
	LogEventStart      = "start"
	LogEventStop       = "stop"
)

// TrafficStats 流量统计模型
type TrafficStats struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	TunnelID    uint      `json:"tunnel_id" gorm:"not null;index"`
	BytesIn     int64     `json:"bytes_in" gorm:"default:0"`     // 入站字节数
	BytesOut    int64     `json:"bytes_out" gorm:"default:0"`    // 出站字节数
	Connections int64     `json:"connections" gorm:"default:0"`  // 连接数
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// 关联的隧道
	Tunnel *Tunnel `json:"tunnel,omitempty" gorm:"foreignKey:TunnelID"`
}

// TableName 指定表名
func (TrafficStats) TableName() string {
	return "traffic_stats"
}

// RealtimeTrafficStats 实时流量统计
type RealtimeTrafficStats struct {
	TunnelID         uint    `json:"tunnel_id"`
	CurrentBytesIn   int64   `json:"current_bytes_in"`
	CurrentBytesOut  int64   `json:"current_bytes_out"`
	ActiveConnections int    `json:"active_connections"`
	SpeedIn          float64 `json:"speed_in"`          // bytes/second
	SpeedOut         float64 `json:"speed_out"`         // bytes/second
	LastUpdateTime   time.Time `json:"last_update_time"`
}