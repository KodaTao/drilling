package service

import (
	"sync"
	"time"

	"github.com/KodaTao/drilling/internal/models"
	"github.com/KodaTao/drilling/internal/repository"
)

// TrafficService 流量统计服务接口
type TrafficService interface {
	LogTraffic(tunnelID uint, bytesIn, bytesOut int64)
	GetTrafficStats(tunnelID uint, startTime, endTime time.Time) ([]models.TrafficStats, error)
	GetRealtimeStats(tunnelID uint) (*models.RealtimeTrafficStats, error)
	GetAllRealtimeStats() (map[uint]*models.RealtimeTrafficStats, error)
}

// trafficService 流量统计服务实现
type trafficService struct {
	tunnelRepo    repository.TunnelRepository
	realtimeStats map[uint]*models.RealtimeTrafficStats
	mutex         sync.RWMutex
}

// NewTrafficService 创建流量统计服务实例
func NewTrafficService(tunnelRepo repository.TunnelRepository) TrafficService {
	return &trafficService{
		tunnelRepo:    tunnelRepo,
		realtimeStats: make(map[uint]*models.RealtimeTrafficStats),
	}
}

// LogTraffic 记录流量统计
func (s *trafficService) LogTraffic(tunnelID uint, bytesIn, bytesOut int64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	stats, exists := s.realtimeStats[tunnelID]
	if !exists {
		stats = &models.RealtimeTrafficStats{
			TunnelID:         tunnelID,
			CurrentBytesIn:   0,
			CurrentBytesOut:  0,
			ActiveConnections: 1,
			LastUpdateTime:   time.Now(),
		}
		s.realtimeStats[tunnelID] = stats
	}

	// 更新累计流量
	stats.CurrentBytesIn += bytesIn
	stats.CurrentBytesOut += bytesOut

	// 计算传输速度（简化实现，基于最近一次更新的时间差）
	now := time.Now()
	timeDiff := now.Sub(stats.LastUpdateTime).Seconds()
	if timeDiff > 0 {
		stats.SpeedIn = float64(bytesIn) / timeDiff
		stats.SpeedOut = float64(bytesOut) / timeDiff
	}

	stats.LastUpdateTime = now
}

// GetTrafficStats 获取历史流量统计
func (s *trafficService) GetTrafficStats(tunnelID uint, startTime, endTime time.Time) ([]models.TrafficStats, error) {
	// 这里应该实现从数据库获取历史统计数据
	// 由于当前没有实现TrafficStats的数据库操作，返回空数组
	return []models.TrafficStats{}, nil
}

// GetRealtimeStats 获取实时流量统计
func (s *trafficService) GetRealtimeStats(tunnelID uint) (*models.RealtimeTrafficStats, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	stats, exists := s.realtimeStats[tunnelID]
	if !exists {
		// 返回空统计
		return &models.RealtimeTrafficStats{
			TunnelID:         tunnelID,
			CurrentBytesIn:   0,
			CurrentBytesOut:  0,
			ActiveConnections: 0,
			SpeedIn:          0,
			SpeedOut:         0,
			LastUpdateTime:   time.Now(),
		}, nil
	}

	// 返回统计副本
	statsCopy := *stats
	return &statsCopy, nil
}

// GetAllRealtimeStats 获取所有实时流量统计
func (s *trafficService) GetAllRealtimeStats() (map[uint]*models.RealtimeTrafficStats, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// 创建副本
	result := make(map[uint]*models.RealtimeTrafficStats)
	for tunnelID, stats := range s.realtimeStats {
		statsCopy := *stats
		result[tunnelID] = &statsCopy
	}

	return result, nil
}

// IncrementConnection 增加连接数
func (s *trafficService) IncrementConnection(tunnelID uint) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	stats, exists := s.realtimeStats[tunnelID]
	if !exists {
		stats = &models.RealtimeTrafficStats{
			TunnelID:         tunnelID,
			CurrentBytesIn:   0,
			CurrentBytesOut:  0,
			ActiveConnections: 0,
			LastUpdateTime:   time.Now(),
		}
		s.realtimeStats[tunnelID] = stats
	}

	stats.ActiveConnections++
}

// DecrementConnection 减少连接数
func (s *trafficService) DecrementConnection(tunnelID uint) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	stats, exists := s.realtimeStats[tunnelID]
	if exists && stats.ActiveConnections > 0 {
		stats.ActiveConnections--
	}
}

// TunnelTrafficLogger 隧道流量记录器
type TunnelTrafficLogger struct {
	tunnelID       uint
	trafficService TrafficService
}

// NewTunnelTrafficLogger 创建隧道流量记录器
func NewTunnelTrafficLogger(tunnelID uint, trafficService TrafficService) *TunnelTrafficLogger {
	return &TunnelTrafficLogger{
		tunnelID:       tunnelID,
		trafficService: trafficService,
	}
}

// LogTraffic 记录流量
func (l *TunnelTrafficLogger) LogTraffic(bytesIn, bytesOut int64) {
	l.trafficService.LogTraffic(l.tunnelID, bytesIn, bytesOut)
}