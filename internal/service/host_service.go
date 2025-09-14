package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"time"

	"github.com/KodaTao/drilling/internal/models"
	"github.com/KodaTao/drilling/internal/repository"
	"golang.org/x/crypto/ssh"
)

// HostService 主机服务接口
type HostService interface {
	CreateHost(host *models.Host) error
	GetHost(id uint) (*models.Host, error)
	GetAllHosts() ([]models.Host, error)
	UpdateHost(host *models.Host) error
	DeleteHost(id uint) error
	TestConnection(id uint) error
	CheckHostStatus(id uint) error
	EncryptSensitiveData(host *models.Host) error
	DecryptSensitiveData(host *models.Host) error
}

// hostService 主机服务实现
type hostService struct {
	hostRepo   repository.HostRepository
	encryptKey []byte
}

// NewHostService 创建主机服务实例
func NewHostService(hostRepo repository.HostRepository, encryptKey string) HostService {
	key := []byte(encryptKey)
	// 确保密钥长度为32字节（AES-256）
	if len(key) < 32 {
		// 扩展密钥到32字节
		extendedKey := make([]byte, 32)
		copy(extendedKey, key)
		key = extendedKey
	} else if len(key) > 32 {
		key = key[:32]
	}

	return &hostService{
		hostRepo:   hostRepo,
		encryptKey: key,
	}
}

// CreateHost 创建主机
func (s *hostService) CreateHost(host *models.Host) error {
	// 检查名称是否已存在
	existingHost, err := s.hostRepo.GetByName(host.Name)
	if err == nil && existingHost != nil {
		return errors.New("host name already exists")
	}

	// 验证认证方式
	if err := s.validateAuthConfig(host); err != nil {
		return err
	}

	// 加密敏感数据
	if err := s.EncryptSensitiveData(host); err != nil {
		return fmt.Errorf("failed to encrypt sensitive data: %v", err)
	}

	// 设置默认状态
	host.Status = models.HostStatusInactive

	return s.hostRepo.Create(host)
}

// GetHost 获取主机
func (s *hostService) GetHost(id uint) (*models.Host, error) {
	host, err := s.hostRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// 解密敏感数据
	if err := s.DecryptSensitiveData(host); err != nil {
		return nil, fmt.Errorf("failed to decrypt sensitive data: %v", err)
	}

	return host, nil
}

// GetAllHosts 获取所有主机
func (s *hostService) GetAllHosts() ([]models.Host, error) {
	hosts, err := s.hostRepo.GetAll()
	if err != nil {
		return nil, err
	}

	// 解密每个主机的敏感数据
	for i := range hosts {
		if err := s.DecryptSensitiveData(&hosts[i]); err != nil {
			// 记录错误但继续处理其他主机
			continue
		}
	}

	return hosts, nil
}

// UpdateHost 更新主机
func (s *hostService) UpdateHost(host *models.Host) error {
	// 验证认证方式
	if err := s.validateAuthConfig(host); err != nil {
		return err
	}

	// 加密敏感数据
	if err := s.EncryptSensitiveData(host); err != nil {
		return fmt.Errorf("failed to encrypt sensitive data: %v", err)
	}

	return s.hostRepo.Update(host)
}

// DeleteHost 删除主机
func (s *hostService) DeleteHost(id uint) error {
	return s.hostRepo.Delete(id)
}

// TestConnection 测试SSH连接
func (s *hostService) TestConnection(id uint) error {
	host, err := s.hostRepo.GetByID(id)
	if err != nil {
		return err
	}

	// 解密敏感数据
	if err := s.DecryptSensitiveData(host); err != nil {
		return fmt.Errorf("failed to decrypt sensitive data: %v", err)
	}

	// 创建SSH配置
	config, err := s.createSSHConfig(host)
	if err != nil {
		return fmt.Errorf("failed to create SSH config: %v", err)
	}

	// 建立连接
	address := fmt.Sprintf("%s:%d", host.Hostname, host.Port)
	client, err := ssh.Dial("tcp", address, config)
	if err != nil {
		// 更新主机状态为错误
		s.hostRepo.UpdateStatus(host.ID, models.HostStatusError)
		return fmt.Errorf("SSH connection failed: %v", err)
	}
	defer client.Close()

	// 执行简单命令测试
	session, err := client.NewSession()
	if err != nil {
		s.hostRepo.UpdateStatus(host.ID, models.HostStatusError)
		return fmt.Errorf("failed to create SSH session: %v", err)
	}
	defer session.Close()

	// 执行echo命令
	output, err := session.Output("echo 'connection test'")
	if err != nil {
		s.hostRepo.UpdateStatus(host.ID, models.HostStatusError)
		return fmt.Errorf("command execution failed: %v", err)
	}

	if string(output) == "" {
		s.hostRepo.UpdateStatus(host.ID, models.HostStatusError)
		return errors.New("empty command output")
	}

	// 更新主机状态为活动
	now := time.Now()
	host.Status = models.HostStatusActive
	host.LastCheck = &now
	s.hostRepo.UpdateStatus(host.ID, models.HostStatusActive)

	return nil
}

// CheckHostStatus 检查主机连接状态
func (s *hostService) CheckHostStatus(id uint) error {
	return s.TestConnection(id)
}

