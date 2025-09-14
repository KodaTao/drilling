package socks5

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strconv"

	"golang.org/x/crypto/ssh"
)

// SOCKS5 协议常量
const (
	Socks5Version = 0x05
	// 认证方式
	AuthMethodNoAuth       = 0x00
	AuthMethodPassword     = 0x02
	AuthMethodNoAcceptable = 0xFF
	// 命令类型
	CmdConnect      = 0x01
	CmdBind         = 0x02
	CmdUDPAssociate = 0x03
	// 地址类型
	AtypIPV4   = 0x01
	AtypDomain = 0x03
	AtypIPV6   = 0x04
	// 响应状态
	RepSuccess             = 0x00
	RepGeneralFailure      = 0x01
	RepConnectionNotAllowed = 0x02
	RepNetworkUnreachable  = 0x03
	RepHostUnreachable     = 0x04
	RepConnectionRefused   = 0x05
	RepTTLExpired          = 0x06
	RepCommandNotSupported = 0x07
	RepAddressNotSupported = 0x08
)

// SOCKS5Server SOCKS5代理服务器
type SOCKS5Server struct {
	sshClient     *ssh.Client
	trafficLogger TrafficLogger
}

// TrafficLogger 流量记录接口
type TrafficLogger interface {
	LogTraffic(bytesIn, bytesOut int64)
}

// NewSOCKS5Server 创建新的SOCKS5服务器
func NewSOCKS5Server(sshClient *ssh.Client) *SOCKS5Server {
	return &SOCKS5Server{
		sshClient: sshClient,
	}
}

// NewSOCKS5ServerWithTrafficLogger 创建带流量统计的SOCKS5服务器
func NewSOCKS5ServerWithTrafficLogger(sshClient *ssh.Client, logger TrafficLogger) *SOCKS5Server {
	return &SOCKS5Server{
		sshClient:     sshClient,
		trafficLogger: logger,
	}
}

// HandleConnection 处理SOCKS5连接
func (s *SOCKS5Server) HandleConnection(ctx context.Context, conn net.Conn) error {
	defer conn.Close()

	// 1. 认证协商
	if err := s.handleAuth(conn); err != nil {
		return fmt.Errorf("authentication failed: %v", err)
	}

	// 2. 处理请求
	if err := s.handleRequest(ctx, conn); err != nil {
		return fmt.Errorf("request handling failed: %v", err)
	}

	return nil
}

// handleAuth 处理认证协商
func (s *SOCKS5Server) handleAuth(conn net.Conn) error {
	// 读取客户端认证请求
	buf := make([]byte, 256)
	n, err := conn.Read(buf)
	if err != nil {
		return err
	}

	if n < 2 || buf[0] != Socks5Version {
		return fmt.Errorf("invalid SOCKS5 version")
	}

	nMethods := int(buf[1])
	if n < 2+nMethods {
		return fmt.Errorf("invalid authentication request")
	}

	// 检查支持的认证方式
	supportNoAuth := false
	for i := 0; i < nMethods; i++ {
		if buf[2+i] == AuthMethodNoAuth {
			supportNoAuth = true
			break
		}
	}

	// 响应认证方式
	var response []byte
	if supportNoAuth {
		response = []byte{Socks5Version, AuthMethodNoAuth}
	} else {
		response = []byte{Socks5Version, AuthMethodNoAcceptable}
	}

	if _, err := conn.Write(response); err != nil {
		return err
	}

	if !supportNoAuth {
		return fmt.Errorf("no acceptable authentication method")
	}

	return nil
}

// handleRequest 处理SOCKS5请求
func (s *SOCKS5Server) handleRequest(ctx context.Context, conn net.Conn) error {
	// 读取请求
	buf := make([]byte, 256)
	n, err := conn.Read(buf)
	if err != nil {
		return err
	}

	if n < 4 || buf[0] != Socks5Version {
		return fmt.Errorf("invalid request")
	}

	cmd := buf[1]
	atyp := buf[3]

	if cmd != CmdConnect {
		// 发送不支持的命令响应
		response := []byte{Socks5Version, RepCommandNotSupported, 0x00, AtypIPV4, 0, 0, 0, 0, 0, 0}
		conn.Write(response)
		return fmt.Errorf("unsupported command: %d", cmd)
	}

	// 解析目标地址
	var destAddr string
	var destPort int
	offset := 4

	switch atyp {
	case AtypIPV4:
		if n < offset+6 {
			return fmt.Errorf("incomplete IPv4 address")
		}
		ip := net.IPv4(buf[offset], buf[offset+1], buf[offset+2], buf[offset+3])
		destAddr = ip.String()
		destPort = int(binary.BigEndian.Uint16(buf[offset+4 : offset+6]))
	case AtypDomain:
		if n < offset+1 {
			return fmt.Errorf("incomplete domain length")
		}
		domainLen := int(buf[offset])
		if n < offset+1+domainLen+2 {
			return fmt.Errorf("incomplete domain address")
		}
		destAddr = string(buf[offset+1 : offset+1+domainLen])
		destPort = int(binary.BigEndian.Uint16(buf[offset+1+domainLen : offset+1+domainLen+2]))
	case AtypIPV6:
		if n < offset+18 {
			return fmt.Errorf("incomplete IPv6 address")
		}
		ip := net.IP(buf[offset : offset+16])
		destAddr = ip.String()
		destPort = int(binary.BigEndian.Uint16(buf[offset+16 : offset+18]))
	default:
		response := []byte{Socks5Version, RepAddressNotSupported, 0x00, AtypIPV4, 0, 0, 0, 0, 0, 0}
		conn.Write(response)
		return fmt.Errorf("unsupported address type: %d", atyp)
	}

	// 通过SSH连接到目标地址
	target := net.JoinHostPort(destAddr, strconv.Itoa(destPort))
	remoteConn, err := s.sshClient.Dial("tcp", target)
	if err != nil {
		response := []byte{Socks5Version, RepHostUnreachable, 0x00, AtypIPV4, 0, 0, 0, 0, 0, 0}
		conn.Write(response)
		return fmt.Errorf("failed to connect to %s: %v", target, err)
	}
	defer remoteConn.Close()

	// 发送成功响应
	response := []byte{Socks5Version, RepSuccess, 0x00, AtypIPV4, 0, 0, 0, 0, 0, 0}
	if _, err := conn.Write(response); err != nil {
		return err
	}

	// 双向数据转发
	return s.relay(ctx, conn, remoteConn)
}

// relay 双向数据转发
func (s *SOCKS5Server) relay(ctx context.Context, conn1, conn2 net.Conn) error {
	done := make(chan error, 2)
	var bytesIn, bytesOut int64

	// conn1 -> conn2 (client to remote)
	go func() {
		n, err := io.Copy(conn2, conn1)
		bytesOut = n
		done <- err
	}()

	// conn2 -> conn1 (remote to client)
	go func() {
		n, err := io.Copy(conn1, conn2)
		bytesIn = n
		done <- err
	}()

	// 等待任一方向完成或上下文取消
	var err error
	select {
	case err = <-done:
	case <-ctx.Done():
		err = ctx.Err()
	}

	// 记录流量统计
	if s.trafficLogger != nil {
		s.trafficLogger.LogTraffic(bytesIn, bytesOut)
	}

	return err
}