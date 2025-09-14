import { useState, useEffect } from 'react'
import { Host } from '../types'

interface HostFormProps {
  host?: Host | null
  onSave: (host: Omit<Host, 'id' | 'created_at' | 'updated_at' | 'status' | 'last_check'>) => Promise<void>
  onCancel: () => void
  loading?: boolean
}

const HostForm = ({ host, onSave, onCancel, loading = false }: HostFormProps) => {
  const [formData, setFormData] = useState({
    name: '',
    hostname: '',
    port: 22,
    username: '',
    auth_type: 'password' as 'password' | 'key' | 'key_password',
    password: '',
    private_key: '',
    key_path: '',
    passphrase: '',
    description: '',
  })
  const [errors, setErrors] = useState<Record<string, string>>({})

  useEffect(() => {
    if (host) {
      setFormData({
        name: host.name,
        hostname: host.hostname,
        port: host.port,
        username: host.username,
        auth_type: host.auth_type,
        password: '', // Don't populate password for security
        private_key: '', // Don't populate private key for security
        key_path: host.key_path || '',
        passphrase: '', // Don't populate passphrase for security
        description: host.description,
      })
    }
  }, [host])

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>) => {
    const { name, value } = e.target
    setFormData(prev => ({
      ...prev,
      [name]: name === 'port' ? parseInt(value) || 22 : value,
    }))
    // Clear error when user starts typing
    if (errors[name]) {
      setErrors(prev => ({ ...prev, [name]: '' }))
    }
  }

  const validateForm = () => {
    const newErrors: Record<string, string> = {}

    if (!formData.name.trim()) {
      newErrors.name = 'Name is required'
    }
    if (!formData.hostname.trim()) {
      newErrors.hostname = 'Hostname is required'
    }
    if (!formData.username.trim()) {
      newErrors.username = 'Username is required'
    }
    if (formData.port < 1 || formData.port > 65535) {
      newErrors.port = 'Port must be between 1 and 65535'
    }

    // Validate auth-specific fields
    switch (formData.auth_type) {
      case 'password':
        if (!formData.password && !host) {
          newErrors.password = 'Password is required'
        }
        break
      case 'key':
        if (!formData.private_key && !formData.key_path && !host) {
          newErrors.private_key = 'Private key or key path is required'
        }
        break
      case 'key_password':
        if (!formData.private_key && !formData.key_path && !host) {
          newErrors.private_key = 'Private key or key path is required'
        }
        if (!formData.passphrase && !host) {
          newErrors.passphrase = 'Passphrase is required'
        }
        break
    }

    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    if (!validateForm()) {
      return
    }

    try {
      // Only include non-empty sensitive fields
      const submitData = {
        ...formData,
        password: formData.password || undefined,
        private_key: formData.private_key || undefined,
        passphrase: formData.passphrase || undefined,
      }

      await onSave(submitData)
    } catch (error) {
      console.error('Form submission error:', error)
    }
  }

  const inputStyle: React.CSSProperties = {
    width: '100%',
    padding: '0.5rem',
    border: '1px solid #d1d5db',
    borderRadius: '4px',
    fontSize: '0.875rem',
  }

  const errorInputStyle: React.CSSProperties = {
    ...inputStyle,
    borderColor: '#ef4444',
  }

  const labelStyle: React.CSSProperties = {
    display: 'block',
    fontSize: '0.875rem',
    fontWeight: '500',
    marginBottom: '0.25rem',
    color: '#374151',
  }

  const errorStyle: React.CSSProperties = {
    color: '#ef4444',
    fontSize: '0.75rem',
    marginTop: '0.25rem',
  }

  return (
    <div style={{ maxWidth: '600px', margin: '0 auto', padding: '2rem' }}>
      <h2 style={{ marginBottom: '2rem', fontSize: '1.5rem', fontWeight: '600' }}>
        {host ? 'Edit Host' : 'Add New Host'}
      </h2>

      <form onSubmit={handleSubmit}>
        <div style={{ marginBottom: '1rem' }}>
          <label style={labelStyle}>
            Name <span style={{ color: '#ef4444' }}>*</span>
          </label>
          <input
            type="text"
            name="name"
            value={formData.name}
            onChange={handleInputChange}
            style={errors.name ? errorInputStyle : inputStyle}
            placeholder="My Server"
          />
          {errors.name && <div style={errorStyle}>{errors.name}</div>}
        </div>

        <div style={{ display: 'grid', gridTemplateColumns: '2fr 1fr', gap: '1rem', marginBottom: '1rem' }}>
          <div>
            <label style={labelStyle}>
              Hostname <span style={{ color: '#ef4444' }}>*</span>
            </label>
            <input
              type="text"
              name="hostname"
              value={formData.hostname}
              onChange={handleInputChange}
              style={errors.hostname ? errorInputStyle : inputStyle}
              placeholder="example.com"
            />
            {errors.hostname && <div style={errorStyle}>{errors.hostname}</div>}
          </div>

          <div>
            <label style={labelStyle}>Port</label>
            <input
              type="number"
              name="port"
              value={formData.port}
              onChange={handleInputChange}
              style={errors.port ? errorInputStyle : inputStyle}
              min="1"
              max="65535"
            />
            {errors.port && <div style={errorStyle}>{errors.port}</div>}
          </div>
        </div>

        <div style={{ marginBottom: '1rem' }}>
          <label style={labelStyle}>
            Username <span style={{ color: '#ef4444' }}>*</span>
          </label>
          <input
            type="text"
            name="username"
            value={formData.username}
            onChange={handleInputChange}
            style={errors.username ? errorInputStyle : inputStyle}
            placeholder="root"
          />
          {errors.username && <div style={errorStyle}>{errors.username}</div>}
        </div>

        <div style={{ marginBottom: '1rem' }}>
          <label style={labelStyle}>Authentication Type</label>
          <select
            name="auth_type"
            value={formData.auth_type}
            onChange={handleInputChange}
            style={inputStyle}
          >
            <option value="password">Password</option>
            <option value="key">SSH Key</option>
            <option value="key_password">SSH Key with Passphrase</option>
          </select>
        </div>

        {formData.auth_type === 'password' && (
          <div style={{ marginBottom: '1rem' }}>
            <label style={labelStyle}>
              Password {!host && <span style={{ color: '#ef4444' }}>*</span>}
            </label>
            <input
              type="password"
              name="password"
              value={formData.password}
              onChange={handleInputChange}
              style={errors.password ? errorInputStyle : inputStyle}
              placeholder={host ? "Leave empty to keep current password" : "Enter password"}
            />
            {errors.password && <div style={errorStyle}>{errors.password}</div>}
          </div>
        )}

        {(formData.auth_type === 'key' || formData.auth_type === 'key_password') && (
          <>
            <div style={{ marginBottom: '1rem' }}>
              <label style={labelStyle}>Key Path</label>
              <input
                type="text"
                name="key_path"
                value={formData.key_path}
                onChange={handleInputChange}
                style={inputStyle}
                placeholder="/path/to/private/key (optional if private key provided below)"
              />
            </div>

            <div style={{ marginBottom: '1rem' }}>
              <label style={labelStyle}>
                Private Key {!host && !formData.key_path && <span style={{ color: '#ef4444' }}>*</span>}
              </label>
              <textarea
                name="private_key"
                value={formData.private_key}
                onChange={handleInputChange}
                style={{
                  ...inputStyle,
                  ...(errors.private_key ? { borderColor: '#ef4444' } : {}),
                  height: '120px',
                  fontFamily: 'monospace',
                  fontSize: '0.75rem',
                }}
                placeholder={host ? "Leave empty to keep current key" : "Paste private key content here (alternative to key path)"}
              />
              {errors.private_key && <div style={errorStyle}>{errors.private_key}</div>}
            </div>
          </>
        )}

        {formData.auth_type === 'key_password' && (
          <div style={{ marginBottom: '1rem' }}>
            <label style={labelStyle}>
              Passphrase {!host && <span style={{ color: '#ef4444' }}>*</span>}
            </label>
            <input
              type="password"
              name="passphrase"
              value={formData.passphrase}
              onChange={handleInputChange}
              style={errors.passphrase ? errorInputStyle : inputStyle}
              placeholder={host ? "Leave empty to keep current passphrase" : "Enter key passphrase"}
            />
            {errors.passphrase && <div style={errorStyle}>{errors.passphrase}</div>}
          </div>
        )}

        <div style={{ marginBottom: '2rem' }}>
          <label style={labelStyle}>Description</label>
          <textarea
            name="description"
            value={formData.description}
            onChange={handleInputChange}
            style={{ ...inputStyle, height: '80px' }}
            placeholder="Optional description for this host"
          />
        </div>

        <div style={{ display: 'flex', gap: '1rem', justifyContent: 'flex-end' }}>
          <button
            type="button"
            onClick={onCancel}
            disabled={loading}
            style={{
              padding: '0.75rem 1.5rem',
              border: '1px solid #d1d5db',
              borderRadius: '4px',
              backgroundColor: 'white',
              cursor: loading ? 'not-allowed' : 'pointer',
              opacity: loading ? 0.6 : 1,
            }}
          >
            Cancel
          </button>
          <button
            type="submit"
            disabled={loading}
            style={{
              padding: '0.75rem 1.5rem',
              border: 'none',
              borderRadius: '4px',
              backgroundColor: '#3b82f6',
              color: 'white',
              cursor: loading ? 'not-allowed' : 'pointer',
              opacity: loading ? 0.6 : 1,
            }}
          >
            {loading ? 'Saving...' : (host ? 'Update Host' : 'Add Host')}
          </button>
        </div>
      </form>
    </div>
  )
}

export default HostForm