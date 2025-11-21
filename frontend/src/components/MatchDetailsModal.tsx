import { useState, useEffect } from 'react';
import { GetMatchGames } from '../../wailsjs/go/main/App';
import { models } from '../../wailsjs/go/models';
import LoadingSpinner from './LoadingSpinner';
import './MatchDetailsModal.css';

interface MatchDetailsModalProps {
  match: models.Match;
  onClose: () => void;
}

const MatchDetailsModal = ({ match, onClose }: MatchDetailsModalProps) => {
  const [games, setGames] = useState<models.Game[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadGames();
  }, [match.ID]);

  const loadGames = async () => {
    try {
      setLoading(true);
      setError(null);
      const gamesData = await GetMatchGames(match.ID);
      setGames(gamesData || []);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load games');
      console.error('Error loading games:', err);
    } finally {
      setLoading(false);
    }
  };

  const formatTimestamp = (timestamp: any) => {
    return new Date(timestamp).toLocaleString();
  };

  const formatDuration = (seconds: number | null | undefined) => {
    if (!seconds) return 'N/A';
    const minutes = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${minutes}:${secs.toString().padStart(2, '0')}`;
  };

  const formatResultReason = (reason: string | null | undefined) => {
    if (!reason) return 'N/A';
    switch (reason.toLowerCase()) {
      case 'concede':
        return 'Conceded';
      case 'timeout':
        return 'Timeout';
      case 'normal':
        return 'Win Condition';
      default:
        return reason;
    }
  };

  const getRankChange = () => {
    if (!match.RankBefore || !match.RankAfter) return null;
    if (match.RankBefore === match.RankAfter) return 'No Change';
    return `${match.RankBefore} → ${match.RankAfter}`;
  };

  // Handle click on backdrop to close modal
  const handleBackdropClick = (e: React.MouseEvent<HTMLDivElement>) => {
    if (e.target === e.currentTarget) {
      onClose();
    }
  };

  // Handle Escape key to close modal
  useEffect(() => {
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        onClose();
      }
    };
    document.addEventListener('keydown', handleEscape);
    return () => document.removeEventListener('keydown', handleEscape);
  }, [onClose]);

  return (
    <div className="modal-backdrop" onClick={handleBackdropClick}>
      <div className="modal-content match-details-modal">
        <div className="modal-header">
          <h2>Match Details</h2>
          <button className="modal-close" onClick={onClose}>
            ×
          </button>
        </div>

        <div className="modal-body">
          {/* Match Summary */}
          <div className="match-summary">
            <div className="summary-row">
              <span className="summary-label">Format:</span>
              <span className="summary-value">{match.Format}</span>
            </div>
            <div className="summary-row">
              <span className="summary-label">Event:</span>
              <span className="summary-value">{match.EventName}</span>
            </div>
            <div className="summary-row">
              <span className="summary-label">Result:</span>
              <span className={`summary-value result-badge ${match.Result.toLowerCase()}`}>
                {match.Result.toUpperCase()} {match.PlayerWins}-{match.OpponentWins}
              </span>
            </div>
            <div className="summary-row">
              <span className="summary-label">Date:</span>
              <span className="summary-value">{formatTimestamp(match.Timestamp)}</span>
            </div>
            {match.OpponentName && (
              <div className="summary-row">
                <span className="summary-label">Opponent:</span>
                <span className="summary-value">{match.OpponentName}</span>
              </div>
            )}
            {getRankChange() && (
              <div className="summary-row">
                <span className="summary-label">Rank:</span>
                <span className="summary-value">{getRankChange()}</span>
              </div>
            )}
          </div>

          {/* Games Breakdown */}
          <div className="games-section">
            <h3>Game Breakdown</h3>

            {loading && <LoadingSpinner message="Loading games..." />}

            {error && (
              <div className="error-message">
                Failed to load games: {error}
              </div>
            )}

            {!loading && !error && games.length === 0 && (
              <div className="no-games">
                No game data available for this match.
              </div>
            )}

            {!loading && !error && games.length > 0 && (
              <div className="games-list">
                {games.map((game) => (
                  <div key={game.ID} className={`game-item ${game.Result.toLowerCase()}`}>
                    <div className="game-header">
                      <span className="game-number">Game {game.GameNumber}</span>
                      <span className={`game-result ${game.Result.toLowerCase()}`}>
                        {game.Result === 'win' ? 'WIN' : 'LOSS'}
                      </span>
                    </div>
                    <div className="game-details">
                      <div className="game-detail">
                        <span className="detail-label">Duration:</span>
                        <span className="detail-value">{formatDuration(game.DurationSeconds)}</span>
                      </div>
                      <div className="game-detail">
                        <span className="detail-label">Result:</span>
                        <span className="detail-value">{formatResultReason(game.ResultReason)}</span>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>

        <div className="modal-footer">
          <button className="btn-secondary" onClick={onClose}>
            Close
          </button>
        </div>
      </div>
    </div>
  );
};

export default MatchDetailsModal;
