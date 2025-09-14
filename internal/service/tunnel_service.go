package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/KodaTao/drilling/internal/models"
	"github.com/KodaTao/drilling/internal/repository"
	"github.com/KodaTao/drilling/internal/socks5"
	"golang.org/x/crypto/ssh"
)

// TunnelService 隧道服务接口
type TunnelService interface {
	CreateTunnel(tunnel *models.Tunnel) error
	CreateMultipleLocalForwards(hostID uint, localServices []LocalServiceMapping) ([]models.Tunnel, error)
	CreateDynamicSOCKS5Tunnel(hostID uint, name, description string, autoStart bool) (*models.Tunnel, error)
	FindAvailablePort(startPort, endPort int, address string) (int, error)
	GetTunnel(id uint) (*models.Tunnel, error)
	GetAllTunnels() ([]models.Tunnel, error)
	GetTunnelsByHost(hostID uint) ([]models.Tunnel, error)
	UpdateTunnel(tunnel *models.Tunnel) error
	DeleteTunnel(id uint) error
	StartTunnel(id uint) error
	StopTunnel(id uint) error
	RestartTunnel(id uint) error
	GetTunnelStatus(id uint) (string, error)
	StartAutoTunnels() error
	StopAllTunnels() error
	GetConnectionLogs(tunnelID uint, limit int) ([]models.ConnectionLog, error)
	CheckServiceHealth(localAddress string, localPort int) error
}

// tunnelService 隧道服务实现
type tunnelService struct {
	tunnelRepo     repository.TunnelRepository
	hostService    HostService
	trafficService TrafficService
	activeTunnels  map[uint]*activeTunnel
	mutex          sync.RWMutex
}

// activeTunnel 活动隧道信息
type activeTunnel struct {
	tunnel     *models.Tunnel
	sshClient  *ssh.Client
	listener   net.Listener
	ctx        context.Context
	cancel     context.CancelFunc
	startTime  time.Time
}

// NewTunnelService 创建隧道服务实例
func NewTunnelService(tunnelRepo repository.TunnelRepository, hostService HostService) TunnelService {
	trafficService := NewTrafficService(tunnelRepo)
	return &tunnelService{
		tunnelRepo:     tunnelRepo,
		hostService:    hostService,
		trafficService: trafficService,
		activeTunnels:  make(map[uint]*activeTunnel),
	}
}

// CreateTunnel 创建隧道
func (s *tunnelService) CreateTunnel(tunnel *models.Tunnel) error {
	// 验证隧道配置
	if err := s.validateTunnelConfig(tunnel); err != nil {
		return err
	}

	// 检查端口是否已被占用
	if err := s.checkPortAvailability(tunnel); err != nil {
		return err
	}

	// 设置默认状态
	tunnel.Status = models.TunnelStatusInactive

	return s.tunnelRepo.Create(tunnel)
}

// CreateMultipleLocalForwards 创建多个本地服务的远程端口映射
func (s *tunnelService) CreateMultipleLocalForwards(hostID uint, localServices []LocalServiceMapping) ([]models.Tunnel, error) {
	var createdTunnels []models.Tunnel
	var errors []string

	for _, service := range localServices {
		tunnel := &models.Tunnel{
			Name:          service.Name,
			Type:          models.TunnelTypeRemoteForward,
			HostID:        hostID,
			LocalAddress:  service.LocalAddress,
			LocalPort:     service.LocalPort,
			RemoteAddress: service.RemoteAddress,
			RemotePort:    service.RemotePort,
			AutoStart:     service.AutoStart,
			Description:   service.Description,
		}

		if err := s.CreateTunnel(tunnel); err != nil {
			errors = append(errors, fmt.Sprintf("Failed to create tunnel for %s:%d - %v", service.LocalAddress, service.LocalPort, err))
			continue
		}

		createdTunnels = append(createdTunnels, *tunnel)
	}

	if len(errors) > 0 {
		return createdTunnels, fmt.Errorf("some tunnels failed to create: %s", strings.Join(errors, "; "))
	}

	return createdTunnels, nil
}

