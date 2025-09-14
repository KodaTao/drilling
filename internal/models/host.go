package models

import (
	"time"

	"gorm.io/gorm"
)

// Host 主机模型
type Host struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"uniqueIndex;not null" binding:"required"`
	Hostname    string         `json:"hostname" gorm:"not null" binding:"required"`
	Port        int            `json:"port" gorm:"default:22"`
	Username    string         `json:"username" gorm:"not null" binding:"required"`
	AuthType    string         `json:"auth_type" gorm:"not null" binding:"required,oneof=password key key_password"`
	Password    string         `json:"password,omitempty" gorm:"type:text"`                    // 加密存储
	PrivateKey  string         `json:"private_key,omitempty" gorm:"type:text"`                 // 私钥内容，加密存储
	KeyPath     string         `json:"key_path,omitempty"`                                     // 私钥文件路径
	Passphrase  string         `json:"passphrase,omitempty" gorm:"type:text"`                  // 私钥密码，加密存储
	Description string         `json:"description"`                                            // 描述
	Status      string         `json:"status" gorm:"default:inactive"`                        // active, inactive, error
	LastCheck   *time.Time     `json:"last_check"`                                             // 最后检查时间
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联的隧道
	Tunnels []Tunnel `json:"tunnels,omitempty" gorm:"foreignKey:HostID"`
}

// TableName 指定表名
func (Host) TableName() string {
	return "hosts"
}

// AuthTypes 认证类型常量
const (
	AuthTypePassword    = "password"
	AuthTypeKey         = "key"
	AuthTypeKeyPassword = "key_password"
)

// HostStatus 主机状态常量
const (
	HostStatusActive   = "active"
	HostStatusInactive = "inactive"
	HostStatusError    = "error"
)