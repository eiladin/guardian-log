import { useState, useEffect } from 'react';
import { GuardianAPI } from './api';
import { AnomalyCard } from './components/AnomalyCard';
import { StatsPanel } from './components/StatsPanel';
import type { Anomaly, Stats } from './types';
import './App.css';

function App() {
    const [anomalies, setAnomalies] = useState<Anomaly[]>([]);
    const [stats, setStats] = useState<Stats | null>(null);
    const [filter, setFilter] = useState<'all' | 'pending' | 'approved' | 'blocked'>('pending');
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [isOnline, setIsOnline] = useState(true);

    // Fetch anomalies and stats
    const fetchData = async () => {
        try {
            const filterValue = filter === 'all' ? undefined : filter;
            const [anomaliesData, statsData] = await Promise.all([
                GuardianAPI.getAnomalies(filterValue),
                GuardianAPI.getStats(),
            ]);

            setAnomalies(anomaliesData);
            setStats(statsData);
            setError(null);
            setIsOnline(true);
        } catch (err) {
            console.error('Failed to fetch data:', err);
            setError('Failed to connect to Guardian Log API');
            setIsOnline(false);
        } finally {
            setLoading(false);
        }
    };

    // Initial load
    useEffect(() => {
        fetchData();
    }, [filter]);

    // Poll for updates every 10 seconds
    useEffect(() => {
        const interval = setInterval(fetchData, 10000);
        return () => clearInterval(interval);
    }, [filter]);

    // Handle approve action
    const handleApprove = async (id: string) => {
        await GuardianAPI.approveAnomaly(id);
        // Refresh data
        await fetchData();
    };

    // Handle block action
    const handleBlock = async (id: string) => {
        await GuardianAPI.blockAnomaly(id);
        // Refresh data
        await fetchData();
    };

    return (
        <div className="app">
            <header className="header">
                <div className="header-content">
                    <h1 className="title">
                        🛡️ Guardian Log
                        <span className="subtitle">DNS Anomaly Detection Dashboard</span>
                    </h1>
                    <div className="header-status">
                        <div className={`status-indicator ${isOnline ? 'online' : 'offline'}`}>
                            {isOnline ? '● Online' : '● Offline'}
                        </div>
                    </div>
                </div>
            </header>

            <main className="main">
                <StatsPanel stats={stats} />

                <div className="anomalies-section">
                    <div className="section-header">
                        <h2 className="section-title">Detected Anomalies</h2>
                        <div className="filter-buttons">
                            <button
                                className={`filter-btn ${filter === 'pending' ? 'active' : ''}`}
                                onClick={() => setFilter('pending')}
                            >
                                Pending ({stats?.pending_anomalies || 0})
                            </button>
                            <button
                                className={`filter-btn ${filter === 'all' ? 'active' : ''}`}
                                onClick={() => setFilter('all')}
                            >
                                All ({stats?.total_anomalies || 0})
                            </button>
                            <button
                                className={`filter-btn ${filter === 'approved' ? 'active' : ''}`}
                                onClick={() => setFilter('approved')}
                            >
                                Approved ({stats?.approved_anomalies || 0})
                            </button>
                            <button
                                className={`filter-btn ${filter === 'blocked' ? 'active' : ''}`}
                                onClick={() => setFilter('blocked')}
                            >
                                Blocked ({stats?.blocked_anomalies || 0})
                            </button>
                        </div>
                    </div>

                    {loading && (
                        <div className="loading">
                            <div className="loading-spinner"></div>
                            <p>Loading anomalies...</p>
                        </div>
                    )}

                    {error && !loading && (
                        <div className="error">
                            <h3>⚠️ Connection Error</h3>
                            <p>{error}</p>
                            <p className="error-help">
                                Make sure Guardian Log is running on http://localhost:8080
                            </p>
                            <button className="retry-btn" onClick={fetchData}>
                                Retry Connection
                            </button>
                        </div>
                    )}

                    {!loading && !error && anomalies.length === 0 && (
                        <div className="empty-state">
                            <h3>✓ No anomalies found</h3>
                            <p>
                                {filter === 'pending'
                                    ? 'All anomalies have been reviewed!'
                                    : `No ${filter} anomalies to display.`}
                            </p>
                        </div>
                    )}

                    {!loading && !error && anomalies.length > 0 && (
                        <div className="anomalies-list">
                            {anomalies.map((anomaly) => (
                                <AnomalyCard
                                    key={anomaly.id}
                                    anomaly={anomaly}
                                    onApprove={handleApprove}
                                    onBlock={handleBlock}
                                />
                            ))}
                        </div>
                    )}
                </div>
            </main>

            <footer className="footer">
                <p>Guardian Log v1.0 • Self-Hosted DNS Security</p>
            </footer>
        </div>
    );
}

export default App;
