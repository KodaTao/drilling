import React, { useState, useEffect } from 'react';
import { tunnelApi, ConnectionLog } from '../api/tunnelApi';

interface TunnelLogsProps {
  tunnelId: number;
  maxLogs?: number;
  autoRefresh?: boolean;
  refreshInterval?: number;
}

const TunnelLogs: React.FC<TunnelLogsProps> = ({
  tunnelId,
  maxLogs = 50,
  autoRefresh = false,
  refreshInterval = 5000
}) => {
  const [logs, setLogs] = useState<ConnectionLog[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string>('');

  const loadLogs = async () => {
    try {
      const logsData = await tunnelApi.getConnectionLogs(tunnelId, maxLogs);
      setLogs(logsData);
      setError('');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load logs');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadLogs();

    if (autoRefresh) {
      const interval = setInterval(loadLogs, refreshInterval);
      return () => clearInterval(interval);
    }
  }, [tunnelId, maxLogs, autoRefresh, refreshInterval]);

  const getLogEventIcon = (eventType: string) => {
    switch (eventType.toLowerCase()) {
      case 'connect':
        return '🔗';
      case 'disconnect':
        return '❌';
      case 'error':
        return '⚠️';
      case 'start':
        return '▶️';
      case 'stop':
        return '⏹️';
      default:
        return '📝';
    }
  };

  const getLogEventClass = (eventType: string) => {
    switch (eventType.toLowerCase()) {
      case 'connect':
        return 'log-connect';
      case 'disconnect':
        return 'log-disconnect';
      case 'error':
        return 'log-error';
      case 'start':
        return 'log-start';
      case 'stop':
        return 'log-stop';
      default:
        return 'log-default';
    }
  };

  const formatTimestamp = (timestamp: string) => {
    const date = new Date(timestamp);
    return date.toLocaleString();
  };

  if (loading) {
    return <div className="tunnel-logs-loading">Loading logs...</div>;
  }

  if (error) {
    return (
      <div className="tunnel-logs-error">
        <p>Error loading logs: {error}</p>
        <button onClick={loadLogs} className="btn btn-sm btn-primary">
          Retry
        </button>
      </div>
    );
  }

  if (logs.length === 0) {
    return (
      <div className="tunnel-logs-empty">
        <p>No logs available for this tunnel.</p>
      </div>
    );
  }

  return (
    <div className="tunnel-logs">
      <div className="tunnel-logs-header">
        <h4>Connection Logs</h4>
        <div className="tunnel-logs-actions">
          <button onClick={loadLogs} className="btn btn-sm btn-secondary">
            Refresh
          </button>
        </div>
      </div>
      <div className="tunnel-logs-content">
        {logs.map(log => (
          <div key={log.id} className={`tunnel-log-entry ${getLogEventClass(log.event_type)}`}>
            <div className="log-timestamp">
              {formatTimestamp(log.timestamp)}
            </div>
            <div className="log-event">
              <span className="log-icon">
                {getLogEventIcon(log.event_type)}
              </span>
              <span className="log-type">
                {log.event_type.toUpperCase()}
              </span>
            </div>
            <div className="log-message">
              {log.message}
            </div>
          </div>
        ))}
      </div>
      {logs.length >= maxLogs && (
        <div className="tunnel-logs-footer">
          <p>Showing last {maxLogs} entries. Older logs may be available.</p>
        </div>
      )}
    </div>
  );
};

export default TunnelLogs;