import type { Stats } from '../types';
import '../styles/StatsPanel.css';

interface StatsPanelProps {
  stats: Stats | null;
}

export function StatsPanel({ stats }: StatsPanelProps) {
  if (!stats) {
    return (
      <div className="stats-panel">
        <div className="stats-loading">Loading statistics...</div>
      </div>
    );
  }

  const successRate =
    stats.llm_analyses_total > 0
      ? ((stats.llm_analyses_success / stats.llm_analyses_total) * 100).toFixed(1)
      : '0.0';

  return (
    <div className="stats-panel">
      <h2 className="stats-title">System Statistics</h2>

      <div className="stats-grid">
        <div className="stat-card">
          <div className="stat-label">Total Queries</div>
          <div className="stat-value">{stats.total_queries.toLocaleString()}</div>
        </div>

        <div className="stat-card">
          <div className="stat-label">Unique Clients</div>
          <div className="stat-value">{stats.unique_clients}</div>
        </div>

        <div className="stat-card">
          <div className="stat-label">Total Anomalies</div>
          <div className="stat-value">{stats.total_anomalies}</div>
        </div>

        <div className="stat-card highlight-pending">
          <div className="stat-label">Pending Review</div>
          <div className="stat-value">{stats.pending_anomalies}</div>
        </div>

        <div className="stat-card">
          <div className="stat-label">Approved</div>
          <div className="stat-value">{stats.approved_anomalies}</div>
        </div>

        <div className="stat-card">
          <div className="stat-label">Blocked</div>
          <div className="stat-value">{stats.blocked_anomalies}</div>
        </div>

        <div className="stat-card danger">
          <div className="stat-label">Malicious</div>
          <div className="stat-value">{stats.malicious_count}</div>
        </div>

        <div className="stat-card warning">
          <div className="stat-label">Suspicious</div>
          <div className="stat-value">{stats.suspicious_count}</div>
        </div>

        <div className="stat-card">
          <div className="stat-label">LLM Analyses</div>
          <div className="stat-value">{stats.llm_analyses_total}</div>
        </div>

        <div className="stat-card success">
          <div className="stat-label">Success Rate</div>
          <div className="stat-value">{successRate}%</div>
        </div>
      </div>
    </div>
  );
}
