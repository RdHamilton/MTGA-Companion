import { useState, useEffect, useRef } from 'react';
import * as gameplays from '@/services/api/gameplays';
import type { PlayTimelineEntry, GamePlay, GameStateSnapshot } from '@/services/api/gameplays';
import LoadingSpinner from './LoadingSpinner';
import './GamePlayTimelinePanel.css';

interface GamePlayTimelinePanelProps {
  matchId: string;
  isExpanded?: boolean;
  onToggle?: () => void;
}

const GamePlayTimelinePanel = ({
  matchId,
  isExpanded = false,
  onToggle,
}: GamePlayTimelinePanelProps) => {
  const [timeline, setTimeline] = useState<PlayTimelineEntry[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [selectedTurn, setSelectedTurn] = useState<number | null>(null);
  const isMountedRef = useRef(true);

  useEffect(() => {
    isMountedRef.current = true;
    return () => {
      isMountedRef.current = false;
    };
  }, []);

  useEffect(() => {
    if (isExpanded && timeline.length === 0 && !loading) {
      loadTimeline();
    }
  }, [isExpanded, matchId]);

  const loadTimeline = async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await gameplays.getMatchTimeline(matchId);
      if (!isMountedRef.current) return;
      setTimeline(data || []);
      // Auto-select first turn if available
      if (data && data.length > 0) {
        setSelectedTurn(data[0].turn);
      }
    } catch (err) {
      if (!isMountedRef.current) return;
      setError(err instanceof Error ? err.message : 'Failed to load game timeline');
      console.error('Error loading game timeline:', err);
    } finally {
      if (isMountedRef.current) {
        setLoading(false);
      }
    }
  };

  const formatActionType = (actionType: string): string => {
    switch (actionType) {
      case 'play_card':
        return 'Played';
      case 'land_drop':
        return 'Land';
      case 'attack':
        return 'Attack';
      case 'block':
        return 'Block';
      case 'mulligan':
        return 'Mulligan';
      default:
        return actionType;
    }
  };

  const getActionIcon = (actionType: string): string => {
    switch (actionType) {
      case 'play_card':
        return '🃏';
      case 'land_drop':
        return '🏔️';
      case 'attack':
        return '⚔️';
      case 'block':
        return '🛡️';
      case 'mulligan':
        return '🔄';
      default:
        return '•';
    }
  };

  const renderPlay = (play: GamePlay) => {
    return (
      <div key={play.id} className={`play-item ${play.player_type}`}>
        <span className="play-icon">{getActionIcon(play.action_type)}</span>
        <span className="play-action">{formatActionType(play.action_type)}</span>
        {play.card_name && <span className="play-card">{play.card_name}</span>}
        {play.zone_from && play.zone_to && (
          <span className="play-zones">
            {play.zone_from} → {play.zone_to}
          </span>
        )}
      </div>
    );
  };

  const renderSnapshot = (snapshot: GameStateSnapshot | undefined) => {
    if (!snapshot) return null;

    return (
      <div className="turn-snapshot">
        <div className="snapshot-row">
          <span className="snapshot-label">Life:</span>
          <span className="snapshot-value life-values">
            <span className="player-life">You: {snapshot.player_life ?? '?'}</span>
            <span className="separator">|</span>
            <span className="opponent-life">Opp: {snapshot.opponent_life ?? '?'}</span>
          </span>
        </div>
        <div className="snapshot-row">
          <span className="snapshot-label">Cards:</span>
          <span className="snapshot-value">
            <span className="player-cards">You: {snapshot.player_cards_in_hand ?? '?'}</span>
            <span className="separator">|</span>
            <span className="opponent-cards">Opp: {snapshot.opponent_cards_in_hand ?? '?'}</span>
          </span>
        </div>
        <div className="snapshot-row">
          <span className="snapshot-label">Lands:</span>
          <span className="snapshot-value">
            <span className="player-lands">You: {snapshot.player_lands_in_play ?? '?'}</span>
            <span className="separator">|</span>
            <span className="opponent-lands">Opp: {snapshot.opponent_lands_in_play ?? '?'}</span>
          </span>
        </div>
      </div>
    );
  };

  const renderTurnDetails = (entry: PlayTimelineEntry) => {
    const hasPlayerPlays = entry.player_plays && entry.player_plays.length > 0;
    const hasOpponentPlays = entry.opponent_plays && entry.opponent_plays.length > 0;

    return (
      <div className="turn-details">
        {renderSnapshot(entry.snapshot)}

        <div className="plays-columns">
          <div className="plays-column player-column">
            <h5>Your Plays</h5>
            {hasPlayerPlays ? (
              <div className="plays-list">
                {entry.player_plays.map((play) => renderPlay(play))}
              </div>
            ) : (
              <div className="no-plays">No plays</div>
            )}
          </div>

          <div className="plays-column opponent-column">
            <h5>Opponent Plays</h5>
            {hasOpponentPlays ? (
              <div className="plays-list">
                {entry.opponent_plays.map((play) => renderPlay(play))}
              </div>
            ) : (
              <div className="no-plays">No plays</div>
            )}
          </div>
        </div>
      </div>
    );
  };

  const selectedEntry = timeline.find((e) => e.turn === selectedTurn);

  return (
    <div className="game-play-timeline-panel">
      <button
        type="button"
        className="panel-header"
        onClick={onToggle}
        aria-expanded={isExpanded}
        aria-controls="game-play-timeline-content"
      >
        <h3>Game Timeline</h3>
        <span className={`expand-icon ${isExpanded ? 'expanded' : ''}`} aria-hidden="true">
          {isExpanded ? '\u25BC' : '\u25B6'}
        </span>
      </button>

      {isExpanded && (
        <div id="game-play-timeline-content" className="panel-content">
          {loading && <LoadingSpinner message="Loading timeline..." />}

          {error && <div className="error-message">{error}</div>}

          {!loading && !error && timeline.length === 0 && (
            <div className="empty-state">No game play data available for this match</div>
          )}

          {!loading && !error && timeline.length > 0 && (
            <div className="timeline-container">
              {/* Turn Navigation */}
              <div className="turn-navigation">
                <div className="turn-buttons">
                  {timeline.map((entry) => (
                    <button
                      key={entry.turn}
                      type="button"
                      className={`turn-btn ${selectedTurn === entry.turn ? 'active' : ''}`}
                      onClick={() => setSelectedTurn(entry.turn)}
                    >
                      T{entry.turn}
                    </button>
                  ))}
                </div>
              </div>

              {/* Selected Turn Details */}
              {selectedEntry && (
                <div className="selected-turn">
                  <div className="turn-header">
                    <h4>Turn {selectedEntry.turn}</h4>
                    <span className="active-player">
                      Active: {selectedEntry.active_player === 'player' ? 'You' : 'Opponent'}
                    </span>
                  </div>
                  {renderTurnDetails(selectedEntry)}
                </div>
              )}
            </div>
          )}
        </div>
      )}
    </div>
  );
};

export default GamePlayTimelinePanel;