// LocalServiceMapping 本地服务映射配置
type LocalServiceMapping struct {
	Name          string `json:"name"`
	LocalAddress  string `json:"local_address"`
	LocalPort     int    `json:"local_port"`
	RemoteAddress string `json:"remote_address"`
	RemotePort    int    `json:"remote_port"`
	AutoStart     bool   `json:"auto_start"`
	Description   string `json:"description,omitempty"`
}

// GetTunnel 获取隧道
func (s *tunnelService) GetTunnel(id uint) (*models.Tunnel, error) {
	return s.tunnelRepo.GetByID(id)
}

// GetAllTunnels 获取所有隧道
func (s *tunnelService) GetAllTunnels() ([]models.Tunnel, error) {
	return s.tunnelRepo.GetAll()
}

// GetTunnelsByHost 根据主机获取隧道
func (s *tunnelService) GetTunnelsByHost(hostID uint) ([]models.Tunnel, error) {
	return s.tunnelRepo.GetByHostID(hostID)
}

// UpdateTunnel 更新隧道
func (s *tunnelService) UpdateTunnel(tunnel *models.Tunnel) error {
	// 验证隧道配置
	if err := s.validateTunnelConfig(tunnel); err != nil {
		return err
	}

	// 如果隧道正在运行，需要重启
	s.mutex.RLock()
	isActive := s.activeTunnels[tunnel.ID] != nil
	s.mutex.RUnlock()

	if isActive {
		// 停止隧道
		if err := s.StopTunnel(tunnel.ID); err != nil {
			return fmt.Errorf("failed to stop tunnel for update: %v", err)
		}

		// 更新隧道
		if err := s.tunnelRepo.Update(tunnel); err != nil {
			return err
		}

		// 重新启动隧道
		return s.StartTunnel(tunnel.ID)
	}

	return s.tunnelRepo.Update(tunnel)
}

// DeleteTunnel 删除隧道
func (s *tunnelService) DeleteTunnel(id uint) error {
	// 先停止隧道
	s.StopTunnel(id)

	return s.tunnelRepo.Delete(id)
}

