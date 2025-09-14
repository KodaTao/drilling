interface HeaderProps {
  currentView: string
  onNavigate: (view: string) => void
}

const Header = ({ currentView, onNavigate }: HeaderProps) => {
  const linkStyle = (isActive: boolean): React.CSSProperties => ({
    marginRight: '1rem',
    padding: '0.5rem 1rem',
    textDecoration: 'none',
    color: isActive ? '#3b82f6' : '#64748b',
    backgroundColor: isActive ? '#f1f5f9' : 'transparent',
    borderRadius: '4px',
    cursor: 'pointer',
    border: 'none',
    fontSize: '0.875rem',
    fontWeight: isActive ? '600' : '400',
  })

  return (
    <header style={{
      padding: '1rem 2rem',
      borderBottom: '1px solid #e2e8f0',
      backgroundColor: 'white',
      boxShadow: '0 1px 3px rgba(0, 0, 0, 0.1)'
    }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <h1 style={{ margin: 0, fontSize: '1.5rem', fontWeight: '700', color: '#1e293b' }}>
          Drilling Platform
        </h1>
        <nav style={{ display: 'flex', alignItems: 'center' }}>
          <button
            onClick={() => onNavigate('dashboard')}
            style={linkStyle(currentView === 'dashboard')}
          >
            Dashboard
          </button>
          <button
            onClick={() => onNavigate('hosts')}
            style={linkStyle(currentView === 'hosts')}
          >
            Hosts
          </button>
          <button
            onClick={() => onNavigate('tunnels')}
            style={linkStyle(currentView === 'tunnels')}
          >
            Tunnels
          </button>
          <button
            onClick={() => onNavigate('export')}
            style={linkStyle(currentView === 'export')}
          >
            Export
          </button>
          <button
            onClick={() => onNavigate('settings')}
            style={linkStyle(currentView === 'settings')}
          >
            Settings
          </button>
        </nav>
      </div>
    </header>
  )
}

export default Header