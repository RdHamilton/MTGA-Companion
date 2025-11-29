import { useState, useEffect } from 'react';
import {
  GetMetaDashboard,
  RefreshMetaData,
  GetSupportedFormats,
} from '../../wailsjs/go/main/App';
import { gui } from '../../wailsjs/go/models';
import './Meta.css';

export default function Meta() {
  const [format, setFormat] = useState<string>('standard');
  const [supportedFormats, setSupportedFormats] = useState<string[]>([]);
  const [dashboardData, setDashboardData] = useState<gui.MetaDashboardResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Load supported formats on mount
  useEffect(() => {
    const loadFormats = async () => {
      try {
        const formats = await GetSupportedFormats();
        setSupportedFormats(formats);
      } catch (err) {
        console.error('Failed to load formats:', err);
        setSupportedFormats(['standard', 'historic', 'explorer', 'pioneer', 'modern']);
      }
    };
    loadFormats();
  }, []);

  // Load dashboard data when format changes
  useEffect(() => {
    const loadDashboard = async () => {
      setLoading(true);
      setError(null);
      try {
        const data = await GetMetaDashboard(format);
        if (data.error) {
          setError(data.error);
        } else {
          setDashboardData(data);
        }
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load meta data');
      } finally {
        setLoading(false);
      }
    };
    loadDashboard();
  }, [format]);

  const handleRefresh = async () => {
    setRefreshing(true);
    setError(null);
    try {
      const data = await RefreshMetaData(format);
      if (data.error) {
        setError(data.error);
      } else {
        setDashboardData(data);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to refresh meta data');
    } finally {
      setRefreshing(false);
    }
  };

  const formatDate = (dateValue: unknown) => {
    try {
      // Handle time.Time class (has no useful properties, but the raw source is a string)
      // or string values from JSON serialization
      const dateStr = typeof dateValue === 'string'
        ? dateValue
        : (dateValue as { toString?: () => string })?.toString?.() || '';
      if (!dateStr) return 'Unknown';
      const date = new Date(dateStr);
      if (isNaN(date.getTime())) return 'Unknown';
      return date.toLocaleString();
    } catch {
      return 'Unknown';
    }
  };

  const getColorBadge = (colors: string[]) => {
    if (!colors || colors.length === 0) return null;
    return (
      <span className="color-badge">
        {colors.map((c) => (
          <span key={c} className={`color-pip color-${c.toLowerCase()}`} title={c}>
            {c}
          </span>
        ))}
      </span>
    );
  };

  const getTierLabel = (tier: number) => {
    switch (tier) {
      case 1:
        return <span className="tier-badge tier-1">Tier 1</span>;
      case 2:
        return <span className="tier-badge tier-2">Tier 2</span>;
      case 3:
        return <span className="tier-badge tier-3">Tier 3</span>;
      default:
        return <span className="tier-badge tier-4">Tier 4</span>;
    }
  };

  const getTrendIcon = (trend: string) => {
    switch (trend) {
      case 'up':
        return <span className="trend-icon trend-up" title="Trending up">↗</span>;
      case 'down':
        return <span className="trend-icon trend-down" title="Trending down">↘</span>;
      default:
        return <span className="trend-icon trend-stable" title="Stable">→</span>;
    }
  };

  const groupArchetypesByTier = () => {
    if (!dashboardData?.archetypes) return {};
    const grouped: Record<number, gui.ArchetypeInfo[]> = {};
    for (const arch of dashboardData.archetypes) {
      const tier = arch.tier || 4;
      if (!grouped[tier]) grouped[tier] = [];
      grouped[tier].push(arch);
    }
    return grouped;
  };

  return (
    <div className="meta-page">
      {/* Header */}
      <div className="meta-header">
        <div className="meta-title">
          <h1>Metagame Dashboard</h1>
          <p className="meta-description">
            Current metagame data from MTGGoldfish and MTGTop8
          </p>
        </div>
        <div className="meta-controls">
          <select
            className="format-select"
            value={format}
            onChange={(e) => setFormat(e.target.value)}
            disabled={loading || refreshing}
          >
            {supportedFormats.map((f) => (
              <option key={f} value={f}>
                {f.charAt(0).toUpperCase() + f.slice(1)}
              </option>
            ))}
          </select>
          <button
            className="refresh-button"
            onClick={handleRefresh}
            disabled={loading || refreshing}
          >
            {refreshing ? '⟳ Refreshing...' : '⟳ Refresh'}
          </button>
        </div>
      </div>

      {/* Error */}
      {error && (
        <div className="meta-error">
          <strong>Error:</strong> {error}
        </div>
      )}

      {/* Loading */}
      {loading && (
        <div className="meta-loading">
          <div className="loading-spinner" />
          <span>Loading meta data for {format}...</span>
        </div>
      )}

      {/* Dashboard content */}
      {!loading && dashboardData && (
        <div className="meta-content">
          {/* Summary stats */}
          <div className="meta-summary">
            <div className="summary-stat">
              <span className="stat-value">{dashboardData.totalArchetypes}</span>
              <span className="stat-label">Archetypes</span>
            </div>
            <div className="summary-stat">
              <span className="stat-value">{dashboardData.tournaments?.length || 0}</span>
              <span className="stat-label">Recent Tournaments</span>
            </div>
            <div className="summary-stat">
              <span className="stat-value">{dashboardData.sources?.join(', ') || 'N/A'}</span>
              <span className="stat-label">Data Sources</span>
            </div>
            <div className="summary-stat">
              <span className="stat-value">{formatDate(dashboardData.lastUpdated)}</span>
              <span className="stat-label">Last Updated</span>
            </div>
          </div>

          {/* Tier lists */}
          <div className="tier-lists">
            {[1, 2, 3, 4].map((tier) => {
              const archetypes = groupArchetypesByTier()[tier] || [];
              if (archetypes.length === 0 && tier < 4) return null;

              return (
                <div key={tier} className={`tier-section tier-${tier}-section`}>
                  <h2 className="tier-header">
                    {getTierLabel(tier)}
                    <span className="tier-count">({archetypes.length} decks)</span>
                  </h2>
                  <div className="archetype-list">
                    {archetypes.length === 0 ? (
                      <div className="no-archetypes">No archetypes in this tier</div>
                    ) : (
                      archetypes.map((arch, idx) => (
                        <div key={`${arch.name}-${idx}`} className="archetype-card">
                          <div className="archetype-header">
                            <span className="archetype-name">{arch.name}</span>
                            {getColorBadge(arch.colors)}
                            {getTrendIcon(arch.trendDirection)}
                          </div>
                          <div className="archetype-stats">
                            {arch.metaShare > 0 && (
                              <div className="stat-item">
                                <span className="stat-icon">📊</span>
                                <span className="stat-text">{(arch.metaShare).toFixed(1)}% meta share</span>
                              </div>
                            )}
                            {arch.tournamentTop8s > 0 && (
                              <div className="stat-item">
                                <span className="stat-icon">🏆</span>
                                <span className="stat-text">{arch.tournamentTop8s} Top 8s</span>
                              </div>
                            )}
                            {arch.tournamentWins > 0 && (
                              <div className="stat-item">
                                <span className="stat-icon">🥇</span>
                                <span className="stat-text">{arch.tournamentWins} Wins</span>
                              </div>
                            )}
                            {arch.confidenceScore > 0 && (
                              <div className="stat-item confidence">
                                <span className="stat-text">{Math.round(arch.confidenceScore * 100)}% confidence</span>
                              </div>
                            )}
                          </div>
                        </div>
                      ))
                    )}
                  </div>
                </div>
              );
            })}
          </div>

          {/* Recent Tournaments */}
          {dashboardData.tournaments && dashboardData.tournaments.length > 0 && (
            <div className="tournaments-section">
              <h2>Recent Tournaments</h2>
              <div className="tournament-list">
                {dashboardData.tournaments.slice(0, 10).map((tournament, idx) => (
                  <div key={`${tournament.name}-${idx}`} className="tournament-card">
                    <div className="tournament-name">{tournament.name}</div>
                    <div className="tournament-meta">
                      <span>{tournament.format}</span>
                      {tournament.players > 0 && <span>{tournament.players} players</span>}
                      <span>{formatDate(tournament.date)}</span>
                    </div>
                    {tournament.topDecks && tournament.topDecks.length > 0 && (
                      <div className="tournament-decks">
                        <strong>Top Decks:</strong> {tournament.topDecks.slice(0, 3).join(', ')}
                        {tournament.topDecks.length > 3 && ` +${tournament.topDecks.length - 3} more`}
                      </div>
                    )}
                    {tournament.sourceUrl && (
                      <a
                        href={tournament.sourceUrl}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="tournament-link"
                      >
                        View Details →
                      </a>
                    )}
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* No data message */}
          {(!dashboardData.archetypes || dashboardData.archetypes.length === 0) &&
           (!dashboardData.tournaments || dashboardData.tournaments.length === 0) && (
            <div className="no-data">
              <div className="no-data-icon">📭</div>
              <h3>No Meta Data Available</h3>
              <p>
                Unable to fetch metagame data for {format}. This could be because:
              </p>
              <ul>
                <li>The format is not supported by our data sources</li>
                <li>There was a network error fetching the data</li>
                <li>The data sources are temporarily unavailable</li>
              </ul>
              <button onClick={handleRefresh} className="retry-button">
                Try Again
              </button>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
