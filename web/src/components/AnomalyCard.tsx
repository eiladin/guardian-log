import { useState } from 'react';
import type { Anomaly } from '../types';
import '../styles/AnomalyCard.css';

interface AnomalyCardProps {
  anomaly: Anomaly;
  onApprove: (id: string) => Promise<void>;
  onBlock: (id: string) => Promise<void>;
}

export function AnomalyCard({ anomaly, onApprove, onBlock }: AnomalyCardProps) {
  const [loading, setLoading] = useState(false);

  const handleApprove = async () => {
    setLoading(true);
    try {
      await onApprove(anomaly.id);
    } catch (error) {
      console.error('Failed to approve anomaly:', error);
      alert('Failed to approve anomaly. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  const handleBlock = async () => {
    if (!confirm(`Are you sure you want to block ${anomaly.domain}?`)) {
      return;
    }
    setLoading(true);
    try {
      await onBlock(anomaly.id);
    } catch (error) {
      console.error('Failed to block anomaly:', error);
      alert('Failed to block anomaly. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  const getRiskColor = (score: number) => {
    if (score >= 8) return '#dc2626'; // red
    if (score >= 6) return '#ea580c'; // orange
    return '#eab308'; // yellow
  };

  const getStatusBadge = (status: string) => {
    const badges = {
      pending: { text: 'Pending Review', color: '#eab308' },
      approved: { text: 'Approved', color: '#16a34a' },
      blocked: { text: 'Blocked', color: '#dc2626' },
    };
    return badges[status as keyof typeof badges] || badges.pending;
  };

  const statusBadge = getStatusBadge(anomaly.status);
  const isPending = anomaly.status === 'pending';

  return (
    <div className="anomaly-card">
      <div className="anomaly-header">
        <div className="anomaly-domain">
          <span className="domain-name">{anomaly.domain}</span>
          <span
            className="classification-badge"
            data-classification={anomaly.classification.toLowerCase()}
          >
            {anomaly.classification}
          </span>
        </div>
        <div
          className="risk-score"
          style={{ backgroundColor: getRiskColor(anomaly.risk_score) }}
        >
          {anomaly.risk_score}/10
        </div>
      </div>

      <div className="anomaly-details">
        <div className="detail-row">
          <span className="detail-label">Client:</span>
          <span className="detail-value">
            {anomaly.client_name} ({anomaly.client_id})
          </span>
        </div>
        <div className="detail-row">
          <span className="detail-label">Query Type:</span>
          <span className="detail-value">{anomaly.query_type}</span>
        </div>
        <div className="detail-row">
          <span className="detail-label">Detected:</span>
          <span className="detail-value">
            {new Date(anomaly.detected_at).toLocaleString()}
          </span>
        </div>
        <div className="detail-row">
          <span className="detail-label">Status:</span>
          <span
            className="status-badge"
            style={{ backgroundColor: statusBadge.color }}
          >
            {statusBadge.text}
          </span>
        </div>
      </div>

      <div className="anomaly-explanation">
        <strong>Analysis:</strong>
        <p>{anomaly.explanation}</p>
      </div>

      <div className="anomaly-action">
        <strong>Suggested Action:</strong> {anomaly.suggested_action}
      </div>

      {isPending && (
        <div className="anomaly-buttons">
          <button
            className="btn btn-approve"
            onClick={handleApprove}
            disabled={loading}
          >
            ✓ Approve (Add to Baseline)
          </button>
          <button
            className="btn btn-block"
            onClick={handleBlock}
            disabled={loading}
          >
            🚫 Block Domain
          </button>
        </div>
      )}
    </div>
  );
}