// StartTunnel 启动隧道
func (s *tunnelService) StartTunnel(id uint) error {
	tunnel, err := s.tunnelRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get tunnel: %v", err)
	}

	// 检查隧道是否已经在运行
	s.mutex.RLock()
	if s.activeTunnels[id] != nil {
		s.mutex.RUnlock()
		return errors.New("tunnel is already running")
	}
	s.mutex.RUnlock()

	// 获取主机信息
	host, err := s.hostService.GetHost(tunnel.HostID)
	if err != nil {
		return fmt.Errorf("failed to get host: %v", err)
	}

	// 创建SSH连接
	sshClient, err := s.createSSHConnection(host)
	if err != nil {
		s.addConnectionLog(tunnel.ID, models.LogEventError, fmt.Sprintf("SSH connection failed: %v", err))
		s.tunnelRepo.UpdateStatus(id, models.TunnelStatusError)
		return fmt.Errorf("failed to create SSH connection: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	// 根据隧道类型启动相应的转发
	switch tunnel.Type {
	case models.TunnelTypeLocalForward:
		err = s.startLocalForward(ctx, tunnel, sshClient)
	case models.TunnelTypeRemoteForward:
		err = s.startRemoteForward(ctx, tunnel, sshClient)
	case models.TunnelTypeDynamic:
		err = s.startDynamicForward(ctx, tunnel, sshClient)
	default:
		cancel()
		sshClient.Close()
		return errors.New("unsupported tunnel type")
	}

	if err != nil {
		cancel()
		sshClient.Close()
		s.addConnectionLog(tunnel.ID, models.LogEventError, fmt.Sprintf("Failed to start tunnel: %v", err))
		s.tunnelRepo.UpdateStatus(id, models.TunnelStatusError)
		return fmt.Errorf("failed to start tunnel: %v", err)
	}

	// 保存活动隧道信息
	activeTunnel := &activeTunnel{
		tunnel:    tunnel,
		sshClient: sshClient,
		ctx:       ctx,
		cancel:    cancel,
		startTime: time.Now(),
	}

	s.mutex.Lock()
	s.activeTunnels[id] = activeTunnel
	s.mutex.Unlock()

	// 更新隧道状态
	s.tunnelRepo.UpdateStatus(id, models.TunnelStatusActive)
	s.addConnectionLog(tunnel.ID, models.LogEventStart, "Tunnel started successfully")

	return nil
}

// startLocalForward 启动本地转发（远程服务映射到本地）
func (s *tunnelService) startLocalForward(ctx context.Context, tunnel *models.Tunnel, sshClient *ssh.Client) error {
	// 监听本地端口
	localAddr := fmt.Sprintf("%s:%d", tunnel.LocalAddress, tunnel.LocalPort)
	listener, err := net.Listen("tcp", localAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %v", localAddr, err)
	}

	// 保存listener到activeTunnel（需要在调用处设置）
	s.mutex.Lock()
	if activeTunnel := s.activeTunnels[tunnel.ID]; activeTunnel != nil {
		activeTunnel.listener = listener
	}
	s.mutex.Unlock()

	go func() {
		defer func() {
			log.Printf("Local forward goroutine exiting for tunnel %d", tunnel.ID)
			listener.Close()
		}()

		for {
			select {
			case <-ctx.Done():
				log.Printf("Context cancelled for tunnel %d local forward", tunnel.ID)
				return
			default:
				// 设置accept的超时，这样可以定期检查context
				if tcpListener, ok := listener.(*net.TCPListener); ok {
					tcpListener.SetDeadline(time.Now().Add(1 * time.Second))
				}

				// 接受本地连接
				localConn, err := listener.Accept()
				if err != nil {
					// 检查是否是超时错误
					if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
						// 超时是正常的，继续循环检查context
						continue
					}
					if ctx.Err() != nil {
						return // 上下文已取消
					}
					log.Printf("Failed to accept local connection for tunnel %d: %v", tunnel.ID, err)
					continue
				}

				// 清除deadline
				if tcpListener, ok := listener.(*net.TCPListener); ok {
					tcpListener.SetDeadline(time.Time{})
				}

				// 异步处理连接
				go s.handleLocalForward(ctx, tunnel, sshClient, localConn)
			}
		}
	}()

	return nil
}

// handleLocalForward 处理本地转发连接
func (s *tunnelService) handleLocalForward(ctx context.Context, tunnel *models.Tunnel, sshClient *ssh.Client, localConn net.Conn) {
	defer localConn.Close()

	// 连接远程地址
	remoteAddr := fmt.Sprintf("%s:%d", tunnel.RemoteAddress, tunnel.RemotePort)
	remoteConn, err := sshClient.Dial("tcp", remoteAddr)
	if err != nil {
		log.Printf("Failed to dial remote address %s for tunnel %d: %v", remoteAddr, tunnel.ID, err)
		s.addConnectionLog(tunnel.ID, models.LogEventError, fmt.Sprintf("Failed to connect to %s: %v", remoteAddr, err))
		return
	}
	defer remoteConn.Close()

	s.addConnectionLog(tunnel.ID, models.LogEventConnect, fmt.Sprintf("Connection established: %s -> %s", localConn.RemoteAddr(), remoteAddr))

	// 双向数据转发
	done := make(chan struct{}, 2)

	// 本地 -> 远程
	go func() {
		defer func() { done <- struct{}{} }()
		io.Copy(remoteConn, localConn)
	}()

	// 远程 -> 本地
	go func() {
		defer func() { done <- struct{}{} }()
		io.Copy(localConn, remoteConn)
	}()

	// 等待任一方向完成或上下文取消
	select {
	case <-done:
	case <-ctx.Done():
	}

	s.addConnectionLog(tunnel.ID, models.LogEventDisconnect, "Connection closed")
}

