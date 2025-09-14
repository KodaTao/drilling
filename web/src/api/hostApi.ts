import { Host, ConnectionTestResponse, StatusCheckResponse } from '../types'

export type { Host } from '../types'

const API_BASE = '/api/v1'

class HostApi {
  async getAllHosts(): Promise<Host[]> {
    const response = await fetch(`${API_BASE}/hosts`)
    if (!response.ok) {
      throw new Error('Failed to fetch hosts')
    }
    const data = await response.json()
    // 后端返回格式：{ "hosts": [...], "count": N }
    return data.hosts || []
  }

  async getHost(id: number): Promise<Host> {
    const response = await fetch(`${API_BASE}/hosts/${id}`)
    if (!response.ok) {
      throw new Error('Failed to fetch host')
    }
    const data = await response.json()
    // 后端返回格式：{ "host": {...} }
    return data.host
  }

  async createHost(host: Omit<Host, 'id' | 'created_at' | 'updated_at' | 'status' | 'last_check'>): Promise<Host> {
    const response = await fetch(`${API_BASE}/hosts`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(host),
    })

    if (!response.ok) {
      const error = await response.json()
      throw new Error(error.details || error.error || 'Failed to create host')
    }

    const data = await response.json()
    // 后端返回格式：{ "message": "...", "host": {...} }
    return data.host
  }

  async updateHost(id: number, host: Partial<Host>): Promise<Host> {
    const response = await fetch(`${API_BASE}/hosts/${id}`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(host),
    })

    if (!response.ok) {
      const error = await response.json()
      throw new Error(error.details || error.error || 'Failed to update host')
    }

    const data = await response.json()
    // 后端返回格式：{ "message": "...", "host": {...} }
    return data.host
  }

  async deleteHost(id: number): Promise<void> {
    const response = await fetch(`${API_BASE}/hosts/${id}`, {
      method: 'DELETE',
    })

    if (!response.ok) {
      const error = await response.json()
      throw new Error(error.details || error.error || 'Failed to delete host')
    }
  }

  async testConnection(id: number): Promise<ConnectionTestResponse> {
    const response = await fetch(`${API_BASE}/hosts/${id}/test`, {
      method: 'POST',
    })

    const data: ConnectionTestResponse = await response.json()

    if (!response.ok) {
      throw new Error(data.details || data.message || 'Connection test failed')
    }

    return data
  }

  async checkStatus(id: number): Promise<StatusCheckResponse> {
    const response = await fetch(`${API_BASE}/hosts/${id}/status`, {
      method: 'POST',
    })

    const data: StatusCheckResponse = await response.json()

    if (!response.ok) {
      throw new Error(data.message || 'Status check failed')
    }

    return data
  }
}

export const hostApi = new HostApi()