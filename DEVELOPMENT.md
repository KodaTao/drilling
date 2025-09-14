# Drilling 开发计划

## 项目概述

Drilling 是一个基于 SSH 隧道的异地联网管理工具，采用 Go + Vue.js 技术栈，为本地部署设计，提供简洁易用的 Web 管理界面来管理各种 SSH 隧道连接。无需用户认证，开箱即用。

## 核心功能分析

根据 README.md，项目的核心功能包括：

1. **主机管理** - 管理远程主机连接信息
2. **远程服务映射** - 将远程主机的服务端口映射到本地
3. **本地服务映射** - 将本地服务端口映射到远程主机
4. **动态隧道/SOCKS5代理** - 提供动态代理服务
5. **Clash配置导出** - 将动态隧道导出为Clash配置

## 技术架构

### 后端技术栈
- **Go 1.19+** - 主要编程语言
- **Gin** - Web框架
- **GORM** - ORM框架
- **SQLite3** - 数据库
- **embed** - 静态资源嵌入
- **Viper** - 配置管理
- **logrus/zap** - 日志系统

### 前端技术栈
- **Vue.js 3** - 前端框架（Composition API）
- **TypeScript** - 类型安全
- **Vite** - 构建工具
- **Element Plus** - UI组件库
- **Pinia** - 状态管理
- **Vue Router 4** - 路由管理
- **Axios** - HTTP客户端
- **SCSS** - 样式预处理

## 项目结构设计

```
drilling/
├── cmd/                    # 命令行入口
│   └── main.go
├── internal/               # 内部包
│   ├── api/               # API控制器
│   │   ├── host.go        # 主机管理
│   │   ├── tunnel.go      # 隧道管理
│   │   └── export.go      # 配置导出
│   ├── config/            # 配置管理
│   │   └── config.go
│   ├── database/          # 数据库
│   │   ├── db.go
│   │   └── migrate.go
│   ├── middleware/        # 中间件
│   │   ├── cors.go
│   │   └── logger.go
│   ├── models/            # 数据模型
│   │   ├── host.go        # 主机模型
│   │   └── tunnel.go      # 隧道模型
│   ├── services/          # 业务逻辑
│   │   ├── ssh.go         # SSH连接管理
│   │   ├── tunnel.go      # 隧道服务
│   │   ├── export.go      # 配置导出
│   │   └── monitor.go     # 连接监控
│   └── utils/             # 工具函数
│       ├── ssh_utils.go   # SSH工具
│       ├── crypto.go      # 加密工具
│       └── response.go    # 响应工具
├── web/                   # 前端项目
│   ├── src/
│   │   ├── api/           # API调用
│   │   ├── components/    # 组件
│   │   │   ├── HostManager.vue      # 主机管理组件
│   │   │   ├── TunnelManager.vue    # 隧道管理组件
│   │   │   └── ExportConfig.vue     # 配置导出组件
│   │   ├── layouts/       # 布局
│   │   ├── pages/         # 页面
│   │   │   ├── HostManage.vue       # 主机管理页面
│   │   │   ├── RemoteService.vue    # 远程服务管理
│   │   │   ├── LocalService.vue     # 本地服务管理
│   │   │   └── DynamicTunnel.vue    # 动态隧道管理
│   │   ├── router/        # 路由配置
│   │   ├── stores/        # 状态管理
│   │   └── types/         # 类型定义
│   ├── package.json
│   └── vite.config.ts
├── configs/               # 配置文件
├── go.mod
├── go.sum
└── README.md
```

## 开发阶段规划

### 第一阶段：项目基础搭建（1-2周）

#### 后端基础
- [ ] 初始化Go模块，配置基础依赖
- [ ] 搭建Gin Web服务器
- [ ] 配置GORM和SQLite数据库
- [ ] 实现基础中间件（CORS、日志）
- [ ] 设计数据库模型（主机、隧道）
- [ ] 实现配置管理系统

#### 前端基础
- [ ] 初始化Vue.js 3项目
- [ ] 配置TypeScript + Vite构建环境
- [ ] 集成Element Plus UI框架
- [ ] 配置Vue Router和Pinia
- [ ] 实现基础布局和导航

#### 构建系统
- [ ] 配置前端资源嵌入
- [ ] 编写构建脚本
- [ ] 实现单文件打包

### 第二阶段：主机管理功能（1周）

