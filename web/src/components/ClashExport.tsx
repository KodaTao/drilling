import React, { useState, useEffect } from 'react'
import { exportApi } from '../api/exportApi'
import { ClashExportResponse, Socks5StatusResponse } from '../types'

const ClashExport: React.FC = () => {
  const [socks5Status, setSocks5Status] = useState<Socks5StatusResponse | null>(null)
  const [clashPreview, setClashPreview] = useState<ClashExportResponse | null>(null)
  const [loading, setLoading] = useState(false)
  const [downloading, setDownloading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [showPreview, setShowPreview] = useState(false)

  // 样式定义
  const cardStyle: React.CSSProperties = {
    backgroundColor: 'white',
    border: '1px solid #e2e8f0',
    borderRadius: '8px',
    padding: '1.5rem',
    marginBottom: '1.5rem',
    boxShadow: '0 1px 3px rgba(0, 0, 0, 0.1)'
  }

  const buttonStyle: React.CSSProperties = {
    padding: '0.5rem 1rem',
    borderRadius: '6px',
    border: '1px solid #d1d5db',
    backgroundColor: 'white',
    cursor: 'pointer',
    fontSize: '0.875rem',
    marginRight: '0.75rem',
    display: 'inline-flex',
    alignItems: 'center',
    gap: '0.5rem'
  }

  const primaryButtonStyle: React.CSSProperties = {
    ...buttonStyle,
    backgroundColor: '#3b82f6',
    color: 'white',
    border: '1px solid #3b82f6'
  }

  const badgeStyle: React.CSSProperties = {
    padding: '0.25rem 0.75rem',
    borderRadius: '9999px',
    fontSize: '0.75rem',
    fontWeight: '500',
    backgroundColor: '#f1f5f9',
    color: '#475569'
  }

  const alertStyle: React.CSSProperties = {
    padding: '1rem',
    borderRadius: '6px',
    marginBottom: '1rem',
    border: '1px solid #d1d5db',
    backgroundColor: '#f8fafc'
  }

  const errorAlertStyle: React.CSSProperties = {
    ...alertStyle,
    backgroundColor: '#fef2f2',
    border: '1px solid #fecaca',
    color: '#dc2626'
  }

  // 获取SOCKS5状态
  const fetchSocks5Status = async () => {
    try {
      setLoading(true)
      setError(null)
      const status = await exportApi.getSocks5Status()
      setSocks5Status(status)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch SOCKS5 status')
      setSocks5Status(null)
    } finally {
      setLoading(false)
    }
  }

  // 获取Clash配置预览
  const fetchClashPreview = async () => {
    try {
      setLoading(true)
      setError(null)
      const preview = await exportApi.getClashConfigPreview()
      setClashPreview(preview)
      setShowPreview(true)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch Clash config preview')
      setClashPreview(null)
    } finally {
      setLoading(false)
    }
  }

  // 下载Clash配置
  const handleDownloadConfig = async () => {
    try {
      setDownloading(true)
      setError(null)
      await exportApi.downloadClashConfig()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to download Clash config')
    } finally {
      setDownloading(false)
    }
  }

  // 刷新数据
  const handleRefresh = () => {
    fetchSocks5Status()
    if (showPreview) {
      fetchClashPreview()
    }
  }

  // 初始化加载
  useEffect(() => {
    fetchSocks5Status()
  }, [])

  return (
    <div>
      <div style={cardStyle}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
          <h2 style={{ margin: 0, fontSize: '1.25rem', fontWeight: '600', color: '#1e293b' }}>
            🛡️ Clash 配置导出
          </h2>
          <button
            style={buttonStyle}
            onClick={handleRefresh}
            disabled={loading}
          >
            {loading ? '🔄' : '↻'} 刷新状态
          </button>
        </div>

        {error && (
          <div style={errorAlertStyle}>
            ⚠️ {error}
          </div>
        )}

        {/* SOCKS5状态 */}
        {socks5Status && (
          <div style={{ marginBottom: '1.5rem' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '1rem', border: '1px solid #e2e8f0', borderRadius: '6px', marginBottom: '1rem' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                  <span style={{ fontSize: '1.25rem' }}>{socks5Status.can_export ? '✅' : '❌'}</span>
                  <span style={{ fontWeight: '500' }}>SOCKS5 隧道状态</span>
                </div>
                <span style={badgeStyle}>
                  {socks5Status.active_count} 个活跃隧道
                </span>
              </div>
              <div style={{ fontSize: '0.875rem', color: '#64748b' }}>
                最后检查: {new Date(socks5Status.last_check).toLocaleString()}
              </div>
            </div>

            {/* 活跃隧道列表 */}
            {socks5Status.tunnels.length > 0 && (
              <div style={{ marginBottom: '1.5rem' }}>
                <h4 style={{ margin: '0 0 0.75rem 0', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                  ⚡ 活跃的 SOCKS5 隧道
                </h4>
                <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
                  {socks5Status.tunnels.map((tunnel) => (
                    <div key={tunnel.id} style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '0.75rem', backgroundColor: '#f8fafc', borderRadius: '6px' }}>
                      <div>
                        <div style={{ fontWeight: '500' }}>{tunnel.name}</div>
                        <div style={{ fontSize: '0.875rem', color: '#64748b' }}>
                          {tunnel.local_address}:{tunnel.local_port}
                        </div>
                      </div>
                      <span style={badgeStyle}>{tunnel.status}</span>
                    </div>
                  ))}
                </div>
              </div>
            )}

            {/* 操作按钮 */}
            <div style={{ display: 'flex', gap: '0.75rem', marginBottom: '1rem' }}>
              <button
                style={buttonStyle}
                onClick={fetchClashPreview}
                disabled={!socks5Status.can_export || loading}
              >
                👁️ 预览配置
              </button>
              <button
                style={socks5Status.can_export && !downloading ? primaryButtonStyle : { ...buttonStyle, opacity: 0.5 }}
                onClick={handleDownloadConfig}
                disabled={!socks5Status.can_export || downloading}
              >
                {downloading ? '⬇️ 下载中...' : '💾 下载配置'}
              </button>
            </div>

            {/* 提示信息 */}
            {!socks5Status.can_export ? (
              <div style={alertStyle}>
                ⚠️ 没有找到活跃的 SOCKS5 隧道。请先启动至少一个 SOCKS5 类型的隧道，然后再导出 Clash 配置。
              </div>
            ) : (
              <div style={alertStyle}>
                <div style={{ marginBottom: '0.5rem' }}>
                  ✅ 找到 {socks5Status.active_count} 个活跃的 SOCKS5 隧道，可以导出 Clash 配置。
                </div>
                <div style={{ fontSize: '0.875rem' }}>
                  <div style={{ fontWeight: '600', marginBottom: '0.5rem' }}>使用说明:</div>
                  <ol style={{ marginLeft: '1.5rem', margin: 0 }}>
                    <li>下载配置文件并保存为 config.yaml</li>
                    <li>将文件放置到 Clash 配置目录</li>
                    <li>启动 Clash 客户端并选择合适的代理组</li>
                    <li>配置系统代理使用 Clash (HTTP: 7890, SOCKS5: 7891)</li>
                  </ol>
                </div>
              </div>
            )}
          </div>
        )}
      </div>

      {/* Clash配置预览 */}
      {showPreview && clashPreview && (
        <div style={cardStyle}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
            <h3 style={{ margin: 0, display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
              👁️ 配置预览
            </h3>
            <button
              style={buttonStyle}
              onClick={() => setShowPreview(false)}
            >
              隐藏预览
            </button>
          </div>

          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: '1rem', marginBottom: '1.5rem' }}>
            <div style={{ padding: '1rem', backgroundColor: '#dbeafe', borderRadius: '6px' }}>
              <div style={{ fontSize: '0.875rem', color: '#1e40af', fontWeight: '500' }}>代理节点</div>
              <div style={{ fontSize: '2rem', fontWeight: '700', color: '#1e3a8a' }}>{clashPreview.config.proxies.length}</div>
            </div>
            <div style={{ padding: '1rem', backgroundColor: '#dcfce7', borderRadius: '6px' }}>
              <div style={{ fontSize: '0.875rem', color: '#166534', fontWeight: '500' }}>代理组</div>
              <div style={{ fontSize: '2rem', fontWeight: '700', color: '#14532d' }}>{clashPreview.config['proxy-groups'].length}</div>
            </div>
            <div style={{ padding: '1rem', backgroundColor: '#f3e8ff', borderRadius: '6px' }}>
              <div style={{ fontSize: '0.875rem', color: '#7c3aed', fontWeight: '500' }}>规则数量</div>
              <div style={{ fontSize: '2rem', fontWeight: '700', color: '#581c87' }}>{clashPreview.config.rules.length}</div>
            </div>
          </div>

          {/* 代理节点列表 */}
          <div style={{ marginBottom: '1.5rem' }}>
            <h4 style={{ margin: '0 0 0.75rem 0' }}>代理节点:</h4>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
              {clashPreview.config.proxies.map((proxy, index) => (
                <div key={index} style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '0.75rem', backgroundColor: '#f8fafc', borderRadius: '6px' }}>
                  <div>
                    <div style={{ fontWeight: '500' }}>{proxy.name}</div>
                    <div style={{ fontSize: '0.875rem', color: '#64748b' }}>{proxy.type.toUpperCase()}</div>
                  </div>
                  <div style={{ textAlign: 'right' }}>
                    <div style={{ fontFamily: 'monospace', fontSize: '0.875rem' }}>{proxy.server}:{proxy.port}</div>
                  </div>
                </div>
              ))}
            </div>
          </div>

          {/* 代理组列表 */}
          <div style={{ marginBottom: '1rem' }}>
            <h4 style={{ margin: '0 0 0.75rem 0' }}>代理组:</h4>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
              {clashPreview.config['proxy-groups'].map((group, index) => (
                <div key={index} style={{ padding: '0.75rem', backgroundColor: '#f8fafc', borderRadius: '6px' }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.5rem' }}>
                    <div style={{ fontWeight: '500' }}>{group.name}</div>
                    <span style={badgeStyle}>{group.type}</span>
                  </div>
                  <div style={{ fontSize: '0.875rem', color: '#64748b' }}>
                    包含 {group.proxies.length} 个代理: {group.proxies.join(', ')}
                  </div>
                </div>
              ))}
            </div>
          </div>

          <div style={{ fontSize: '0.875rem', color: '#64748b' }}>
            配置生成时间: {new Date(clashPreview.generated_at).toLocaleString()}
          </div>
        </div>
      )}
    </div>
  )
}

export default ClashExport