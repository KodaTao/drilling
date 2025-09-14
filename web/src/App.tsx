import { useState } from 'react'
import Header from './components/Header'
import Dashboard from './pages/Dashboard'
import HostManagement from './pages/HostManagement'
import TunnelManagement from './components/TunnelManagement'
import ClashExport from './components/ClashExport'

function App() {
  const [currentView, setCurrentView] = useState('dashboard')

  const renderCurrentView = () => {
    switch (currentView) {
      case 'hosts':
        return <HostManagement />
      case 'tunnels':
        return <TunnelManagement />
      case 'export':
        return (
          <div style={{ padding: '2rem', maxWidth: '1200px', margin: '0 auto' }}>
            <ClashExport />
          </div>
        )
      case 'settings':
        return <div style={{ padding: '2rem', textAlign: 'center', color: '#64748b' }}>Settings coming soon...</div>
      case 'dashboard':
      default:
        return <Dashboard />
    }
  }

  return (
    <div style={{ minHeight: '100vh', backgroundColor: '#f8fafc' }}>
      <Header currentView={currentView} onNavigate={setCurrentView} />
      <main>
        {renderCurrentView()}
      </main>
    </div>
  )
}

export default App