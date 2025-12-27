import React, { useState, useEffect } from 'react';
import { drafts } from '@/services/api';
import { metrics } from '@/types/models';
import './PerformanceMetrics.css';

// No-op stub - reset not implemented in REST API
async function resetDraftPerformanceMetrics(): Promise<void> {
  console.warn('ResetDraftPerformanceMetrics: Not implemented in REST API');
}

interface PerformanceMetricsProps {
    autoRefresh?: boolean; // Auto-refresh metrics every N seconds
    refreshInterval?: number; // Refresh interval in milliseconds
}

const PerformanceMetrics: React.FC<PerformanceMetricsProps> = ({
    autoRefresh = false,
    refreshInterval = 5000,
}) => {
    const [stats, setStats] = useState<metrics.DraftStats | null>(null);
    const [isCollapsed, setIsCollapsed] = useState(true);
    const [loading, setLoading] = useState(false);

    useEffect(() => {
        if (!isCollapsed) {
            loadMetrics();
        }

        if (autoRefresh && !isCollapsed) {
            const interval = setInterval(() => {
                loadMetrics();
            }, refreshInterval);
            return () => clearInterval(interval);
        }
    }, [isCollapsed, autoRefresh, refreshInterval]);

    const loadMetrics = async () => {
        try {
            setLoading(true);
            const data = await drafts.getDraftPerformanceMetrics();
            setStats(data);
        } catch (err) {
            console.error('Error loading performance metrics:', err);
        } finally {
            setLoading(false);
        }
    };

    const handleReset = async () => {
        try {
            await resetDraftPerformanceMetrics();
            await loadMetrics();
        } catch (err) {
            console.error('Error resetting metrics:', err);
        }
    };

    return (
        <div className="performance-metrics">
            <div className="metrics-header" onClick={() => setIsCollapsed(!isCollapsed)}>
                <span className="metrics-title">
                    {isCollapsed ? '▶' : '▼'} Performance Metrics
                </span>
                {!isCollapsed && (
                    <button
                        className="btn-reset-metrics"
                        onClick={(e) => {
                            e.stopPropagation();
                            handleReset();
                        }}
                    >
                        Reset
                    </button>
                )}
            </div>

            {!isCollapsed && (
                <div className="metrics-content">
                    {loading && !stats && (
                        <div className="metrics-loading">Loading metrics...</div>
                    )}

                    {stats && (
                        <>
                            {/* System Info */}
                            <div className="metrics-section">
                                <h4>System Info</h4>
                                <div className="metrics-grid">
                                    <div className="metric-item">
                                        <span className="metric-label">Uptime:</span>
                                        <span className="metric-value">{stats.uptime}</span>
                                    </div>
                                    <div className="metric-item">
                                        <span className="metric-label">Events Processed:</span>
                                        <span className="metric-value">{stats.events_processed.toLocaleString()}</span>
                                    </div>
                                    <div className="metric-item">
                                        <span className="metric-label">Packs Rated:</span>
                                        <span className="metric-value">{stats.packs_rated.toLocaleString()}</span>
                                    </div>
                                </div>
                            </div>

                            {/* Latency Stats */}
                            {stats.end_to_end_latency.count > 0 && (
                                <div className="metrics-section">
                                    <h4>End-to-End Latency</h4>
                                    <div className="latency-stats">
                                        <LatencyRow label="Mean" value={stats.end_to_end_latency.mean} />
                                        <LatencyRow label="P50 (Median)" value={stats.end_to_end_latency.p50} />
                                        <LatencyRow label="P95" value={stats.end_to_end_latency.p95} />
                                        <LatencyRow label="P99" value={stats.end_to_end_latency.p99} />
                                        <LatencyRow label="Min" value={stats.end_to_end_latency.min} />
                                        <LatencyRow label="Max" value={stats.end_to_end_latency.max} />
                                        <div className="latency-row">
                                            <span className="latency-label">Samples:</span>
                                            <span className="latency-value">{stats.end_to_end_latency.count.toLocaleString()}</span>
                                        </div>
                                    </div>
                                </div>
                            )}

                            {/* Parse Latency */}
                            {stats.parse_latency.count > 0 && (
                                <div className="metrics-section">
                                    <h4>Parse Latency</h4>
                                    <div className="latency-stats">
                                        <LatencyRow label="Mean" value={stats.parse_latency.mean} />
                                        <LatencyRow label="P95" value={stats.parse_latency.p95} />
                                        <LatencyRow label="P99" value={stats.parse_latency.p99} />
                                    </div>
                                </div>
                            )}

                            {/* Ratings API Latency */}
                            {stats.ratings_latency.count > 0 && (
                                <div className="metrics-section">
                                    <h4>Ratings API Latency</h4>
                                    <div className="latency-stats">
                                        <LatencyRow label="Mean" value={stats.ratings_latency.mean} />
                                        <LatencyRow label="P95" value={stats.ratings_latency.p95} />
                                        <LatencyRow label="P99" value={stats.ratings_latency.p99} />
                                    </div>
                                </div>
                            )}

                            {/* API Stats */}
                            {stats.api_requests > 0 && (
                                <div className="metrics-section">
                                    <h4>API Statistics</h4>
                                    <div className="metrics-grid">
                                        <div className="metric-item">
                                            <span className="metric-label">API Requests:</span>
                                            <span className="metric-value">{stats.api_requests.toLocaleString()}</span>
                                        </div>
                                        <div className="metric-item">
                                            <span className="metric-label">API Errors:</span>
                                            <span className="metric-value error-value">{stats.api_errors.toLocaleString()}</span>
                                        </div>
                                        <div className="metric-item">
                                            <span className="metric-label">Success Rate:</span>
                                            <span className="metric-value success-value">
                                                {stats.api_success_rate.toFixed(1)}%
                                            </span>
                                        </div>
                                    </div>
                                </div>
                            )}

                            {/* Cache Stats */}
                            {(stats.cache_hits + stats.cache_misses) > 0 && (
                                <div className="metrics-section">
                                    <h4>Cache Statistics</h4>
                                    <div className="metrics-grid">
                                        <div className="metric-item">
                                            <span className="metric-label">Cache Hits:</span>
                                            <span className="metric-value">{stats.cache_hits.toLocaleString()}</span>
                                        </div>
                                        <div className="metric-item">
                                            <span className="metric-label">Cache Misses:</span>
                                            <span className="metric-value">{stats.cache_misses.toLocaleString()}</span>
                                        </div>
                                        <div className="metric-item">
                                            <span className="metric-label">Hit Rate:</span>
                                            <span className="metric-value success-value">
                                                {stats.cache_hit_rate.toFixed(1)}%
                                            </span>
                                        </div>
                                    </div>
                                </div>
                            )}

                            {/* No Data Message */}
                            {stats.events_processed === 0 && (
                                <div className="metrics-empty">
                                    No performance data collected yet. Metrics will appear after draft events are processed.
                                </div>
                            )}
                        </>
                    )}
                </div>
            )}
        </div>
    );
};

// Helper component for latency rows
const LatencyRow: React.FC<{ label: string; value: number }> = ({ label, value }) => {
    const formatLatency = (ms: number): string => {
        if (ms < 1) return `${(ms * 1000).toFixed(0)}µs`;
        if (ms < 1000) return `${ms.toFixed(2)}ms`;
        return `${(ms / 1000).toFixed(2)}s`;
    };

    return (
        <div className="latency-row">
            <span className="latency-label">{label}:</span>
            <span className="latency-value">{formatLatency(value)}</span>
        </div>
    );
};

export default PerformanceMetrics;
