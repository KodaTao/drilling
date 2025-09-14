import { apiClient } from './client';

export interface Tunnel {
  id: number;
  host_id: number;
  name: string;
  type: 'local_forward' | 'remote_forward' | 'dynamic';
  local_address: string;
  local_port: number;
  remote_address?: string;
  remote_port?: number;
  description?: string;
  status: 'active' | 'inactive' | 'error';
  auto_start: boolean;
  created_at: string;
  updated_at: string;
  host?: {
    id: number;
    name: string;
    address: string;
    port: number;
    username: string;
    status: string;
  };
}

export interface CreateTunnelRequest {
  host_id: number;
  name: string;
  type: 'local_forward' | 'remote_forward' | 'dynamic';
  local_address: string;
  local_port: number;
  remote_address?: string;
  remote_port?: number;
  description?: string;
  auto_start?: boolean;
}

export interface LocalServiceMapping {
  name: string;
  local_address: string;
  local_port: number;
  remote_address: string;
  remote_port: number;
  auto_start: boolean;
  description?: string;
}

export interface CreateMultipleLocalForwardsRequest {
  services: LocalServiceMapping[];
}

export interface CreateDynamicSOCKS5Request {
  name: string;
  description?: string;
  auto_start: boolean;
}

export interface HealthCheckRequest {
  local_address: string;
  local_port: number;
}

export interface HealthCheckResponse {
  healthy: boolean;
  message: string;
}

export interface FindAvailablePortRequest {
  start_port: number;
  end_port: number;
  address?: string;
}

export interface FindAvailablePortResponse {
  available_port: number;
  address: string;
}

export interface ConnectionLog {
  id: number;
  tunnel_id: number;
  event_type: string;
  message: string;
  timestamp: string;
}

export interface TunnelStatus {
  status: string;
}

class TunnelApi {
  // 基本CRUD操作
  async createTunnel(tunnel: CreateTunnelRequest): Promise<Tunnel> {
    const response = await apiClient.post('/tunnels', tunnel);
    return response.data.tunnel;
  }

  async getAllTunnels(): Promise<Tunnel[]> {
    const response = await apiClient.get('/tunnels');
    return response.data.tunnels || [];
  }

  async getTunnel(id: number): Promise<Tunnel> {
    const response = await apiClient.get(`/tunnels/${id}`);
    return response.data.tunnel;
  }

  async getTunnelsByHost(hostId: number): Promise<Tunnel[]> {
    const response = await apiClient.get(`/tunnels/by-host/${hostId}`);
    return response.data.tunnels || [];
  }

  async updateTunnel(id: number, tunnel: Partial<CreateTunnelRequest>): Promise<Tunnel> {
    const response = await apiClient.put(`/tunnels/${id}`, tunnel);
    return response.data.tunnel;
  }

  async deleteTunnel(id: number): Promise<void> {
    await apiClient.delete(`/tunnels/${id}`);
  }

  // 隧道控制操作
  async startTunnel(id: number): Promise<void> {
    await apiClient.post(`/tunnels/${id}/start`);
  }

  async stopTunnel(id: number): Promise<void> {
    await apiClient.post(`/tunnels/${id}/stop`);
  }

  async restartTunnel(id: number): Promise<void> {
    await apiClient.post(`/tunnels/${id}/restart`);
  }

  async getTunnelStatus(id: number): Promise<TunnelStatus> {
    const response = await apiClient.get(`/tunnels/${id}/status`);
    return response.data;
  }

  async getConnectionLogs(id: number, limit = 100): Promise<ConnectionLog[]> {
    const response = await apiClient.get(`/tunnels/${id}/logs`, {
      params: { limit }
    });
    return response.data.logs || [];
  }

  // 批量操作
  async createMultipleLocalForwards(
    hostId: number,
    request: CreateMultipleLocalForwardsRequest
  ): Promise<Tunnel[]> {
    const response = await apiClient.post(`/tunnels/by-host/${hostId}/multiple`, request);
    return response.data.tunnels || [];
  }

  async createDynamicSOCKS5Tunnel(
    hostId: number,
    request: CreateDynamicSOCKS5Request
  ): Promise<Tunnel> {
    const response = await apiClient.post(`/tunnels/by-host/${hostId}/socks5`, request);
    return response.data.tunnel;
  }

  async startAutoTunnels(): Promise<void> {
    await apiClient.post('/tunnels/auto-start');
  }

  async stopAllTunnels(): Promise<void> {
    await apiClient.post('/tunnels/stop-all');
  }

  // 服务健康检查
  async checkServiceHealth(request: HealthCheckRequest): Promise<HealthCheckResponse> {
    const response = await apiClient.post('/service/health-check', request);
    return response.data;
  }

  // 端口管理
  async findAvailablePort(request: FindAvailablePortRequest): Promise<FindAvailablePortResponse> {
    const response = await apiClient.post('/port/find-available', request);
    return response.data;
  }
}

export const tunnelApi = new TunnelApi();