// startRemoteForward 启动远程转发（本地服务映射到远程）
func (s *tunnelService) startRemoteForward(ctx context.Context, tunnel *models.Tunnel, sshClient *ssh.Client) error {
	// 在远程主机上监听端口
	remoteAddr := fmt.Sprintf("%s:%d", tunnel.RemoteAddress, tunnel.RemotePort)
	listener, err := sshClient.Listen("tcp", remoteAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on remote %s: %v", remoteAddr, err)
	}

	s.mutex.Lock()
	if activeTunnel := s.activeTunnels[tunnel.ID]; activeTunnel != nil {
		activeTunnel.listener = listener
	}
	s.mutex.Unlock()

	go func() {
		defer func() {
			log.Printf("Remote forward goroutine exiting for tunnel %d", tunnel.ID)
			listener.Close()
		}()

		for {
			select {
			case <-ctx.Done():
				log.Printf("Context cancelled for tunnel %d remote forward", tunnel.ID)
				return
			default:
				// 接受远程连接
				remoteConn, err := listener.Accept()
				if err != nil {
					if ctx.Err() != nil {
						return
					}
					log.Printf("Failed to accept remote connection for tunnel %d: %v", tunnel.ID, err)
					continue
				}

				// 异步处理连接
				go s.handleRemoteForward(ctx, tunnel, remoteConn)
			}
		}
	}()

	return nil
}

// handleRemoteForward 处理远程转发连接
func (s *tunnelService) handleRemoteForward(ctx context.Context, tunnel *models.Tunnel, remoteConn net.Conn) {
	defer remoteConn.Close()

	// 连接本地地址
	localAddr := fmt.Sprintf("%s:%d", tunnel.LocalAddress, tunnel.LocalPort)
	localConn, err := net.Dial("tcp", localAddr)
	if err != nil {
		log.Printf("Failed to dial local address %s for tunnel %d: %v", localAddr, tunnel.ID, err)
		s.addConnectionLog(tunnel.ID, models.LogEventError, fmt.Sprintf("Failed to connect to %s: %v", localAddr, err))
		return
	}
	defer localConn.Close()

	s.addConnectionLog(tunnel.ID, models.LogEventConnect, fmt.Sprintf("Connection established: %s -> %s", remoteConn.RemoteAddr(), localAddr))

	// 双向数据转发
	done := make(chan struct{}, 2)

	// 远程 -> 本地
	go func() {
		defer func() { done <- struct{}{} }()
		io.Copy(localConn, remoteConn)
	}()

	// 本地 -> 远程
	go func() {
		defer func() { done <- struct{}{} }()
		io.Copy(remoteConn, localConn)
	}()

	select {
	case <-done:
	case <-ctx.Done():
	}

	s.addConnectionLog(tunnel.ID, models.LogEventDisconnect, "Connection closed")
}

// startDynamicForward 启动动态转发（SOCKS5代理）
func (s *tunnelService) startDynamicForward(ctx context.Context, tunnel *models.Tunnel, sshClient *ssh.Client) error {
	// 监听本地端口作为SOCKS5代理
	localAddr := fmt.Sprintf("%s:%d", tunnel.LocalAddress, tunnel.LocalPort)
	listener, err := net.Listen("tcp", localAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %v", localAddr, err)
	}

	s.mutex.Lock()
	if activeTunnel := s.activeTunnels[tunnel.ID]; activeTunnel != nil {
		activeTunnel.listener = listener
	}
	s.mutex.Unlock()

	go func() {
		defer func() {
			log.Printf("SOCKS5 goroutine exiting for tunnel %d", tunnel.ID)
			listener.Close()
		}()

		for {
			select {
			case <-ctx.Done():
				log.Printf("Context cancelled for tunnel %d SOCKS5", tunnel.ID)
				return
			default:
				// 设置accept的超时，这样可以定期检查context
				if tcpListener, ok := listener.(*net.TCPListener); ok {
					tcpListener.SetDeadline(time.Now().Add(1 * time.Second))
				}

				conn, err := listener.Accept()
				if err != nil {
					// 检查是否是超时错误
					if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
						// 超时是正常的，继续循环检查context
						continue
					}
					if ctx.Err() != nil {
						return
					}
					log.Printf("Failed to accept SOCKS connection for tunnel %d: %v", tunnel.ID, err)
					continue
				}

				// 清除deadline
				if tcpListener, ok := listener.(*net.TCPListener); ok {
					tcpListener.SetDeadline(time.Time{})
				}

				// 异步处理SOCKS5连接
				go s.handleSOCKS5(ctx, tunnel, sshClient, conn)
			}
		}
	}()

	return nil
}

