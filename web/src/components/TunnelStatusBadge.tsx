import React from 'react';

interface TunnelStatusBadgeProps {
  status: string;
  size?: 'sm' | 'md' | 'lg';
}

const TunnelStatusBadge: React.FC<TunnelStatusBadgeProps> = ({ status, size = 'md' }) => {
  const getStatusConfig = (status: string) => {
    switch (status.toLowerCase()) {
      case 'active':
        return {
          className: 'status-active',
          label: 'Active',
          color: '#28a745'
        };
      case 'inactive':
        return {
          className: 'status-inactive',
          label: 'Inactive',
          color: '#6c757d'
        };
      case 'error':
        return {
          className: 'status-error',
          label: 'Error',
          color: '#dc3545'
        };
      default:
        return {
          className: 'status-unknown',
          label: status,
          color: '#ffc107'
        };
    }
  };

  const config = getStatusConfig(status);

  return (
    <span
      className={`status-badge ${config.className} size-${size}`}
      style={{ backgroundColor: config.color }}
    >
      <span className="status-dot"></span>
      {config.label}
    </span>
  );
};

export default TunnelStatusBadge;