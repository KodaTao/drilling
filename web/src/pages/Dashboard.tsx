import { useState, useEffect } from 'react'
import { Host } from '../types'
import { hostApi } from '../api/hostApi'

const Dashboard = () => {
  const [hosts, setHosts] = useState<Host[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const loadHosts = async () => {
      try {
        const hostsData = await hostApi.getAllHosts()
        setHosts(hostsData)
      } catch (error) {
        console.error('Failed to load hosts:', error)
      } finally {
        setLoading(false)
      }
    }

    loadHosts()
  }, [])

  const getHostStats = () => {
    if (loading) return { total: 0, active: 0, inactive: 0, error: 0 }

    return {
      total: hosts.length,
      active: hosts.filter(h => h.status === 'active').length,
      inactive: hosts.filter(h => h.status === 'inactive').length,
      error: hosts.filter(h => h.status === 'error').length,
    }
  }

  const stats = getHostStats()

  const cardStyle: React.CSSProperties = {
    backgroundColor: 'white',
    border: '1px solid #e2e8f0',
    padding: '1.5rem',
    borderRadius: '8px',
    boxShadow: '0 1px 3px rgba(0, 0, 0, 0.1)',
  }

  return (
    <div style={{ padding: '2rem' }}>
      <h1 style={{ marginBottom: '2rem', fontSize: '2rem', fontWeight: '700', color: '#1e293b' }}>
        Dashboard
      </h1>

      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(280px, 1fr))', gap: '1.5rem', marginBottom: '2rem' }}>
        <div style={cardStyle}>
          <h3 style={{ margin: '0 0 0.5rem 0', fontSize: '1.125rem', fontWeight: '600', color: '#374151' }}>
            Total Hosts
          </h3>
          <p style={{ margin: '0', fontSize: '2rem', fontWeight: '700', color: '#1e293b' }}>
            {loading ? '...' : stats.total}
          </p>
          <p style={{ margin: '0.5rem 0 0 0', fontSize: '0.875rem', color: '#64748b' }}>
            Configured servers
          </p>
        </div>

        <div style={cardStyle}>
          <h3 style={{ margin: '0 0 0.5rem 0', fontSize: '1.125rem', fontWeight: '600', color: '#374151' }}>
            Active Connections
          </h3>
          <div style={{ display: 'flex', alignItems: 'baseline', gap: '0.5rem' }}>
            <span style={{ fontSize: '2rem', fontWeight: '700', color: '#22c55e' }}>
              {loading ? '...' : stats.active}
            </span>
            <span style={{ fontSize: '1rem', color: '#64748b' }}>
              / {loading ? '...' : stats.total}
            </span>
          </div>
          <p style={{ margin: '0.5rem 0 0 0', fontSize: '0.875rem', color: '#64748b' }}>
            Hosts online
          </p>
        </div>

        <div style={cardStyle}>
          <h3 style={{ margin: '0 0 0.5rem 0', fontSize: '1.125rem', fontWeight: '600', color: '#374151' }}>
            Connection Issues
          </h3>
          <p style={{ margin: '0', fontSize: '2rem', fontWeight: '700', color: stats.error > 0 ? '#ef4444' : '#64748b' }}>
            {loading ? '...' : stats.error}
          </p>
          <p style={{ margin: '0.5rem 0 0 0', fontSize: '0.875rem', color: '#64748b' }}>
            {stats.error > 0 ? 'Hosts with errors' : 'All connections healthy'}
          </p>
        </div>

        <div style={cardStyle}>
          <h3 style={{ margin: '0 0 0.5rem 0', fontSize: '1.125rem', fontWeight: '600', color: '#374151' }}>
            System Status
          </h3>
          <p style={{ margin: '0', fontSize: '1.25rem', fontWeight: '600', color: '#22c55e' }}>
            Operational
          </p>
          <p style={{ margin: '0.5rem 0 0 0', fontSize: '0.875rem', color: '#64748b' }}>
            Service running normally
          </p>
        </div>
      </div>

      {!loading && hosts.length > 0 && (
        <div style={cardStyle}>
          <h3 style={{ margin: '0 0 1rem 0', fontSize: '1.125rem', fontWeight: '600', color: '#374151' }}>
            Recent Hosts
          </h3>
          <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
            {hosts.slice(0, 5).map((host) => (
              <div
                key={host.id}
                style={{
                  display: 'flex',
                  justifyContent: 'space-between',
                  alignItems: 'center',
                  padding: '0.75rem',
                  backgroundColor: '#f8fafc',
                  borderRadius: '6px',
                }}
              >
                <div>
                  <span style={{ fontWeight: '500', color: '#374151' }}>{host.name}</span>
                  <span style={{ marginLeft: '0.5rem', color: '#64748b', fontSize: '0.875rem' }}>
                    {host.hostname}
                  </span>
                </div>
                <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                  <span
                    style={{
                      display: 'inline-block',
                      width: '8px',
                      height: '8px',
                      borderRadius: '50%',
                      backgroundColor:
                        host.status === 'active' ? '#22c55e' :
                        host.status === 'error' ? '#ef4444' : '#94a3b8',
                    }}
                  ></span>
                  <span style={{ fontSize: '0.875rem', color: '#374151', textTransform: 'capitalize' }}>
                    {host.status}
                  </span>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {!loading && hosts.length === 0 && (
        <div style={{ ...cardStyle, textAlign: 'center', padding: '3rem' }}>
          <h3 style={{ margin: '0 0 1rem 0', color: '#64748b' }}>No Hosts Configured</h3>
          <p style={{ margin: '0 0 1.5rem 0', color: '#9ca3af' }}>
            Get started by adding your first SSH host
          </p>
          <button
            onClick={() => window.dispatchEvent(new CustomEvent('navigate', { detail: 'hosts' }))}
            style={{
              padding: '0.75rem 1.5rem',
              backgroundColor: '#3b82f6',
              color: 'white',
              border: 'none',
              borderRadius: '6px',
              cursor: 'pointer',
              fontSize: '0.875rem',
              fontWeight: '500',
            }}
          >
            Add First Host
          </button>
        </div>
      )}
    </div>
  )
}

export default Dashboard