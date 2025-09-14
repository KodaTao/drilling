package service

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/KodaTao/drilling/internal/models"
	"github.com/KodaTao/drilling/internal/repository"
	"gopkg.in/yaml.v3"
)

// ClashExportService Clash配置导出服务接口
type ClashExportService interface {
	GenerateClashConfig() (*ClashConfig, error)
	ExportClashConfigYAML() ([]byte, error)
	GetActiveSocks5Tunnels() ([]models.Tunnel, error)
}

// clashExportService Clash配置导出服务实现
type clashExportService struct {
	tunnelRepo repository.TunnelRepository
	hostRepo   repository.HostRepository
}

// NewClashExportService 创建Clash导出服务实例
func NewClashExportService(tunnelRepo repository.TunnelRepository, hostRepo repository.HostRepository) ClashExportService {
	return &clashExportService{
		tunnelRepo: tunnelRepo,
		hostRepo:   hostRepo,
	}
}

// ClashConfig Clash配置结构
type ClashConfig struct {
	Port          int                    `yaml:"port"`
	SocksPort     int                    `yaml:"socks-port"`
	AllowLan      bool                   `yaml:"allow-lan"`
	Mode          string                 `yaml:"mode"`
	LogLevel      string                 `yaml:"log-level"`
	ExternalUI    string                 `yaml:"external-ui"`
	ExternalController string            `yaml:"external-controller"`
	Secret        string                 `yaml:"secret,omitempty"`
	Proxies       []ClashProxy           `yaml:"proxies"`
	ProxyGroups   []ClashProxyGroup      `yaml:"proxy-groups"`
	Rules         []string               `yaml:"rules"`
	DNS           ClashDNS               `yaml:"dns"`
}

// ClashProxy Clash代理配置
type ClashProxy struct {
	Name   string `yaml:"name"`
	Type   string `yaml:"type"`
	Server string `yaml:"server"`
	Port   int    `yaml:"port"`
}

// ClashProxyGroup Clash代理组配置
type ClashProxyGroup struct {
	Name     string   `yaml:"name"`
	Type     string   `yaml:"type"`
	Proxies  []string `yaml:"proxies"`
	URL      string   `yaml:"url,omitempty"`
	Interval int      `yaml:"interval,omitempty"`
}

// ClashDNS Clash DNS配置
type ClashDNS struct {
	Enable           bool     `yaml:"enable"`
	Listen           string   `yaml:"listen"`
	NameServer       []string `yaml:"nameserver"`
	EnhancedMode     string   `yaml:"enhanced-mode"`
	FakeIPRange      string   `yaml:"fake-ip-range"`
	UseHosts         bool     `yaml:"use-hosts"`
	FakeIPFilter     []string `yaml:"fake-ip-filter"`
}

// GenerateClashConfig 生成Clash配置
func (s *clashExportService) GenerateClashConfig() (*ClashConfig, error) {
	// 获取所有活跃的SOCKS5隧道
	socks5Tunnels, err := s.GetActiveSocks5Tunnels()
	if err != nil {
		return nil, fmt.Errorf("failed to get SOCKS5 tunnels: %v", err)
	}

	if len(socks5Tunnels) == 0 {
		return nil, fmt.Errorf("no active SOCKS5 tunnels found")
	}

	// 创建基础配置
	config := &ClashConfig{
		Port:               7890, // HTTP代理端口
		SocksPort:          7891, // SOCKS5代理端口
		AllowLan:           false, // 默认不允许局域网访问
		Mode:               "rule", // 规则模式
		LogLevel:           "info",
		ExternalUI:         "", // 清空外部UI配置，避免路径错误
		ExternalController: "127.0.0.1:9090", // Clash面板地址
		Proxies:            []ClashProxy{},
		ProxyGroups:        []ClashProxyGroup{},
		Rules:              []string{},
		DNS: ClashDNS{
			Enable:       true,
			Listen:       "0.0.0.0:53",
			NameServer:   []string{"223.5.5.5", "1.1.1.1"},
			EnhancedMode: "fake-ip",
			FakeIPRange:  "198.18.0.1/16",
			UseHosts:     true,
			FakeIPFilter: []string{
				"*.lan",
				"localhost.ptlogin2.qq.com",
				"dns.msftncsi.com",
				"www.msftncsi.com",
				"www.msftconnecttest.com",
			},
		},
	}

	// 生成代理节点
	proxyNames := []string{}
	for _, tunnel := range socks5Tunnels {
		// 获取主机信息
		host, err := s.hostRepo.GetByID(tunnel.HostID)
		if err != nil {
			continue
		}

		// 生成代理节点名称
		proxyName := fmt.Sprintf("drilling-%s-%d", sanitizeName(host.Name), tunnel.LocalPort)

		// 创建SOCKS5代理配置
		proxy := ClashProxy{
			Name:   proxyName,
			Type:   "socks5",
			Server: tunnel.LocalAddress,
			Port:   int(tunnel.LocalPort),
		}

		config.Proxies = append(config.Proxies, proxy)
		proxyNames = append(proxyNames, proxyName)
	}

	// 如果有多个代理，创建代理组
	if len(proxyNames) > 0 {
		// 自动选择组（延迟测试）
		autoGroup := ClashProxyGroup{
			Name:     "Auto",
			Type:     "url-test",
			Proxies:  append([]string{}, proxyNames...),
			URL:      "http://www.gstatic.com/generate_204",
			Interval: 300,
		}
		config.ProxyGroups = append(config.ProxyGroups, autoGroup)

		// 手动选择组
		selectGroup := ClashProxyGroup{
			Name:    "Proxy",
			Type:    "select",
			Proxies: append([]string{"Auto", "DIRECT"}, proxyNames...),
		}
		config.ProxyGroups = append(config.ProxyGroups, selectGroup)

		// 如果有多个代理，还可以添加负载均衡组
		if len(proxyNames) > 1 {
			loadBalanceGroup := ClashProxyGroup{
				Name:     "LoadBalance",
				Type:     "load-balance",
				Proxies:  append([]string{}, proxyNames...),
				URL:      "http://www.gstatic.com/generate_204",
				Interval: 300,
			}
			config.ProxyGroups = append(config.ProxyGroups, loadBalanceGroup)

			// 更新手动选择组，添加负载均衡选项
			selectGroup.Proxies = append([]string{"Auto", "LoadBalance", "DIRECT"}, proxyNames...)
			config.ProxyGroups[1] = selectGroup
		}
	}

	// 生成基础规则
	config.Rules = []string{
		// 本地地址直连
		"DOMAIN-SUFFIX,local,DIRECT",
		"DOMAIN-SUFFIX,localhost,DIRECT",
		"DOMAIN-SUFFIX,lan,DIRECT",
		"IP-CIDR,127.0.0.0/8,DIRECT",
		"IP-CIDR,169.254.0.0/16,DIRECT",
		"IP-CIDR,192.168.0.0/16,DIRECT",
		"IP-CIDR,10.0.0.0/8,DIRECT",
		"IP-CIDR,172.16.0.0/12,DIRECT",
		"IP-CIDR,224.0.0.0/4,DIRECT",
		"IP-CIDR,240.0.0.0/4,DIRECT",

		// 默认规则：国外网站使用代理，国内直连
		"GEOIP,CN,DIRECT",
		"MATCH,Proxy",
	}

	return config, nil
}

