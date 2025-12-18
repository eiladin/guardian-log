// API response types matching Go backend

export interface Anomaly {
  id: string;
  domain: string;
  client_id: string;
  client_name: string;
  query_type: string;
  classification: "Suspicious" | "Malicious";
  risk_score: number;
  explanation: string;
  suggested_action: "Investigate" | "Block";
  detected_at: string;
  status: "pending" | "approved" | "blocked";
}

export interface Stats {
  total_queries: number;
  unique_clients: number;
  total_anomalies: number;
  pending_anomalies: number;
  approved_anomalies: number;
  blocked_anomalies: number;
  malicious_count: number;
  suspicious_count: number;
  llm_analyses_total: number;
  llm_analyses_success: number;
  llm_analyses_failed: number;
}

export interface Settings {
  adguard_url: string;
  poll_interval: string;
  llm_enabled: boolean;
  llm_provider: string;
  gemini_model?: string;
  has_gemini_api_key?: boolean;
}