// handleSOCKS5 处理SOCKS5连接
func (s *tunnelService) handleSOCKS5(ctx context.Context, tunnel *models.Tunnel, sshClient *ssh.Client, conn net.Conn) {
	defer conn.Close()

	s.addConnectionLog(tunnel.ID, models.LogEventConnect, fmt.Sprintf("SOCKS5 connection from %s", conn.RemoteAddr()))

	// 增加连接计数
	if ts, ok := s.trafficService.(*trafficService); ok {
		ts.IncrementConnection(tunnel.ID)
		defer ts.DecrementConnection(tunnel.ID)
	}

	// 创建流量记录器
	trafficLogger := NewTunnelTrafficLogger(tunnel.ID, s.trafficService)

	// 创建带流量统计的SOCKS5服务器实例
	socksServer := socks5.NewSOCKS5ServerWithTrafficLogger(sshClient, trafficLogger)

	// 处理SOCKS5连接
	if err := socksServer.HandleConnection(ctx, conn); err != nil {
		s.addConnectionLog(tunnel.ID, models.LogEventError, fmt.Sprintf("SOCKS5 connection error: %v", err))
		log.Printf("SOCKS5 connection error for tunnel %d: %v", tunnel.ID, err)
	}

	s.addConnectionLog(tunnel.ID, models.LogEventDisconnect, "SOCKS5 connection closed")
}

// StopTunnel 停止隧道
func (s *tunnelService) StopTunnel(id uint) error {
	s.mutex.Lock()
	activeTunnel := s.activeTunnels[id]
	if activeTunnel != nil {
		delete(s.activeTunnels, id)
	}
	s.mutex.Unlock()

	if activeTunnel == nil {
		return errors.New("tunnel is not running")
	}

	log.Printf("Stopping tunnel %d", id)

	// 第一步：立即关闭监听器以释放端口
	if activeTunnel.listener != nil {
		log.Printf("Closing listener for tunnel %d", id)
		if err := activeTunnel.listener.Close(); err != nil {
			// listener.Close() 可能被多次调用，忽略已关闭的错误
			if !strings.Contains(err.Error(), "use of closed network connection") {
				log.Printf("Error closing listener for tunnel %d: %v", id, err)
			}
		} else {
			log.Printf("Successfully closed listener for tunnel %d", id)
		}
	}

	// 第二步：取消上下文，通知goroutines退出
	activeTunnel.cancel()

	// 第三步：关闭SSH连接
	if activeTunnel.sshClient != nil {
		log.Printf("Closing SSH client for tunnel %d", id)
		if err := activeTunnel.sshClient.Close(); err != nil {
			if !strings.Contains(err.Error(), "use of closed network connection") {
				log.Printf("Error closing SSH client for tunnel %d: %v", id, err)
			}
		}
	}

	// 第四步：等待一小段时间确保资源完全释放
	time.Sleep(200 * time.Millisecond)

	// 更新状态
	s.tunnelRepo.UpdateStatus(id, models.TunnelStatusInactive)
	s.addConnectionLog(id, models.LogEventStop, "Tunnel stopped and port released")

	log.Printf("Tunnel %d stopped successfully", id)
	return nil
}

// RestartTunnel 重启隧道
func (s *tunnelService) RestartTunnel(id uint) error {
	if err := s.StopTunnel(id); err != nil && err.Error() != "tunnel is not running" {
		return err
	}

	time.Sleep(1 * time.Second) // 短暂延迟

	return s.StartTunnel(id)
}

// GetTunnelStatus 获取隧道状态
func (s *tunnelService) GetTunnelStatus(id uint) (string, error) {
	s.mutex.RLock()
	activeTunnel := s.activeTunnels[id]
	s.mutex.RUnlock()

	if activeTunnel != nil {
		return models.TunnelStatusActive, nil
	}

	tunnel, err := s.tunnelRepo.GetByID(id)
	if err != nil {
		return "", err
	}

	return tunnel.Status, nil
}

