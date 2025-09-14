export interface Host {
  id: number
  name: string
  hostname: string
  port: number
  username: string
  auth_type: 'password' | 'key' | 'key_password'
  password?: string
  private_key?: string
  key_path?: string
  passphrase?: string
  description: string
  status: 'active' | 'inactive' | 'error'
  last_check?: string
  created_at: string
  updated_at: string
  tunnels?: Tunnel[]
}

export interface Tunnel {
  id: number
  host_id: number
  name: string
  type: 'local_forward' | 'remote_forward' | 'dynamic'
  local_address: string
  local_port: number
  remote_address?: string
  remote_port?: number
  description: string
  status: 'active' | 'inactive' | 'error'
  auto_start: boolean
  created_at: string
  updated_at: string
}

export interface ConnectionLog {
  id: number
  tunnel_id: number
  event_type: 'connect' | 'disconnect' | 'error' | 'start' | 'stop'
  message: string
  timestamp: string
}

export interface ApiResponse<T> {
  success?: boolean
  data?: T
  message?: string
  error?: string
  details?: string
}

export interface ClashProxy {
  name: string
  type: string
  server: string
  port: number
}

export interface ClashProxyGroup {
  name: string
  type: string
  proxies: string[]
  url?: string
  interval?: number
}

export interface ClashDNS {
  enable: boolean
  listen: string
  nameserver: string[]
  'enhanced-mode': string
  'fake-ip-range': string
  'use-hosts': boolean
  'fake-ip-filter': string[]
}

export interface ClashConfig {
  port: number
  'socks-port': number
  'allow-lan': boolean
  mode: string
  'log-level': string
  'external-ui': string
  'external-controller': string
  secret?: string
  proxies: ClashProxy[]
  'proxy-groups': ClashProxyGroup[]
  rules: string[]
  dns: ClashDNS
}

export interface ClashExportResponse {
  message: string
  config: ClashConfig
  tunnel_count: number
  tunnels: Tunnel[]
  generated_at: string
}

export interface Socks5StatusResponse {
  message: string
  active_count: number
  can_export: boolean
  tunnels: {
    id: number
    name: string
    host_id: number
    local_address: string
    local_port: number
    status: string
    created_at: string
  }[]
  last_check: string
}

export interface HostListResponse {
  hosts: Host[]
  count: number
}

export interface ConnectionTestResponse {
  success: boolean
  message: string
  details?: string
}

export interface StatusCheckResponse {
  status: string
  last_check?: string
  message: string
}