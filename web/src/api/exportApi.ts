import { ClashExportResponse, Socks5StatusResponse } from '../types'

const API_BASE = '/api/v1'

class ExportApi {
  async getSocks5Status(): Promise<Socks5StatusResponse> {
    const response = await fetch(`${API_BASE}/export/socks5/status`)
    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: 'Failed to get SOCKS5 status' }))
      throw new Error(error.details || error.error || 'Failed to get SOCKS5 status')
    }
    return response.json()
  }

  async getClashConfigPreview(): Promise<ClashExportResponse> {
    const response = await fetch(`${API_BASE}/export/clash/preview`)
    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: 'Failed to get Clash config preview' }))
      throw new Error(error.details || error.error || 'Failed to get Clash config preview')
    }
    return response.json()
  }

  async downloadClashConfig(): Promise<void> {
    const response = await fetch(`${API_BASE}/export/clash`)
    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: 'Failed to download Clash config' }))
      throw new Error(error.details || error.error || 'Failed to download Clash config')
    }

    // 获取文件名
    const contentDisposition = response.headers.get('Content-Disposition')
    let filename = 'clash-config.yaml'
    if (contentDisposition) {
      const filenameMatch = contentDisposition.match(/filename=([^;]+)/)
      if (filenameMatch) {
        filename = filenameMatch[1].replace(/['"]/g, '')
      }
    }

    // 创建下载
    const blob = await response.blob()
    const url = window.URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.style.display = 'none'
    a.href = url
    a.download = filename
    document.body.appendChild(a)
    a.click()
    window.URL.revokeObjectURL(url)
    document.body.removeChild(a)
  }
}

export const exportApi = new ExportApi()