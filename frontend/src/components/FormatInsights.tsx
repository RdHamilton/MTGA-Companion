import React, { useState, useEffect } from 'react';
import { GetFormatInsights } from '../../wailsjs/go/main/App';
import { insights } from '../../wailsjs/go/models';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Cell } from 'recharts';
import './FormatInsights.css';

interface FormatInsightsProps {
    setCode: string;
    draftFormat: string;
    autoRefresh?: boolean;
    refreshInterval?: number;
}

const FormatInsights: React.FC<FormatInsightsProps> = ({
    setCode,
    draftFormat,
    autoRefresh = false,
    refreshInterval = 60000, // Default: 1 minute
}) => {
    const [data, setData] = useState<insights.FormatInsights | null>(null);
    const [isCollapsed, setIsCollapsed] = useState(true);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        if (!isCollapsed && setCode && draftFormat) {
            loadInsights();
        }

        if (autoRefresh && !isCollapsed && setCode && draftFormat) {
            const interval = setInterval(() => {
                loadInsights();
            }, refreshInterval);
            return () => clearInterval(interval);
        }
    }, [isCollapsed, setCode, draftFormat, autoRefresh, refreshInterval]);

    const loadInsights = async () => {
        if (!setCode || !draftFormat) {
            setError('Set code and draft format are required');
            return;
        }

        try {
            setLoading(true);
            setError(null);
            const insights = await GetFormatInsights(setCode, draftFormat);
            setData(insights);
        } catch (err) {
            console.error('Error loading format insights:', err);
            setError(err instanceof Error ? err.message : 'Failed to load insights');
        } finally {
            setLoading(false);
        }
    };

    const getRatingColor = (rating: string): string => {
        switch (rating) {
            case 'S': return '#ffd700'; // Gold
            case 'A': return '#7dff7d'; // Green
            case 'B': return '#4a9eff'; // Blue
            case 'C': return '#ffaa00'; // Orange
            case 'D': return '#ff7d7d'; // Red
            default: return '#aaaaaa'; // Gray
        }
    };

    const formatWinRate = (rate: number): string => {
        return `${rate.toFixed(1)}%`;
    };

    return (
        <div className="format-insights">
            <div className="insights-header" onClick={() => setIsCollapsed(!isCollapsed)}>
                <span className="insights-title">
                    {isCollapsed ? '▶' : '▼'} Format Meta Insights
                    {setCode && draftFormat && ` - ${setCode} ${draftFormat}`}
                </span>
                {!isCollapsed && !loading && (
                    <button
                        className="btn-refresh-insights"
                        onClick={(e) => {
                            e.stopPropagation();
                            loadInsights();
                        }}
                    >
                        Refresh
                    </button>
                )}
            </div>

            {!isCollapsed && (
                <div className="insights-content">
                    {loading && !data && (
                        <div className="insights-loading">Loading format insights...</div>
                    )}

                    {error && (
                        <div className="insights-error">{error}</div>
                    )}

                    {data && (
                        <>
                            {/* Color Power Rankings */}
                            {data.color_rankings && data.color_rankings.length > 0 && (
                                <div className="insights-section">
                                    <h4>Color Power Rankings</h4>
                                    <div className="color-rankings">
                                        <ResponsiveContainer width="100%" height={300}>
                                            <BarChart data={data.color_rankings}>
                                                <CartesianGrid strokeDasharray="3 3" stroke="#3d3d3d" />
                                                <XAxis
                                                    dataKey="color"
                                                    stroke="#aaaaaa"
                                                    style={{ fontSize: '0.9rem' }}
                                                />
                                                <YAxis
                                                    stroke="#aaaaaa"
                                                    style={{ fontSize: '0.9rem' }}
                                                    domain={[40, 65]}
                                                    label={{ value: 'Win Rate (%)', angle: -90, position: 'insideLeft', fill: '#aaaaaa' }}
                                                />
                                                <Tooltip
                                                    contentStyle={{ backgroundColor: '#2d2d2d', border: '1px solid #4a9eff' }}
                                                    labelStyle={{ color: '#ffffff' }}
                                                    formatter={(value: number) => [`${value.toFixed(1)}%`, 'Win Rate']}
                                                />
                                                <Bar dataKey="win_rate" radius={[8, 8, 0, 0]}>
                                                    {data.color_rankings.map((entry, index) => (
                                                        <Cell key={`cell-${index}`} fill={getRatingColor(entry.rating)} />
                                                    ))}
                                                </Bar>
                                            </BarChart>
                                        </ResponsiveContainer>
                                        <div className="color-rankings-grid">
                                            {data.color_rankings.map((rank, idx) => (
                                                <div key={idx} className="color-rank-item">
                                                    <div className="rank-header">
                                                        <span className="rank-color">{rank.color}</span>
                                                        <span
                                                            className="rank-rating"
                                                            style={{ color: getRatingColor(rank.rating) }}
                                                        >
                                                            {rank.rating}
                                                        </span>
                                                    </div>
                                                    <div className="rank-stats">
                                                        <div className="rank-stat">
                                                            <span className="stat-label">Win Rate:</span>
                                                            <span className="stat-value">{formatWinRate(rank.win_rate)}</span>
                                                        </div>
                                                        <div className="rank-stat">
                                                            <span className="stat-label">Popularity:</span>
                                                            <span className="stat-value">{formatWinRate(rank.popularity)}</span>
                                                        </div>
                                                        <div className="rank-stat">
                                                            <span className="stat-label">Games:</span>
                                                            <span className="stat-value">{rank.games_played.toLocaleString()}</span>
                                                        </div>
                                                    </div>
                                                </div>
                                            ))}
                                        </div>
                                    </div>
                                </div>
                            )}

                            {/* Format Speed */}
                            {data.format_speed && (
                                <div className="insights-section">
                                    <h4>Format Speed</h4>
                                    <div className="format-speed">
                                        <div className="speed-badge">{data.format_speed.speed}</div>
                                        <div className="speed-description">{data.format_speed.description}</div>
                                    </div>
                                </div>
                            )}

                            {/* Color Analysis */}
                            {data.color_analysis && (
                                <div className="insights-section">
                                    <h4>Color Analysis</h4>
                                    <div className="color-analysis-grid">
                                        {data.color_analysis.best_mono_color && (
                                            <div className="analysis-item">
                                                <span className="analysis-label">Best Mono Color:</span>
                                                <span className="analysis-value">{data.color_analysis.best_mono_color}</span>
                                            </div>
                                        )}
                                        {data.color_analysis.best_color_pair && (
                                            <div className="analysis-item">
                                                <span className="analysis-label">Best Color Pair:</span>
                                                <span className="analysis-value">{data.color_analysis.best_color_pair}</span>
                                            </div>
                                        )}
                                    </div>
                                    {data.color_analysis.overdrafted_colors && data.color_analysis.overdrafted_colors.length > 0 && (
                                        <div className="overdrafted-section">
                                            <h5>Overdrafted Colors (Popularity &gt; Win Rate)</h5>
                                            <div className="overdrafted-grid">
                                                {data.color_analysis.overdrafted_colors.map((od, idx) => (
                                                    <div key={idx} className="overdrafted-item">
                                                        <span className="od-color">{od.color}</span>
                                                        <span className="od-stats">
                                                            WR: {formatWinRate(od.win_rate)} |
                                                            Pop: {formatWinRate(od.popularity)} |
                                                            Δ: +{od.delta.toFixed(1)}%
                                                        </span>
                                                    </div>
                                                ))}
                                            </div>
                                        </div>
                                    )}
                                </div>
                            )}

                            {/* Top Cards Sections */}
                            <div className="top-cards-container">
                                {/* Top Bombs */}
                                {data.top_bombs && data.top_bombs.length > 0 && (
                                    <div className="insights-section top-cards-section">
                                        <h4>Top Bombs (Rare/Mythic)</h4>
                                        <TopCardsList cards={data.top_bombs} />
                                    </div>
                                )}

                                {/* Top Removal */}
                                {data.top_removal && data.top_removal.length > 0 && (
                                    <div className="insights-section top-cards-section">
                                        <h4>Top Removal</h4>
                                        <TopCardsList cards={data.top_removal} />
                                    </div>
                                )}

                                {/* Top Creatures */}
                                {data.top_creatures && data.top_creatures.length > 0 && (
                                    <div className="insights-section top-cards-section">
                                        <h4>Top Performers</h4>
                                        <TopCardsList cards={data.top_creatures} limit={10} />
                                    </div>
                                )}

                                {/* Top Commons */}
                                {data.top_commons && data.top_commons.length > 0 && (
                                    <div className="insights-section top-cards-section">
                                        <h4>Best Commons</h4>
                                        <TopCardsList cards={data.top_commons} limit={15} />
                                    </div>
                                )}
                            </div>
                        </>
                    )}

                    {!loading && !error && !data && (
                        <div className="insights-empty">
                            No format insights available. Make sure card ratings are loaded for this format.
                        </div>
                    )}
                </div>
            )}
        </div>
    );
};

// Helper component for displaying top cards lists
const TopCardsList: React.FC<{ cards: insights.TopCard[], limit?: number }> = ({ cards, limit }) => {
    const displayCards = limit ? cards.slice(0, limit) : cards;

    return (
        <div className="top-cards-list">
            {displayCards.map((card, idx) => (
                <div key={idx} className="top-card-item">
                    <div className="card-rank">#{idx + 1}</div>
                    <div className="card-info">
                        <div className="card-name">{card.name}</div>
                        <div className="card-meta">
                            <span className="card-rarity">{card.rarity}</span>
                            {card.color && <span className="card-color">{card.color}</span>}
                        </div>
                    </div>
                    <div className="card-gihwr">
                        <div className="gihwr-value">{card.gihwr.toFixed(1)}%</div>
                        <div className="gihwr-label">GIHWR</div>
                    </div>
                </div>
            ))}
        </div>
    );
};

export default FormatInsights;