// StartAutoTunnels 启动自动启动的隧道
func (s *tunnelService) StartAutoTunnels() error {
	tunnels, err := s.tunnelRepo.GetAutoStartTunnels()
	if err != nil {
		return err
	}

	for _, tunnel := range tunnels {
		if err := s.StartTunnel(tunnel.ID); err != nil {
			log.Printf("Failed to start auto tunnel %d (%s): %v", tunnel.ID, tunnel.Name, err)
		}
	}

	return nil
}

// StopAllTunnels 停止所有隧道
func (s *tunnelService) StopAllTunnels() error {
	s.mutex.RLock()
	tunnelIDs := make([]uint, 0, len(s.activeTunnels))
	for id := range s.activeTunnels {
		tunnelIDs = append(tunnelIDs, id)
	}
	s.mutex.RUnlock()

	for _, id := range tunnelIDs {
		if err := s.StopTunnel(id); err != nil {
			log.Printf("Failed to stop tunnel %d: %v", id, err)
		}
	}

	return nil
}

// GetConnectionLogs 获取连接日志
func (s *tunnelService) GetConnectionLogs(tunnelID uint, limit int) ([]models.ConnectionLog, error) {
	return s.tunnelRepo.GetConnectionLogs(tunnelID, limit)
}

// validateTunnelConfig 验证隧道配置
func (s *tunnelService) validateTunnelConfig(tunnel *models.Tunnel) error {
	if tunnel.Name == "" {
		return errors.New("tunnel name is required")
	}

	if tunnel.HostID == 0 {
		return errors.New("host ID is required")
	}

	if tunnel.LocalPort == 0 {
		return errors.New("local port is required")
	}

	switch tunnel.Type {
	case models.TunnelTypeLocalForward:
		if tunnel.RemoteAddress == "" {
			return errors.New("remote address is required for local forward")
		}
		if tunnel.RemotePort == 0 {
			return errors.New("remote port is required for local forward")
		}
	case models.TunnelTypeRemoteForward:
		if tunnel.RemoteAddress == "" {
			tunnel.RemoteAddress = "0.0.0.0" // 默认绑定所有接口
		}
		if tunnel.RemotePort == 0 {
			return errors.New("remote port is required for remote forward")
		}
	case models.TunnelTypeDynamic:
		// 动态转发不需要远程地址和端口
	default:
		return errors.New("invalid tunnel type")
	}

	return nil
}

// checkPortAvailability 检查端口可用性
func (s *tunnelService) checkPortAvailability(tunnel *models.Tunnel) error {
	// 检查本地端口
	localAddr := fmt.Sprintf("%s:%d", tunnel.LocalAddress, tunnel.LocalPort)
	listener, err := net.Listen("tcp", localAddr)
	if err != nil {
		return fmt.Errorf("local port %d is not available: %v", tunnel.LocalPort, err)
	}
	listener.Close()

	// 检查是否与现有隧道冲突
	existingTunnels, err := s.tunnelRepo.GetAll()
	if err == nil {
		for _, existing := range existingTunnels {
			if existing.ID == tunnel.ID {
				continue
			}
			// 检查本地端口冲突
			if existing.LocalPort == tunnel.LocalPort && existing.LocalAddress == tunnel.LocalAddress {
				return fmt.Errorf("local port %d:%d already in use by tunnel %d", tunnel.LocalPort, tunnel.LocalPort, existing.ID)
			}
			// 对于相同主机的远程转发，检查远程端口冲突
			if tunnel.Type == models.TunnelTypeRemoteForward && existing.Type == models.TunnelTypeRemoteForward &&
				existing.HostID == tunnel.HostID &&
				existing.RemotePort == tunnel.RemotePort && existing.RemoteAddress == tunnel.RemoteAddress {
				return fmt.Errorf("remote port %s:%d already in use by tunnel %d", tunnel.RemoteAddress, tunnel.RemotePort, existing.ID)
			}
		}
	}

	return nil
}

