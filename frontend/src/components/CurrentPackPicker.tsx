import React, { useState, useEffect } from 'react';
import { drafts } from '@/services/api';
import { gui } from '@/types/models';
import './CurrentPackPicker.css';

interface CurrentPackPickerProps {
    sessionID: string;
    onRefresh?: () => void;
}

const CARD_BACK_URL = 'https://backs.scryfall.io/large/59/5/597b79b3-7d77-4261-871a-96d8dba8c93f.jpg?1562636924';

const CurrentPackPicker: React.FC<CurrentPackPickerProps> = ({ sessionID, onRefresh }) => {
    const [packData, setPackData] = useState<gui.CurrentPackResponse | null>(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    const loadPackData = async () => {
        try {
            setLoading(true);
            setError(null);
            const data = await drafts.getCurrentPackWithRecommendation(sessionID);
            setPackData(data);
        } catch (err) {
            setError(err instanceof Error ? err.message : 'Failed to load pack data');
            console.error('Error loading pack data:', err);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        if (sessionID) {
            loadPackData();
        }
    }, [sessionID]);

    const handleRefresh = () => {
        loadPackData();
        if (onRefresh) {
            onRefresh();
        }
    };

    const getTierColor = (tier: string): string => {
        switch (tier) {
            case 'S': return '#ffd700'; // Gold
            case 'A': return '#c0c0c0'; // Silver
            case 'B': return '#cd7f32'; // Bronze
            case 'C': return '#4a9eff'; // Blue
            case 'D': return '#888888'; // Gray
            case 'F': return '#ff4444'; // Red
            default: return '#aaaaaa';
        }
    };

    const getColorSymbol = (color: string): string => {
        switch (color) {
            case 'W': return 'W';
            case 'U': return 'U';
            case 'B': return 'B';
            case 'R': return 'R';
            case 'G': return 'G';
            default: return 'C'; // Colorless
        }
    };

    const renderColorIndicator = (colors: string[] | undefined) => {
        if (!colors || colors.length === 0) {
            return <span className="color-indicator colorless">C</span>;
        }
        return (
            <div className="color-indicators">
                {colors.map((color, idx) => (
                    <span key={idx} className={`color-indicator color-${color.toLowerCase()}`}>
                        {getColorSymbol(color)}
                    </span>
                ))}
            </div>
        );
    };

    if (loading) {
        return (
            <div className="current-pack-loading">
                <div className="loading-spinner"></div>
                <p>Loading current pack...</p>
            </div>
        );
    }

    if (error) {
        return (
            <div className="current-pack-error">
                <p>{error}</p>
                <button onClick={handleRefresh} className="retry-btn">Retry</button>
            </div>
        );
    }

    if (!packData || !packData.cards || packData.cards.length === 0) {
        return (
            <div className="current-pack-empty">
                <p>No pack data available</p>
                <p className="help-text">Pack data will appear when you start a draft pick</p>
            </div>
        );
    }

    return (
        <div className="current-pack-container">
            <div className="current-pack-header">
                <h2>{packData.pack_label}</h2>
                <div className="pack-info">
                    <span className="pool-info">Pool: {packData.pool_size} cards</span>
                    {packData.pool_colors && packData.pool_colors.length > 0 && (
                        <span className="pool-colors">
                            Colors: {renderColorIndicator(packData.pool_colors)}
                        </span>
                    )}
                    <button onClick={handleRefresh} className="refresh-btn" title="Refresh pack data">
                        Refresh
                    </button>
                </div>
            </div>

            {/* Recommended Pick Banner */}
            {packData.recommended_card && (
                <div className="recommended-banner">
                    <span className="rec-label">Recommended Pick:</span>
                    <span className="rec-card-name">{packData.recommended_card.name}</span>
                    <span className="rec-tier" style={{ color: getTierColor(packData.recommended_card.tier) }}>
                        {packData.recommended_card.tier}
                    </span>
                    {packData.recommended_card.reasoning && (
                        <span className="rec-reason">{packData.recommended_card.reasoning}</span>
                    )}
                </div>
            )}

            {/* Pack Cards Grid */}
            <div className="pack-cards-grid">
                {packData.cards.map((card, index) => (
                    <div
                        key={card.arena_id || index}
                        className={`pack-card ${card.is_recommended ? 'recommended' : ''}`}
                    >
                        <div className="card-image-container">
                            <img
                                src={card.image_url || CARD_BACK_URL}
                                alt={card.name}
                                className="card-image"
                                loading="lazy"
                                onError={(e) => {
                                    (e.target as HTMLImageElement).src = CARD_BACK_URL;
                                }}
                            />
                            <div className="tier-badge" style={{ backgroundColor: getTierColor(card.tier) }}>
                                {card.tier}
                            </div>
                            {card.is_recommended && (
                                <div className="recommended-indicator">Best Pick</div>
                            )}
                        </div>
                        <div className="card-info">
                            <div className="card-name">{card.name}</div>
                            <div className="card-stats">
                                {renderColorIndicator(card.colors)}
                                <span className="gihwr" title="Games In Hand Win Rate">
                                    {card.gihwr?.toFixed(1)}%
                                </span>
                                <span className="alsa" title="Average Last Seen At">
                                    ALSA: {card.alsa?.toFixed(1)}
                                </span>
                            </div>
                            {card.reasoning && (
                                <div className="card-reasoning">{card.reasoning}</div>
                            )}
                        </div>
                    </div>
                ))}
            </div>
        </div>
    );
};

export default CurrentPackPicker;
