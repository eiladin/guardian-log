import type { Anomaly, Stats, Settings } from './types';

// Use relative URL so it works both in dev (with Vite proxy) and production (served by Go)
const API_BASE_URL = '/api';

export class GuardianAPI {
  // Get all anomalies, optionally filtered by status
  static async getAnomalies(status?: string): Promise<Anomaly[]> {
    const url = status
      ? `${API_BASE_URL}/anomalies?status=${status}`
      : `${API_BASE_URL}/anomalies`;

    const response = await fetch(url);
    if (!response.ok) {
      throw new Error(`Failed to fetch anomalies: ${response.statusText}`);
    }
    return response.json();
  }

  // Approve an anomaly (adds to baseline)
  static async approveAnomaly(id: string): Promise<void> {
    const response = await fetch(`${API_BASE_URL}/anomalies/${encodeURIComponent(id)}/approve`, {
      method: 'POST',
    });
    if (!response.ok) {
      throw new Error(`Failed to approve anomaly: ${response.statusText}`);
    }
  }

  // Block an anomaly (adds to AdGuard blocklist)
  static async blockAnomaly(id: string): Promise<void> {
    const response = await fetch(`${API_BASE_URL}/anomalies/${encodeURIComponent(id)}/block`, {
      method: 'POST',
    });
    if (!response.ok) {
      throw new Error(`Failed to block anomaly: ${response.statusText}`);
    }
  }

  // Get system statistics
  static async getStats(): Promise<Stats> {
    const response = await fetch(`${API_BASE_URL}/stats`);
    if (!response.ok) {
      throw new Error(`Failed to fetch stats: ${response.statusText}`);
    }
    return response.json();
  }

  // Get current settings
  static async getSettings(): Promise<Settings> {
    const response = await fetch(`${API_BASE_URL}/settings`);
    if (!response.ok) {
      throw new Error(`Failed to fetch settings: ${response.statusText}`);
    }
    return response.json();
  }

  // Health check
  static async healthCheck(): Promise<boolean> {
    try {
      const response = await fetch(`${API_BASE_URL}/health`);
      return response.ok;
    } catch {
      return false;
    }
  }
}