// createSSHConnection 创建SSH连接
func (s *tunnelService) createSSHConnection(host *models.Host) (*ssh.Client, error) {
	// 解密敏感数据
	if err := s.hostService.DecryptSensitiveData(host); err != nil {
		return nil, fmt.Errorf("failed to decrypt host data: %v", err)
	}

	// 创建SSH配置
	config, err := s.createSSHConfig(host)
	if err != nil {
		return nil, err
	}

	// 建立连接
	address := fmt.Sprintf("%s:%d", host.Hostname, host.Port)
	client, err := ssh.Dial("tcp", address, config)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// createSSHConfig 创建SSH配置
func (s *tunnelService) createSSHConfig(host *models.Host) (*ssh.ClientConfig, error) {
	config := &ssh.ClientConfig{
		User: host.Username,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil // 在生产环境中应该验证主机密钥
		},
		Timeout: 30 * time.Second,
	}

	switch host.AuthType {
	case models.AuthTypePassword:
		config.Auth = []ssh.AuthMethod{
			ssh.Password(host.Password),
		}
	case models.AuthTypeKey, models.AuthTypeKeyPassword:
		if host.PrivateKey == "" {
			return nil, errors.New("private key is required")
		}

		var signer ssh.Signer
		var err error

		if host.AuthType == models.AuthTypeKeyPassword && host.Passphrase != "" {
			signer, err = ssh.ParsePrivateKeyWithPassphrase([]byte(host.PrivateKey), []byte(host.Passphrase))
		} else {
			signer, err = ssh.ParsePrivateKey([]byte(host.PrivateKey))
		}

		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %v", err)
		}

		config.Auth = []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		}
	}

	return config, nil
}

// addConnectionLog 添加连接日志
func (s *tunnelService) addConnectionLog(tunnelID uint, eventType, message string) {
	connectionLog := &models.ConnectionLog{
		TunnelID:  tunnelID,
		EventType: eventType,
		Message:   message,
		Timestamp: time.Now(),
	}

	if err := s.tunnelRepo.AddConnectionLog(connectionLog); err != nil {
		log.Printf("Failed to add connection log for tunnel %d: %v", tunnelID, err)
	}
}

// CheckServiceHealth 检查本地服务健康状态
func (s *tunnelService) CheckServiceHealth(localAddress string, localPort int) error {
	address := fmt.Sprintf("%s:%d", localAddress, localPort)

	// 设置连接超时
	timeout := 5 * time.Second
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return fmt.Errorf("service %s is not available: %v", address, err)
	}
	defer conn.Close()

	return nil
}

// FindAvailablePort 查找可用端口
func (s *tunnelService) FindAvailablePort(startPort, endPort int, address string) (int, error) {
	if address == "" {
		address = "localhost"
	}

	for port := startPort; port <= endPort; port++ {
		addr := fmt.Sprintf("%s:%d", address, port)
		listener, err := net.Listen("tcp", addr)
		if err == nil {
			listener.Close()
			return port, nil
		}
	}

	return 0, fmt.Errorf("no available port found in range %d-%d", startPort, endPort)
}

// CreateDynamicSOCKS5Tunnel 创建动态SOCKS5隧道，自动分配端口
func (s *tunnelService) CreateDynamicSOCKS5Tunnel(hostID uint, name, description string, autoStart bool) (*models.Tunnel, error) {
	// 查找可用端口（1080-1090是常见的SOCKS代理端口范围）
	localPort, err := s.FindAvailablePort(1080, 1090, "localhost")
	if err != nil {
		// 如果常见端口不可用，尝试更大的范围
		localPort, err = s.FindAvailablePort(8080, 8090, "localhost")
		if err != nil {
			return nil, fmt.Errorf("failed to find available port: %v", err)
		}
	}

	tunnel := &models.Tunnel{
		Name:         name,
		Type:         models.TunnelTypeDynamic,
		HostID:       hostID,
		LocalAddress: "localhost",
		LocalPort:    localPort,
		AutoStart:    autoStart,
		Description:  description,
	}

	if err := s.CreateTunnel(tunnel); err != nil {
		return nil, fmt.Errorf("failed to create tunnel: %v", err)
	}

	return tunnel, nil
}