#### 后端功能
- [ ] 实现主机数据模型
- [ ] 开发主机管理API（增删改查）
- [ ] 实现SSH连接测试功能
- [ ] 支持多种认证方式（密码、密钥）
- [ ] 添加主机连接状态检测

#### 前端功能
- [ ] 设计主机管理界面
- [ ] 实现主机列表展示
- [ ] 开发主机添加/编辑表单
- [ ] 实现SSH连接测试UI
- [ ] 添加连接状态指示器

### 第三阶段：SSH隧道核心功能（2-3周）

#### 远程服务映射功能
- [ ] 实现远程到本地的端口转发
- [ ] 开发隧道状态管理
- [ ] 添加连接监控和自动重连
- [ ] 实现隧道配置持久化

#### 本地服务映射功能
- [ ] 实现本地到远程的端口转发
- [ ] 支持多端口映射
- [ ] 添加端口冲突检测
- [ ] 实现服务健康检查

#### 动态隧道/SOCKS5代理
- [ ] 实现SOCKS5代理服务
- [ ] 支持动态端口分配
- [ ] 实现流量统计

#### 前端隧道管理界面
- [ ] 设计隧道管理页面
- [ ] 实现隧道列表和状态显示
- [ ] 开发隧道创建/编辑表单
- [ ] 添加一键启动/停止功能
- [ ] 实现实时状态更新

### 第四阶段：配置导出和监控（1-2周）

#### Clash配置导出
- [ ] 实现Clash配置文件生成（全部的socks5整合到一个文件）和下载

#### 监控和统计
- [ ] 实现连接状态监控
- [ ] 添加流量统计功能
- [ ] 开发日志查看功能
- [ ] 实现告警通知机制

#### 前端监控界面
- [ ] 设计监控仪表板
- [ ] 实现实时状态图表
- [ ] 开发日志查看器
- [ ] 添加统计报表功能

### 第五阶段：系统优化和完善（1周）

#### 性能优化
- [ ] 优化SSH连接池管理
- [ ] 实现连接复用机制
- [ ] 添加资源监控和限制
- [ ] 优化前端渲染性能

#### 安全加固
- [ ] 增强数据加密存储
- [ ] 实现操作审计日志
- [ ] 添加本地访问限制（仅localhost访问）

#### 用户体验优化
- [ ] 完善错误处理和提示
- [ ] 实现配置备份恢复
- [ ] 添加批量操作功能
- [ ] 优化移动端适配

## 数据库设计

### 主机表（hosts）
```sql
CREATE TABLE hosts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(100) NOT NULL UNIQUE,
    hostname VARCHAR(255) NOT NULL,
    port INTEGER DEFAULT 22,
    username VARCHAR(100) NOT NULL,
    auth_type VARCHAR(20) NOT NULL, -- password, key, key_password
    password VARCHAR(255), -- 加密存储
    private_key_path VARCHAR(255),
    private_key_passphrase VARCHAR(255), -- 加密存储
    description TEXT,
    status VARCHAR(20) DEFAULT 'inactive', -- active, inactive, error
    last_check TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### 隧道表（tunnels）
```sql
CREATE TABLE tunnels (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    host_id INTEGER NOT NULL,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(20) NOT NULL, -- local_forward, remote_forward, dynamic
    local_address VARCHAR(255) DEFAULT '127.0.0.1',
    local_port INTEGER NOT NULL,
    remote_address VARCHAR(255),
    remote_port INTEGER,
    description TEXT,
    status VARCHAR(20) DEFAULT 'inactive', -- active, inactive, error
    auto_start BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (host_id) REFERENCES hosts(id) ON DELETE CASCADE
);
```

### 连接日志表（connection_logs）
```sql
CREATE TABLE connection_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tunnel_id INTEGER,
    event_type VARCHAR(50) NOT NULL, -- connect, disconnect, error
    message TEXT,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (tunnel_id) REFERENCES tunnels(id) ON DELETE CASCADE
);
```

## API设计

### 主机管理API
```
GET    /api/v1/hosts           # 获取主机列表
POST   /api/v1/hosts           # 创建主机
GET    /api/v1/hosts/:id       # 获取主机详情
PUT    /api/v1/hosts/:id       # 更新主机
DELETE /api/v1/hosts/:id       # 删除主机
POST   /api/v1/hosts/:id/test  # 测试主机连接
```

### 隧道管理API
```
GET    /api/v1/tunnels         # 获取隧道列表
POST   /api/v1/tunnels         # 创建隧道
GET    /api/v1/tunnels/:id     # 获取隧道详情
PUT    /api/v1/tunnels/:id     # 更新隧道
DELETE /api/v1/tunnels/:id     # 删除隧道
POST   /api/v1/tunnels/:id/start   # 启动隧道
POST   /api/v1/tunnels/:id/stop    # 停止隧道
```

### 配置导出API
```
GET    /api/v1/export/clash    # 导出Clash配置
GET    /api/v1/export/tunnels  # 导出隧道配置
```

### 监控API
```
GET    /api/v1/stats           # 获取统计信息
GET    /api/v1/logs            # 获取连接日志
GET    /api/v1/status          # 获取系统状态
```

## 核心服务实现

### SSH连接管理服务
```go
type SSHService struct {
    connections map[int]*ssh.Client // host_id -> connection
    mutex       sync.RWMutex
}

