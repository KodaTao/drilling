import React, { useState } from 'react';
import { Tunnel } from '../api/tunnelApi';
import { Host } from '../api/hostApi';
import TunnelStatusBadge from './TunnelStatusBadge';
import TunnelLogs from './TunnelLogs';

interface TunnelListProps {
  tunnels: Tunnel[];
  hosts: Host[];
  onEdit: (tunnel: Tunnel) => void;
  onDelete: (id: number) => void;
  onStart: (id: number) => void;
  onStop: (id: number) => void;
  onRestart: (id: number) => void;
}

const TunnelList: React.FC<TunnelListProps> = ({
  tunnels,
  hosts,
  onEdit,
  onDelete,
  onStart,
  onStop,
  onRestart
}) => {
  const [expandedTunnel, setExpandedTunnel] = useState<number | null>(null);
  const [showLogs, setShowLogs] = useState<number | null>(null);

  const getHostName = (hostId: number) => {
    const host = hosts.find(h => h.id === hostId);
    return host ? `${host.name} (${host.hostname})` : `Host ${hostId}`;
  };

  const getTunnelTypeDisplay = (type: string) => {
    switch (type) {
      case 'local_forward':
        return 'Local Forward';
      case 'remote_forward':
        return 'Remote Forward';
      case 'dynamic':
        return 'SOCKS5 Proxy';
      default:
        return type;
    }
  };

  const formatTunnelDescription = (tunnel: Tunnel) => {
    switch (tunnel.type) {
      case 'local_forward':
        return `${tunnel.local_address}:${tunnel.local_port} ← ${tunnel.remote_address}:${tunnel.remote_port}`;
      case 'remote_forward':
        return `${tunnel.local_address}:${tunnel.local_port} → ${tunnel.remote_address}:${tunnel.remote_port}`;
      case 'dynamic':
        return `SOCKS5 Proxy on ${tunnel.local_address}:${tunnel.local_port}`;
      default:
        return `${tunnel.local_address}:${tunnel.local_port}`;
    }
  };

  const toggleExpanded = (tunnelId: number) => {
    setExpandedTunnel(expandedTunnel === tunnelId ? null : tunnelId);
  };

  const toggleLogs = (tunnelId: number) => {
    setShowLogs(showLogs === tunnelId ? null : tunnelId);
  };

  if (tunnels.length === 0) {
    return (
      <div className="tunnel-list-empty">
        <p>No tunnels found. Create your first tunnel to get started.</p>
      </div>
    );
  }

  return (
    <div className="tunnel-list">
      {tunnels.map(tunnel => (
        <div key={tunnel.id} className="tunnel-card">
          <div className="tunnel-card-header" onClick={() => toggleExpanded(tunnel.id)}>
            <div className="tunnel-info">
              <h3 className="tunnel-name">{tunnel.name}</h3>
              <div className="tunnel-details">
                <span className="tunnel-type">{getTunnelTypeDisplay(tunnel.type)}</span>
                <span className="tunnel-description">
                  {formatTunnelDescription(tunnel)}
                </span>
                <span className="tunnel-host">Host: {getHostName(tunnel.host_id)}</span>
              </div>
            </div>
            <div className="tunnel-status">
              <TunnelStatusBadge status={tunnel.status} />
              {tunnel.auto_start && (
                <span className="auto-start-badge">Auto Start</span>
              )}
            </div>
            <div className="tunnel-actions">
              {tunnel.status === 'active' ? (
                <>
                  <button
                    onClick={(e) => {
                      e.stopPropagation();
                      onStop(tunnel.id);
                    }}
                    className="btn btn-sm btn-warning"
                  >
                    Stop
                  </button>
                  <button
                    onClick={(e) => {
                      e.stopPropagation();
                      onRestart(tunnel.id);
                    }}
                    className="btn btn-sm btn-secondary"
                  >
                    Restart
                  </button>
                </>
              ) : (
                <button
                  onClick={(e) => {
                    e.stopPropagation();
                    onStart(tunnel.id);
                  }}
                  className="btn btn-sm btn-success"
                >
                  Start
                </button>
              )}
              <button
                onClick={(e) => {
                  e.stopPropagation();
                  onEdit(tunnel);
                }}
                className="btn btn-sm btn-primary"
              >
                Edit
              </button>
              <button
                onClick={(e) => {
                  e.stopPropagation();
                  onDelete(tunnel.id);
                }}
                className="btn btn-sm btn-danger"
              >
                Delete
              </button>
            </div>
          </div>

          {expandedTunnel === tunnel.id && (
            <div className="tunnel-card-expanded">
              <div className="tunnel-details-expanded">
                <div className="detail-row">
                  <label>Description:</label>
                  <span>{tunnel.description || 'No description'}</span>
                </div>
                <div className="detail-row">
                  <label>Created:</label>
                  <span>{new Date(tunnel.created_at).toLocaleString()}</span>
                </div>
                <div className="detail-row">
                  <label>Last Updated:</label>
                  <span>{new Date(tunnel.updated_at).toLocaleString()}</span>
                </div>
                {tunnel.type !== 'dynamic' && (
                  <>
                    <div className="detail-row">
                      <label>Local Address:</label>
                      <span>{tunnel.local_address}:{tunnel.local_port}</span>
                    </div>
                    <div className="detail-row">
                      <label>Remote Address:</label>
                      <span>{tunnel.remote_address}:{tunnel.remote_port}</span>
                    </div>
                  </>
                )}
                {tunnel.type === 'dynamic' && (
                  <div className="detail-row">
                    <label>Proxy Address:</label>
                    <span>{tunnel.local_address}:{tunnel.local_port}</span>
                  </div>
                )}
                <div className="detail-row">
                  <label>Auto Start:</label>
                  <span>{tunnel.auto_start ? 'Enabled' : 'Disabled'}</span>
                </div>
              </div>
              <div className="tunnel-logs-section">
                <button
                  onClick={() => toggleLogs(tunnel.id)}
                  className="btn btn-sm btn-secondary"
                >
                  {showLogs === tunnel.id ? 'Hide Logs' : 'Show Logs'}
                </button>
                {showLogs === tunnel.id && (
                  <TunnelLogs tunnelId={tunnel.id} />
                )}
              </div>
            </div>
          )}
        </div>
      ))}
    </div>
  );
};

export default TunnelList;