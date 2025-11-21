import React, { useState, useEffect } from 'react';
import { GetDraftDeckMetrics } from '../../wailsjs/go/main/App';
import { models } from '../../wailsjs/go/models';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Cell, PieChart, Pie, Legend } from 'recharts';
import './DraftStatistics.css';

interface DraftStatisticsProps {
    sessionID: string;
    pickCount: number; // Used to trigger refresh when picks change
}

const COLOR_MAP: { [key: string]: string } = {
    'W': '#F0E68C',  // White/Yellow
    'U': '#0E68AB',  // Blue
    'B': '#150B00',  // Black
    'R': '#D3202A',  // Red
    'G': '#00733E',  // Green
};

const COLOR_NAMES: { [key: string]: string } = {
    'W': 'White',
    'U': 'Blue',
    'B': 'Black',
    'R': 'Red',
    'G': 'Green',
};

const DraftStatistics: React.FC<DraftStatisticsProps> = ({ sessionID, pickCount }) => {
    const [metrics, setMetrics] = useState<models.DeckMetrics | null>(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        loadMetrics();
    }, [sessionID, pickCount]); // Reload when session or pick count changes

    const loadMetrics = async () => {
        try {
            setLoading(true);
            setError(null);
            const data = await GetDraftDeckMetrics(sessionID);
            setMetrics(data);
        } catch (err) {
            setError(err instanceof Error ? err.message : 'Failed to load metrics');
            console.error('Error loading draft metrics:', err);
        } finally {
            setLoading(false);
        }
    };

    if (loading) {
        return (
            <div className="draft-statistics">
                <div className="statistics-loading">Loading statistics...</div>
            </div>
        );
    }

    if (error) {
        return (
            <div className="draft-statistics">
                <div className="statistics-error">Error: {error}</div>
            </div>
        );
    }

    if (!metrics || metrics.total_cards === 0) {
        return (
            <div className="draft-statistics">
                <div className="statistics-empty">No cards drafted yet</div>
            </div>
        );
    }

    // Prepare mana curve data
    const cmcLabels = ['0', '1', '2', '3', '4', '5', '6+'];
    const manaCurveData = cmcLabels.map((label, index) => ({
        cmc: label,
        all: metrics.distribution_all?.[index] || 0,
        creatures: metrics.distribution_creatures?.[index] || 0,
        spells: metrics.distribution_noncreatures?.[index] || 0,
    }));

    // Prepare color distribution data
    const colorData = Object.keys(COLOR_MAP)
        .filter(color => (metrics.color_counts?.[color] || 0) > 0)
        .map(color => ({
            name: COLOR_NAMES[color],
            value: metrics.color_counts?.[color] || 0,
            color: COLOR_MAP[color],
        }));

    // Prepare type breakdown data (top 5 types)
    const typeData = Object.entries(metrics.type_breakdown || {})
        .sort(([, a], [, b]) => b - a)
        .slice(0, 5)
        .map(([type, count]) => ({
            type,
            count,
        }));

    // Calculate percentages
    const creaturePercent = metrics.total_cards > 0
        ? ((metrics.creature_count / metrics.total_cards) * 100).toFixed(1)
        : '0.0';
    const spellPercent = metrics.total_cards > 0
        ? ((metrics.noncreature_count / metrics.total_cards) * 100).toFixed(1)
        : '0.0';

    return (
        <div className="draft-statistics">
            <h2>Draft Statistics</h2>

            {/* Summary Stats */}
            <div className="stats-summary">
                <div className="stat-box">
                    <span className="stat-label">Total Cards</span>
                    <span className="stat-value">{metrics.total_cards}</span>
                </div>
                <div className="stat-box">
                    <span className="stat-label">Avg CMC</span>
                    <span className="stat-value">{metrics.cmc_average.toFixed(2)}</span>
                </div>
                <div className="stat-box">
                    <span className="stat-label">Creatures</span>
                    <span className="stat-value">{metrics.creature_count} ({creaturePercent}%)</span>
                </div>
                <div className="stat-box">
                    <span className="stat-label">Spells</span>
                    <span className="stat-value">{metrics.noncreature_count} ({spellPercent}%)</span>
                </div>
            </div>

            {/* Mana Curve */}
            <div className="stats-section">
                <h3>Mana Curve</h3>
                <ResponsiveContainer width="100%" height={200}>
                    <BarChart data={manaCurveData} margin={{ top: 5, right: 5, left: 5, bottom: 5 }}>
                        <CartesianGrid strokeDasharray="3 3" stroke="#444" />
                        <XAxis dataKey="cmc" stroke="#aaa" />
                        <YAxis stroke="#aaa" />
                        <Tooltip
                            contentStyle={{ backgroundColor: '#2d2d2d', border: '1px solid #444' }}
                            labelStyle={{ color: '#fff' }}
                        />
                        <Bar dataKey="all" fill="#4a9eff" name="All Cards" />
                    </BarChart>
                </ResponsiveContainer>
            </div>

            {/* Creature vs Spell Curve */}
            <div className="stats-section">
                <h3>Creatures vs Spells by CMC</h3>
                <ResponsiveContainer width="100%" height={200}>
                    <BarChart data={manaCurveData} margin={{ top: 5, right: 5, left: 5, bottom: 5 }}>
                        <CartesianGrid strokeDasharray="3 3" stroke="#444" />
                        <XAxis dataKey="cmc" stroke="#aaa" />
                        <YAxis stroke="#aaa" />
                        <Tooltip
                            contentStyle={{ backgroundColor: '#2d2d2d', border: '1px solid #444' }}
                            labelStyle={{ color: '#fff' }}
                        />
                        <Legend />
                        <Bar dataKey="creatures" fill="#7dff7d" name="Creatures" />
                        <Bar dataKey="spells" fill="#ff7d7d" name="Spells" />
                    </BarChart>
                </ResponsiveContainer>
            </div>

            {/* Color Distribution */}
            {colorData.length > 0 && (
                <div className="stats-section">
                    <h3>Color Distribution</h3>
                    <ResponsiveContainer width="100%" height={200}>
                        <PieChart>
                            <Pie
                                data={colorData}
                                cx="50%"
                                cy="50%"
                                labelLine={false}
                                label={(entry) => `${entry.name}: ${entry.value}`}
                                outerRadius={80}
                                fill="#8884d8"
                                dataKey="value"
                            >
                                {colorData.map((entry, index) => (
                                    <Cell key={`cell-${index}`} fill={entry.color} />
                                ))}
                            </Pie>
                            <Tooltip
                                contentStyle={{ backgroundColor: '#2d2d2d', border: '1px solid #444' }}
                            />
                        </PieChart>
                    </ResponsiveContainer>
                    {metrics.multi_color_count > 0 && (
                        <div className="color-notes">
                            <span>Multi-color cards: {metrics.multi_color_count}</span>
                        </div>
                    )}
                    {metrics.colorless_count > 0 && (
                        <div className="color-notes">
                            <span>Colorless cards: {metrics.colorless_count}</span>
                        </div>
                    )}
                </div>
            )}

            {/* Type Breakdown */}
            {typeData.length > 0 && (
                <div className="stats-section">
                    <h3>Card Types</h3>
                    <div className="type-breakdown">
                        {typeData.map(({ type, count }) => (
                            <div key={type} className="type-row">
                                <span className="type-name">{type}</span>
                                <span className="type-count">{count}</span>
                            </div>
                        ))}
                    </div>
                </div>
            )}
        </div>
    );
};

export default DraftStatistics;
