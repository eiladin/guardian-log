# API Reference

REST API documentation for Guardian-Log.

## Base URL

```
http://localhost:8080/api
```

## Endpoints

### GET /api/health

Health check endpoint.

**Response:**
```json
{
  "status": "ok"
}
```

### GET /api/anomalies

Get list of anomalies.

**Query Parameters:**
- `status` (optional) - Filter by status: `pending`, `approved`, `blocked`

**Response:**
```json
[
  {
    "id": "uuid",
    "domain": "suspicious.example.com",
    "client_id": "192.168.1.100",
    "client_name": "iPhone",
    "classification": "Suspicious",
    "risk_score": 7,
    "explanation": "Domain registered recently...",
    "suggested_action": "Investigate",
    "detected_at": "2024-01-01T12:00:00Z",
    "status": "pending"
  }
]
```

### POST /api/anomalies/:id/approve

Approve an anomaly (adds to baseline).

**Response:**
```json
{
  "status": "approved"
}
```

### POST /api/anomalies/:id/block

Block an anomaly (adds to AdGuard blocklist).

**Response:**
```json
{
  "status": "blocked"
}
```

### GET /api/stats

Get statistics.

**Response:**
```json
{
  "total_queries": 1000,
  "total_anomalies": 50,
  "pending_anomalies": 10,
  "approved_anomalies": 30,
  "blocked_anomalies": 10,
  "total_clients": 5,
  "total_baseline_domains": 500,
  "suspicious_count": 30,
  "malicious_count": 20,
  "avg_risk_score": 6.5
}
```

## Error Responses

All endpoints may return:

**400 Bad Request:**
```json
{
  "error": "Invalid request"
}
```

**404 Not Found:**
```json
{
  "error": "Anomaly not found"
}
```

**500 Internal Server Error:**
```json
{
  "error": "Internal server error"
}
```

## CORS

CORS is enabled for all origins in development.

Production deployments should restrict origins.

## Rate Limiting

No rate limiting currently implemented.

---

**[← Back to Docs](../README.md)**
