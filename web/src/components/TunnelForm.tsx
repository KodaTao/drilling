import React, { useState, useEffect } from 'react';
import { Tunnel, CreateTunnelRequest, tunnelApi } from '../api/tunnelApi';
import { Host } from '../api/hostApi';

interface TunnelFormProps {
  hosts: Host[];
  tunnel?: Tunnel | null;
  onSubmit: (data: CreateTunnelRequest) => Promise<void>;
  onCancel: () => void;
}

const TunnelForm: React.FC<TunnelFormProps> = ({ hosts, tunnel, onSubmit, onCancel }) => {
  const [formData, setFormData] = useState<CreateTunnelRequest>({
    host_id: hosts.length > 0 ? hosts[0].id : 0,
    name: '',
    type: 'local_forward',
    local_address: '127.0.0.1',
    local_port: 8080,
    remote_address: '',
    remote_port: 80,
    description: '',
    auto_start: false
  });

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string>('');
  const [checking, setChecking] = useState(false);
  const [portAvailable, setPortAvailable] = useState<boolean | null>(null);

  // 如果是编辑模式，填充表单数据
  useEffect(() => {
    if (tunnel) {
      setFormData({
        host_id: tunnel.host_id,
        name: tunnel.name,
        type: tunnel.type,
        local_address: tunnel.local_address,
        local_port: tunnel.local_port,
        remote_address: tunnel.remote_address || '',
        remote_port: tunnel.remote_port || 80,
        description: tunnel.description || '',
        auto_start: tunnel.auto_start
      });
    }
  }, [tunnel]);

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>) => {
    const { name, value, type } = e.target;
    const checked = (e.target as HTMLInputElement).checked;

    setFormData(prev => ({
      ...prev,
      [name]: type === 'checkbox' ? checked : type === 'number' ? parseInt(value) || 0 : value
    }));

    // 重置端口检查状态
    if (name === 'local_port' || name === 'local_address') {
      setPortAvailable(null);
    }
  };

  // 检查端口可用性
  const checkPortAvailability = async () => {
    if (!formData.local_address || !formData.local_port) {
      return;
    }

    setChecking(true);
    try {
      const result = await tunnelApi.checkServiceHealth({
        local_address: formData.local_address,
        local_port: formData.local_port
      });
      // 对于端口检查，我们实际上希望端口是空闲的（不健康的）
      setPortAvailable(!result.healthy);
    } catch (err) {
      // 如果连接失败，说明端口是空闲的
      setPortAvailable(true);
    } finally {
      setChecking(false);
    }
  };

  // 自动查找可用端口
  const findAvailablePort = async () => {
    try {
      const result = await tunnelApi.findAvailablePort({
        start_port: formData.local_port,
        end_port: formData.local_port + 100,
        address: formData.local_address
      });
      setFormData(prev => ({
        ...prev,
        local_port: result.available_port
      }));
      setPortAvailable(true);
    } catch (err) {
      setError('No available ports found in the specified range');
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!formData.name.trim()) {
      setError('Name is required');
      return;
    }

    if (formData.type !== 'dynamic' && (!formData.remote_address || !formData.remote_port)) {
      setError('Remote address and port are required for this tunnel type');
      return;
    }

    setLoading(true);
    setError('');

    try {
      // 准备提交数据
      const submitData: CreateTunnelRequest = {
        ...formData,
        name: formData.name.trim(),
        description: formData.description?.trim()
      };

      // 对于动态隧道，不需要远程地址和端口
      if (formData.type === 'dynamic') {
        delete submitData.remote_address;
        delete submitData.remote_port;
      }

      await onSubmit(submitData);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save tunnel');
    } finally {
      setLoading(false);
    }
  };

  const isEditMode = !!tunnel;

  return (
    <div className="tunnel-form-overlay">
      <div className="tunnel-form">
        <div className="tunnel-form-header">
          <h3>{isEditMode ? 'Edit Tunnel' : 'Create New Tunnel'}</h3>
          <button onClick={onCancel} className="btn-close">×</button>
        </div>

        <form onSubmit={handleSubmit} className="tunnel-form-body">
          {error && (
            <div className="error-message">{error}</div>
          )}

          <div className="form-group">
            <label htmlFor="host_id">Host</label>
            <select
              id="host_id"
              name="host_id"
              value={formData.host_id}
              onChange={handleInputChange}
              required
            >
              {hosts.map(host => (
                <option key={host.id} value={host.id}>
                  {host.name} ({host.hostname}:{host.port})
                </option>
              ))}
            </select>
          </div>

          <div className="form-group">
            <label htmlFor="name">Name</label>
            <input
              type="text"
              id="name"
              name="name"
              value={formData.name}
              onChange={handleInputChange}
              placeholder="Enter tunnel name"
              required
            />
          </div>

          <div className="form-group">
            <label htmlFor="type">Tunnel Type</label>
            <select
              id="type"
              name="type"
              value={formData.type}
              onChange={handleInputChange}
            >
              <option value="local_forward">Local Forward (Remote → Local)</option>
              <option value="remote_forward">Remote Forward (Local → Remote)</option>
              <option value="dynamic">SOCKS5 Proxy</option>
            </select>
          </div>

          <div className="form-group">
            <label htmlFor="local_address">Local Address</label>
            <input
              type="text"
              id="local_address"
              name="local_address"
              value={formData.local_address}
              onChange={handleInputChange}
              placeholder="127.0.0.1"
              required
            />
          </div>

          <div className="form-group">
            <label htmlFor="local_port">Local Port</label>
            <div className="port-input-group">
              <input
                type="number"
                id="local_port"
                name="local_port"
                value={formData.local_port}
                onChange={handleInputChange}
                min="1"
                max="65535"
                required
              />
              <button
                type="button"
                onClick={checkPortAvailability}
                className="btn btn-sm btn-secondary"
                disabled={checking}
              >
                {checking ? 'Checking...' : 'Check'}
              </button>
              <button
                type="button"
                onClick={findAvailablePort}
                className="btn btn-sm btn-primary"
              >
                Find Available
              </button>
            </div>
            {portAvailable !== null && (
              <div className={`port-status ${portAvailable ? 'available' : 'unavailable'}`}>
                Port {formData.local_port} is {portAvailable ? 'available' : 'in use'}
              </div>
            )}
          </div>

          {formData.type !== 'dynamic' && (
            <>
              <div className="form-group">
                <label htmlFor="remote_address">Remote Address</label>
                <input
                  type="text"
                  id="remote_address"
                  name="remote_address"
                  value={formData.remote_address}
                  onChange={handleInputChange}
                  placeholder="example.com or 192.168.1.1"
                  required
                />
              </div>

              <div className="form-group">
                <label htmlFor="remote_port">Remote Port</label>
                <input
                  type="number"
                  id="remote_port"
                  name="remote_port"
                  value={formData.remote_port}
                  onChange={handleInputChange}
                  min="1"
                  max="65535"
                  required
                />
              </div>
            </>
          )}

          <div className="form-group">
            <label htmlFor="description">Description (Optional)</label>
            <textarea
              id="description"
              name="description"
              value={formData.description}
              onChange={handleInputChange}
              placeholder="Enter tunnel description"
              rows={3}
            />
          </div>

          <div className="form-group">
            <label className="checkbox-label">
              <input
                type="checkbox"
                name="auto_start"
                checked={formData.auto_start}
                onChange={handleInputChange}
              />
              Auto-start this tunnel
            </label>
          </div>

          <div className="tunnel-form-footer">
            <button
              type="button"
              onClick={onCancel}
              className="btn btn-secondary"
            >
              Cancel
            </button>
            <button
              type="submit"
              className="btn btn-primary"
              disabled={loading}
            >
              {loading ? 'Saving...' : (isEditMode ? 'Update Tunnel' : 'Create Tunnel')}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

export default TunnelForm;