// validateAuthConfig 验证认证配置
func (s *hostService) validateAuthConfig(host *models.Host) error {
	switch host.AuthType {
	case models.AuthTypePassword:
		if host.Password == "" {
			return errors.New("password is required for password authentication")
		}
	case models.AuthTypeKey:
		if host.PrivateKey == "" && host.KeyPath == "" {
			return errors.New("private key or key path is required for key authentication")
		}
	case models.AuthTypeKeyPassword:
		if host.PrivateKey == "" && host.KeyPath == "" {
			return errors.New("private key or key path is required for key with password authentication")
		}
		if host.Passphrase == "" {
			return errors.New("passphrase is required for key with password authentication")
		}
	default:
		return errors.New("invalid authentication type")
	}
	return nil
}

// createSSHConfig 创建SSH配置
func (s *hostService) createSSHConfig(host *models.Host) (*ssh.ClientConfig, error) {
	config := &ssh.ClientConfig{
		User: host.Username,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			// 在生产环境中，应该验证主机密钥
			return nil
		},
		Timeout: 30 * time.Second,
	}

	switch host.AuthType {
	case models.AuthTypePassword:
		config.Auth = []ssh.AuthMethod{
			ssh.Password(host.Password),
		}
	case models.AuthTypeKey, models.AuthTypeKeyPassword:
		var privateKey []byte
		var err error

		if host.PrivateKey != "" {
			privateKey = []byte(host.PrivateKey)
		} else if host.KeyPath != "" {
			// 这里应该读取密钥文件，暂时返回错误
			return nil, errors.New("key file reading not implemented yet")
		}

		var signer ssh.Signer
		if host.AuthType == models.AuthTypeKeyPassword && host.Passphrase != "" {
			signer, err = ssh.ParsePrivateKeyWithPassphrase(privateKey, []byte(host.Passphrase))
		} else {
			signer, err = ssh.ParsePrivateKey(privateKey)
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

// EncryptSensitiveData 加密敏感数据
func (s *hostService) EncryptSensitiveData(host *models.Host) error {
	if host.Password != "" {
		encrypted, err := s.encrypt(host.Password)
		if err != nil {
			return err
		}
		host.Password = encrypted
	}

	if host.PrivateKey != "" {
		encrypted, err := s.encrypt(host.PrivateKey)
		if err != nil {
			return err
		}
		host.PrivateKey = encrypted
	}

	if host.Passphrase != "" {
		encrypted, err := s.encrypt(host.Passphrase)
		if err != nil {
			return err
		}
		host.Passphrase = encrypted
	}

	return nil
}

// DecryptSensitiveData 解密敏感数据
func (s *hostService) DecryptSensitiveData(host *models.Host) error {
	if host.Password != "" {
		// 尝试解密，如果失败可能是未加密的数据
		if decrypted, err := s.decrypt(host.Password); err == nil {
			host.Password = decrypted
		} else {
			// 如果解密失败，检查是否是base64编码错误
			if strings.Contains(err.Error(), "illegal base64 data") {
				// 数据可能未加密，保持原样并重新加密存储
				log.Printf("Host %d password appears to be unencrypted, keeping as-is", host.ID)
			} else {
				return fmt.Errorf("failed to decrypt password: %v", err)
			}
		}
	}

	if host.PrivateKey != "" {
		// 尝试解密，如果失败可能是未加密的数据
		if decrypted, err := s.decrypt(host.PrivateKey); err == nil {
			host.PrivateKey = decrypted
		} else {
			// 如果解密失败，检查是否是base64编码错误
			if strings.Contains(err.Error(), "illegal base64 data") {
				// 数据可能未加密，保持原样
				log.Printf("Host %d private key appears to be unencrypted, keeping as-is", host.ID)
			} else {
				return fmt.Errorf("failed to decrypt private key: %v", err)
			}
		}
	}

	if host.Passphrase != "" {
		// 尝试解密，如果失败可能是未加密的数据
		if decrypted, err := s.decrypt(host.Passphrase); err == nil {
			host.Passphrase = decrypted
		} else {
			// 如果解密失败，检查是否是base64编码错误
			if strings.Contains(err.Error(), "illegal base64 data") {
				// 数据可能未加密，保持原样
				log.Printf("Host %d passphrase appears to be unencrypted, keeping as-is", host.ID)
			} else {
				return fmt.Errorf("failed to decrypt passphrase: %v", err)
			}
		}
	}

	return nil
}

// encrypt 加密字符串
func (s *hostService) encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(s.encryptKey)
	if err != nil {
		return "", err
	}

	plainBytes := []byte(plaintext)
	ciphertext := make([]byte, aes.BlockSize+len(plainBytes))
	iv := ciphertext[:aes.BlockSize]

	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plainBytes)

	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

// decrypt 解密字符串
func (s *hostService) decrypt(ciphertext string) (string, error) {
	ciphertextBytes, err := base64.URLEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(s.encryptKey)
	if err != nil {
		return "", err
	}

	if len(ciphertextBytes) < aes.BlockSize {
		return "", errors.New("ciphertext too short")
	}

	iv := ciphertextBytes[:aes.BlockSize]
	ciphertextBytes = ciphertextBytes[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertextBytes, ciphertextBytes)

	return string(ciphertextBytes), nil
}