import React, { useState, useEffect } from 'react';
import { tunnelApi, Tunnel, CreateTunnelRequest } from '../api/tunnelApi';
import { hostApi, Host } from '../api/hostApi';
import TunnelList from './TunnelList';
import TunnelForm from './TunnelForm';
import './TunnelManagement.css';

const TunnelManagement: React.FC = () => {
  const [tunnels, setTunnels] = useState<Tunnel[]>([]);
  const [hosts, setHosts] = useState<Host[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string>('');
  const [showCreateForm, setShowCreateForm] = useState(false);
  const [editingTunnel, setEditingTunnel] = useState<Tunnel | null>(null);
  const [selectedHost, setSelectedHost] = useState<string>('all');
  const [autoRefresh, setAutoRefresh] = useState(false);
  const [refreshInterval, setRefreshInterval] = useState<number | null>(null);

  // 加载数据
  const loadData = async () => {
    try {
      setLoading(true);
      const [tunnelsData, hostsData] = await Promise.all([
        tunnelApi.getAllTunnels(),
        hostApi.getAllHosts()
      ]);
      setTunnels(tunnelsData);
      setHosts(hostsData);
      setError('');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load data');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, []);

  // 自动刷新功能
  useEffect(() => {
    if (autoRefresh) {
      const interval = setInterval(() => {
        loadData();
      }, 5000); // 每5秒刷新一次
      setRefreshInterval(interval);
      return () => clearInterval(interval);
    } else if (refreshInterval) {
      clearInterval(refreshInterval);
      setRefreshInterval(null);
    }
  }, [autoRefresh]);

  // 切换自动刷新
  const toggleAutoRefresh = () => {
    setAutoRefresh(!autoRefresh);
  };

  // 创建隧道
  const handleCreateTunnel = async (tunnelData: CreateTunnelRequest) => {
    try {
      await tunnelApi.createTunnel(tunnelData);
      await loadData();
      setShowCreateForm(false);
      setError('');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create tunnel');
    }
  };

  // 更新隧道
  const handleUpdateTunnel = async (id: number, tunnelData: Partial<CreateTunnelRequest>) => {
    try {
      await tunnelApi.updateTunnel(id, tunnelData);
      await loadData();
      setEditingTunnel(null);
      setError('');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to update tunnel');
    }
  };

  // 删除隧道
  const handleDeleteTunnel = async (id: number) => {
    if (!window.confirm('Are you sure you want to delete this tunnel?')) {
      return;
    }
    try {
      await tunnelApi.deleteTunnel(id);
      await loadData();
      setError('');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete tunnel');
    }
  };

  // 启动隧道
  const handleStartTunnel = async (id: number) => {
    try {
      await tunnelApi.startTunnel(id);
      await loadData();
      setError('');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to start tunnel');
    }
  };

  // 停止隧道
  const handleStopTunnel = async (id: number) => {
    try {
      await tunnelApi.stopTunnel(id);
      await loadData();
      setError('');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to stop tunnel');
    }
  };

  // 重启隧道
  const handleRestartTunnel = async (id: number) => {
    try {
      await tunnelApi.restartTunnel(id);
      await loadData();
      setError('');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to restart tunnel');
    }
  };

  // 启动所有自动启动的隧道
  const handleStartAutoTunnels = async () => {
    try {
      await tunnelApi.startAutoTunnels();
      await loadData();
      setError('');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to start auto tunnels');
    }
  };

  // 停止所有隧道
  const handleStopAllTunnels = async () => {
    if (!window.confirm('Are you sure you want to stop all tunnels?')) {
      return;
    }
    try {
      await tunnelApi.stopAllTunnels();
      await loadData();
      setError('');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to stop all tunnels');
    }
  };

  // 过滤隧道
  const filteredTunnels = selectedHost === 'all'
    ? tunnels
    : tunnels.filter(tunnel => tunnel.host_id.toString() === selectedHost);

  if (loading) {
    return <div className="loading">Loading tunnels...</div>;
  }

  return (
    <div className="tunnel-management">
      <div className="tunnel-header">
        <h2>Tunnel Management</h2>
        <div className="tunnel-actions">
          <select
            value={selectedHost}
            onChange={(e) => setSelectedHost(e.target.value)}
            className="host-filter"
          >
            <option value="all">All Hosts</option>
            {hosts.map(host => (
              <option key={host.id} value={host.id.toString()}>
                {host.name} ({host.hostname})
              </option>
            ))}
          </select>
          <button
            onClick={() => setShowCreateForm(true)}
            className="btn btn-primary"
          >
            Create Tunnel
          </button>
          <button
            onClick={handleStartAutoTunnels}
            className="btn btn-success"
          >
            Start Auto Tunnels
          </button>
          <button
            onClick={handleStopAllTunnels}
            className="btn btn-danger"
          >
            Stop All
          </button>
          <button
            onClick={loadData}
            className="btn btn-secondary"
          >
            Refresh
          </button>
          <button
            onClick={toggleAutoRefresh}
            className={`btn ${autoRefresh ? 'btn-success' : 'btn-secondary'}`}
          >
            {autoRefresh ? 'Auto Refresh: ON' : 'Auto Refresh: OFF'}
          </button>
        </div>
      </div>

      {error && (
        <div className="error-message">
          {error}
        </div>
      )}

      <TunnelList
        tunnels={filteredTunnels}
        hosts={hosts}
        onEdit={setEditingTunnel}
        onDelete={handleDeleteTunnel}
        onStart={handleStartTunnel}
        onStop={handleStopTunnel}
        onRestart={handleRestartTunnel}
      />

      {(showCreateForm || editingTunnel) && (
        <TunnelForm
          hosts={hosts}
          tunnel={editingTunnel}
          onSubmit={editingTunnel ?
            (data) => handleUpdateTunnel(editingTunnel.id, data) :
            handleCreateTunnel
          }
          onCancel={() => {
            setShowCreateForm(false);
            setEditingTunnel(null);
          }}
        />
      )}
    </div>
  );
};

export default TunnelManagement;