// ExportClashConfigYAML 导出Clash配置为YAML格式
func (s *clashExportService) ExportClashConfigYAML() ([]byte, error) {
	config, err := s.GenerateClashConfig()
	if err != nil {
		return nil, err
	}

	// 序列化为YAML
	yamlData, err := yaml.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal YAML: %v", err)
	}

	// 添加配置文件头部注释
	header := fmt.Sprintf(`# Drilling Platform - Clash Configuration
# Generated at: %s
# Total SOCKS5 proxies: %d
#
# This configuration file was automatically generated by Drilling Platform
# It includes all active SOCKS5 tunnels as proxy nodes
#
# Usage:
# 1. Save this file as config.yaml in your Clash config directory
# 2. Start Clash client and select appropriate proxy group
# 3. Configure your system proxy to use Clash (HTTP: 7890, SOCKS5: 7891)
#
# External Controller: http://127.0.0.1:9090 (for Clash dashboard)
#
# Optional: If you want to use Clash Dashboard UI, uncomment the following line:
# external-ui: /path/to/clash-dashboard
#
# You can download Clash Dashboard from: https://github.com/Dreamacro/clash-dashboard

`, time.Now().Format("2006-01-02 15:04:05"), len(config.Proxies))

	return append([]byte(header), yamlData...), nil
}

// GetActiveSocks5Tunnels 获取所有活跃的SOCKS5隧道
func (s *clashExportService) GetActiveSocks5Tunnels() ([]models.Tunnel, error) {
	// 获取所有隧道
	allTunnels, err := s.tunnelRepo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get tunnels: %v", err)
	}

	var socks5Tunnels []models.Tunnel
	for _, tunnel := range allTunnels {
		// 只包含SOCKS5类型且状态为活跃的隧道
		if tunnel.Type == "dynamic" && tunnel.Status == "active" {
			socks5Tunnels = append(socks5Tunnels, tunnel)
		}
	}

	// 按端口排序，确保配置文件的一致性
	sort.Slice(socks5Tunnels, func(i, j int) bool {
		return socks5Tunnels[i].LocalPort < socks5Tunnels[j].LocalPort
	})

	return socks5Tunnels, nil
}

// sanitizeName 清理名称，移除特殊字符
func sanitizeName(name string) string {
	// 替换常见的特殊字符
	replacer := strings.NewReplacer(
		" ", "-",
		"_", "-",
		".", "-",
		":", "-",
		"/", "-",
		"\\", "-",
		"|", "-",
		"*", "-",
		"?", "-",
		"\"", "",
		"'", "",
		"<", "",
		">", "",
	)

	result := replacer.Replace(name)

	// 移除连续的连字符
	for strings.Contains(result, "--") {
		result = strings.ReplaceAll(result, "--", "-")
	}

	// 移除开头和结尾的连字符
	result = strings.Trim(result, "-")

	// 如果结果为空，使用默认名称
	if result == "" {
		result = "proxy"
	}

	return result
}