func (s *SSHService) Connect(host *models.Host) error
func (s *SSHService) Disconnect(hostID int) error
func (s *SSHService) GetConnection(hostID int) *ssh.Client
func (s *SSHService) TestConnection(host *models.Host) error
```

### 隧道管理服务
```go
type TunnelService struct {
    activeTunnels map[int]*TunnelInstance // tunnel_id -> instance
    sshService    *SSHService
    mutex         sync.RWMutex
}

func (t *TunnelService) StartTunnel(tunnel *models.Tunnel) error
func (t *TunnelService) StopTunnel(tunnelID int) error
func (t *TunnelService) GetStatus(tunnelID int) string
func (t *TunnelService) MonitorTunnels()
```

## 开发环境配置

### 后端开发环境
```bash
# 初始化Go模块
go mod init github.com/KodaTao/drilling

# 安装依赖
go get github.com/gin-gonic/gin
go get gorm.io/gorm
go get gorm.io/driver/sqlite
go get golang.org/x/crypto/ssh
go get github.com/spf13/viper
go get github.com/sirupsen/logrus

# 运行开发服务器
go run cmd/main.go
```

### 前端开发环境
```bash
# 进入前端目录
cd web

# 安装依赖
npm install

# 安装开发依赖
npm install -D @types/node typescript

# 启动开发服务器
npm run dev
```

## 构建和部署

### 开发环境构建
```bash
# 构建前端
cd web && npm run build && cd ..

# 构建后端（嵌入前端资源）
go build -o drilling cmd/main.go

# 运行
./drilling
```

### 生产环境部署
```bash
# 交叉编译
GOOS=linux GOARCH=amd64 go build -o drilling-linux-amd64 cmd/main.go
GOOS=windows GOARCH=amd64 go build -o drilling-windows-amd64.exe cmd/main.go
GOOS=darwin GOARCH=amd64 go build -o drilling-darwin-amd64 cmd/main.go
```

## 测试策略

### 单元测试
- SSH连接功能测试
- 隧道创建和管理测试
- API接口测试
- 数据库操作测试

### 集成测试
- 端到端隧道连接测试
- Web界面功能测试
- 配置导出功能测试

### 性能测试
- 并发连接测试
- 大量隧道管理测试
- 长时间运行稳定性测试

## 安全考虑

### 数据安全
- SSH密钥和密码加密存储
- 传输数据加密
- 敏感信息脱敏日志

### 访问控制
- 本地访问限制（默认仅localhost访问）
- 可选的基础认证配置

### 运行安全
- 端口范围限制
- 连接数量限制
- 资源使用监控

## 部署和运维

### 系统要求
- Linux/Windows/macOS
- 可访问目标SSH主机
- 足够的网络带宽

### 配置文件
```yaml
server:
  port: 8080
  host: "127.0.0.1"  # 默认仅本地访问
  # host: "0.0.0.0"  # 如需远程访问可修改

database:
  path: "./drilling.db"

ssh:
  timeout: 30s
  keepalive: 10s
  max_connections: 100

logging:
  level: "info"
  file: "./drilling.log"

# 可选的基础认证
# auth:
#   enabled: false
#   username: "admin"
#   password: "password"
```

### 监控和维护
- 服务状态监控
- 日志轮转管理
- 数据库备份策略
- 性能指标收集

这个开发计划基于更新后的README.md内容，专注于SSH隧道的核心功能，去除了FRP相关内容，更加贴合项目的实际需求。