package repository

import (
	"errors"

	"github.com/KodaTao/drilling/internal/models"
	"gorm.io/gorm"
)

// HostRepository 主机数据仓库接口
type HostRepository interface {
	Create(host *models.Host) error
	GetByID(id uint) (*models.Host, error)
	GetByName(name string) (*models.Host, error)
	GetAll() ([]models.Host, error)
	Update(host *models.Host) error
	Delete(id uint) error
	UpdateStatus(id uint, status string) error
}

// hostRepository 主机数据仓库实现
type hostRepository struct {
	db *gorm.DB
}

// NewHostRepository 创建主机数据仓库实例
func NewHostRepository(db *gorm.DB) HostRepository {
	return &hostRepository{db: db}
}

// Create 创建主机
func (r *hostRepository) Create(host *models.Host) error {
	return r.db.Create(host).Error
}

// GetByID 根据ID获取主机
func (r *hostRepository) GetByID(id uint) (*models.Host, error) {
	var host models.Host
	err := r.db.Preload("Tunnels").First(&host, id).Error
	if err != nil {
		return nil, err
	}
	return &host, nil
}

// GetByName 根据名称获取主机
func (r *hostRepository) GetByName(name string) (*models.Host, error) {
	var host models.Host
	err := r.db.Where("name = ?", name).First(&host).Error
	if err != nil {
		return nil, err
	}
	return &host, nil
}

// GetAll 获取所有主机
func (r *hostRepository) GetAll() ([]models.Host, error) {
	var hosts []models.Host
	err := r.db.Preload("Tunnels").Find(&hosts).Error
	return hosts, err
}

// Update 更新主机
func (r *hostRepository) Update(host *models.Host) error {
	return r.db.Save(host).Error
}

// Delete 删除主机
func (r *hostRepository) Delete(id uint) error {
	// 检查是否有关联的隧道
	var count int64
	r.db.Model(&models.Tunnel{}).Where("host_id = ?", id).Count(&count)
	if count > 0 {
		return errors.New("cannot delete host with active tunnels")
	}

	return r.db.Unscoped().Delete(&models.Host{}, id).Error
}

// UpdateStatus 更新主机状态
func (r *hostRepository) UpdateStatus(id uint, status string) error {
	return r.db.Model(&models.Host{}).Where("id = ?", id).Update("status", status).Error
}
