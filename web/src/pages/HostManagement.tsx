import { useState } from 'react'
import { Host } from '../types'
import { hostApi } from '../api/hostApi'
import HostList from '../components/HostList'
import HostForm from '../components/HostForm'

const HostManagement = () => {
  const [currentView, setCurrentView] = useState<'list' | 'add' | 'edit'>('list')
  const [selectedHost, setSelectedHost] = useState<Host | null>(null)
  const [loading, setLoading] = useState(false)
  const [refreshTrigger, setRefreshTrigger] = useState(0)

  const handleAddHost = () => {
    setSelectedHost(null)
    setCurrentView('add')
  }

  const handleEditHost = (host: Host) => {
    setSelectedHost(host)
    setCurrentView('edit')
  }

  const handleDeleteHost = async (host: Host) => {
    if (!confirm(`Are you sure you want to delete "${host.name}"?`)) {
      return
    }

    try {
      setLoading(true)
      await hostApi.deleteHost(host.id)
      setRefreshTrigger(prev => prev + 1)
      alert(`Host "${host.name}" deleted successfully`)
    } catch (error) {
      alert(`Failed to delete host: ${error instanceof Error ? error.message : 'Unknown error'}`)
    } finally {
      setLoading(false)
    }
  }

  const handleSaveHost = async (hostData: Omit<Host, 'id' | 'created_at' | 'updated_at' | 'status' | 'last_check'>) => {
    try {
      setLoading(true)

      if (selectedHost) {
        // Update existing host
        await hostApi.updateHost(selectedHost.id, hostData)
        alert('Host updated successfully')
      } else {
        // Create new host
        await hostApi.createHost(hostData)
        alert('Host created successfully')
      }

      setCurrentView('list')
      setSelectedHost(null)
      setRefreshTrigger(prev => prev + 1)
    } catch (error) {
      alert(`Failed to save host: ${error instanceof Error ? error.message : 'Unknown error'}`)
    } finally {
      setLoading(false)
    }
  }

  const handleCancel = () => {
    setCurrentView('list')
    setSelectedHost(null)
  }

  return (
    <div style={{ padding: '2rem' }}>
      {currentView === 'list' ? (
        <>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '2rem' }}>
            <h1 style={{ margin: 0, fontSize: '2rem', fontWeight: '700' }}>Host Management</h1>
            <button
              onClick={handleAddHost}
              style={{
                padding: '0.75rem 1.5rem',
                backgroundColor: '#3b82f6',
                color: 'white',
                border: 'none',
                borderRadius: '6px',
                fontSize: '0.875rem',
                fontWeight: '500',
                cursor: 'pointer',
                display: 'flex',
                alignItems: 'center',
                gap: '0.5rem',
              }}
            >
              <span style={{ fontSize: '1rem' }}>+</span>
              Add Host
            </button>
          </div>

          <HostList
            onEditHost={handleEditHost}
            onDeleteHost={handleDeleteHost}
            refreshTrigger={refreshTrigger}
          />
        </>
      ) : (
        <HostForm
          host={selectedHost}
          onSave={handleSaveHost}
          onCancel={handleCancel}
          loading={loading}
        />
      )}
    </div>
  )
}

export default HostManagement