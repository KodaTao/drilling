import { useState, useEffect } from 'react'
import { Host } from '../types'
import { hostApi } from '../api/hostApi'

interface HostListProps {
  onEditHost: (host: Host) => void
  onDeleteHost: (host: Host) => void
  refreshTrigger: number
}

const HostList = ({ onEditHost, onDeleteHost, refreshTrigger }: HostListProps) => {
  const [hosts, setHosts] = useState<Host[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [testingConnections, setTestingConnections] = useState<Set<number>>(new Set())

  const loadHosts = async () => {
    try {
      setLoading(true)
      setError(null)
      const hostsData = await hostApi.getAllHosts()
      setHosts(hostsData)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load hosts')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadHosts()
  }, [refreshTrigger])

  const handleTestConnection = async (host: Host) => {
    if (testingConnections.has(host.id)) return

    setTestingConnections(prev => new Set(prev).add(host.id))

    try {
      await hostApi.testConnection(host.id)
      // 重新加载主机列表以获取更新的状态
      loadHosts()
    } catch (err) {
      alert(`Connection test failed: ${err instanceof Error ? err.message : 'Unknown error'}`)
    } finally {
      setTestingConnections(prev => {
        const newSet = new Set(prev)
        newSet.delete(host.id)
        return newSet
      })
    }
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'active':
        return '#22c55e'
      case 'error':
        return '#ef4444'
      default:
        return '#94a3b8'
    }
  }

  const getStatusText = (status: string) => {
    switch (status) {
      case 'active':
        return 'Connected'
      case 'error':
        return 'Error'
      default:
        return 'Inactive'
    }
  }

  if (loading) {
    return <div style={{ textAlign: 'center', padding: '2rem' }}>Loading hosts...</div>
  }

  if (error) {
    return (
      <div style={{ color: '#ef4444', textAlign: 'center', padding: '2rem' }}>
        <p>Error: {error}</p>
        <button onClick={loadHosts} style={{ marginTop: '1rem', padding: '0.5rem 1rem' }}>
          Retry
        </button>
      </div>
    )
  }

  if (hosts.length === 0) {
    return (
      <div style={{ textAlign: 'center', padding: '2rem', color: '#64748b' }}>
        <p>No hosts configured yet.</p>
        <p>Click "Add Host" to get started.</p>
      </div>
    )
  }

  return (
    <div style={{ padding: '1rem' }}>
      <div style={{ display: 'grid', gap: '1rem' }}>
        {hosts.map((host) => (
          <div
            key={host.id}
            style={{
              border: '1px solid #e2e8f0',
              borderRadius: '8px',
              padding: '1.5rem',
              backgroundColor: 'white',
              boxShadow: '0 1px 3px rgba(0, 0, 0, 0.1)',
            }}
          >
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '1rem' }}>
              <div>
                <h3 style={{ margin: '0 0 0.5rem 0', fontSize: '1.25rem', fontWeight: '600' }}>
                  {host.name}
                </h3>
                <p style={{ margin: '0 0 0.5rem 0', color: '#64748b' }}>
                  {host.username}@{host.hostname}:{host.port}
                </p>
                {host.description && (
                  <p style={{ margin: '0', color: '#64748b', fontSize: '0.875rem' }}>
                    {host.description}
                  </p>
                )}
              </div>
              <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                <span
                  style={{
                    display: 'inline-block',
                    width: '8px',
                    height: '8px',
                    borderRadius: '50%',
                    backgroundColor: getStatusColor(host.status),
                  }}
                ></span>
                <span style={{ fontSize: '0.875rem', color: '#374151' }}>
                  {getStatusText(host.status)}
                </span>
              </div>
            </div>

            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <div style={{ fontSize: '0.75rem', color: '#9ca3af' }}>
                Auth: {host.auth_type.replace('_', ' + ')}
                {host.last_check && (
                  <span style={{ marginLeft: '1rem' }}>
                    Last checked: {new Date(host.last_check).toLocaleString()}
                  </span>
                )}
              </div>

              <div style={{ display: 'flex', gap: '0.5rem' }}>
                <button
                  onClick={() => handleTestConnection(host)}
                  disabled={testingConnections.has(host.id)}
                  style={{
                    padding: '0.375rem 0.75rem',
                    fontSize: '0.875rem',
                    border: '1px solid #d1d5db',
                    borderRadius: '4px',
                    backgroundColor: testingConnections.has(host.id) ? '#f3f4f6' : 'white',
                    cursor: testingConnections.has(host.id) ? 'not-allowed' : 'pointer',
                    opacity: testingConnections.has(host.id) ? 0.6 : 1,
                  }}
                >
                  {testingConnections.has(host.id) ? 'Testing...' : 'Test'}
                </button>
                <button
                  onClick={() => onEditHost(host)}
                  style={{
                    padding: '0.375rem 0.75rem',
                    fontSize: '0.875rem',
                    border: '1px solid #3b82f6',
                    borderRadius: '4px',
                    backgroundColor: '#3b82f6',
                    color: 'white',
                    cursor: 'pointer',
                  }}
                >
                  Edit
                </button>
                <button
                  onClick={() => onDeleteHost(host)}
                  style={{
                    padding: '0.375rem 0.75rem',
                    fontSize: '0.875rem',
                    border: '1px solid #ef4444',
                    borderRadius: '4px',
                    backgroundColor: '#ef4444',
                    color: 'white',
                    cursor: 'pointer',
                  }}
                >
                  Delete
                </button>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}

export default HostList