package repository

import (
	"github.com/KodaTao/drilling/internal/models"
	"gorm.io/gorm"
)

// TunnelRepository 隧道数据仓库接口
type TunnelRepository interface {
	Create(tunnel *models.Tunnel) error
	GetByID(id uint) (*models.Tunnel, error)
	GetByHostID(hostID uint) ([]models.Tunnel, error)
	GetAll() ([]models.Tunnel, error)
	GetByStatus(status string) ([]models.Tunnel, error)
	Update(tunnel *models.Tunnel) error
	Delete(id uint) error
	UpdateStatus(id uint, status string) error
	GetAutoStartTunnels() ([]models.Tunnel, error)
	AddConnectionLog(log *models.ConnectionLog) error
	GetConnectionLogs(tunnelID uint, limit int) ([]models.ConnectionLog, error)
}

// tunnelRepository 隧道数据仓库实现
type tunnelRepository struct {
	db *gorm.DB
}

// NewTunnelRepository 创建隧道数据仓库实例
func NewTunnelRepository(db *gorm.DB) TunnelRepository {
	return &tunnelRepository{db: db}
}

// Create 创建隧道
func (r *tunnelRepository) Create(tunnel *models.Tunnel) error {
	return r.db.Create(tunnel).Error
}

// GetByID 根据ID获取隧道
func (r *tunnelRepository) GetByID(id uint) (*models.Tunnel, error) {
	var tunnel models.Tunnel
	err := r.db.Preload("Host").Preload("ConnectionLogs").First(&tunnel, id).Error
	if err != nil {
		return nil, err
	}
	return &tunnel, nil
}

// GetByHostID 根据主机ID获取隧道
func (r *tunnelRepository) GetByHostID(hostID uint) ([]models.Tunnel, error) {
	var tunnels []models.Tunnel
	err := r.db.Preload("Host").Where("host_id = ?", hostID).Find(&tunnels).Error
	return tunnels, err
}

// GetAll 获取所有隧道
func (r *tunnelRepository) GetAll() ([]models.Tunnel, error) {
	var tunnels []models.Tunnel
	err := r.db.Preload("Host").Find(&tunnels).Error
	return tunnels, err
}

// GetByStatus 根据状态获取隧道
func (r *tunnelRepository) GetByStatus(status string) ([]models.Tunnel, error) {
	var tunnels []models.Tunnel
	err := r.db.Preload("Host").Where("status = ?", status).Find(&tunnels).Error
	return tunnels, err
}

// Update 更新隧道
func (r *tunnelRepository) Update(tunnel *models.Tunnel) error {
	return r.db.Save(tunnel).Error
}

// Delete 删除隧道
func (r *tunnelRepository) Delete(id uint) error {
	// 先删除关联的连接日志
	r.db.Where("tunnel_id = ?", id).Delete(&models.ConnectionLog{})

	// 删除隧道
	return r.db.Delete(&models.Tunnel{}, id).Error
}

// UpdateStatus 更新隧道状态
func (r *tunnelRepository) UpdateStatus(id uint, status string) error {
	return r.db.Model(&models.Tunnel{}).Where("id = ?", id).Update("status", status).Error
}

// GetAutoStartTunnels 获取自动启动的隧道
func (r *tunnelRepository) GetAutoStartTunnels() ([]models.Tunnel, error) {
	var tunnels []models.Tunnel
	err := r.db.Preload("Host").Where("auto_start = ?", true).Find(&tunnels).Error
	return tunnels, err
}

// AddConnectionLog 添加连接日志
func (r *tunnelRepository) AddConnectionLog(log *models.ConnectionLog) error {
	return r.db.Create(log).Error
}

// GetConnectionLogs 获取连接日志
func (r *tunnelRepository) GetConnectionLogs(tunnelID uint, limit int) ([]models.ConnectionLog, error) {
	var logs []models.ConnectionLog
	query := r.db.Where("tunnel_id = ?", tunnelID).Order("timestamp DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Find(&logs).Error
	return logs, err
}