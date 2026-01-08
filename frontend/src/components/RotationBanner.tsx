import { useState } from 'react';
import { Link } from 'react-router-dom';
import type { UpcomingRotation, RotationAffectedDeck } from '@/services/api/standard';
import './RotationBanner.css';

interface RotationBannerProps {
  rotation: UpcomingRotation;
  affectedDecks: RotationAffectedDeck[];
  onDismiss?: () => void;
  compact?: boolean;
}

export function RotationBanner({
  rotation,
  affectedDecks,
  onDismiss,
  compact = false,
}: RotationBannerProps) {
  const [isExpanded, setIsExpanded] = useState(false);

  const days = rotation.daysUntilRotation;
  const urgency = days <= 7 ? 'critical' : days <= 30 ? 'warning' : 'info';

  // Format rotation date
  const rotationDate = new Date(rotation.nextRotationDate);
  const formattedDate = rotationDate.toLocaleDateString(undefined, {
    month: 'long',
    day: 'numeric',
    year: 'numeric',
  });

  if (compact) {
    return (
      <div className={`rotation-banner rotation-banner--${urgency} rotation-banner--compact`}>
        <span className="rotation-banner__icon">
          {urgency === 'critical' ? '!' : urgency === 'warning' ? '!' : 'i'}
        </span>
        <span className="rotation-banner__text">
          {affectedDecks.length} deck{affectedDecks.length !== 1 ? 's' : ''} affected by rotation in{' '}
          {days} day{days !== 1 ? 's' : ''}
        </span>
        <Link to="/decks?filter=rotating" className="rotation-banner__link">
          View
        </Link>
      </div>
    );
  }

  return (
    <div className={`rotation-banner rotation-banner--${urgency}`}>
      <div className="rotation-banner__header">
        <div className="rotation-banner__icon-container">
          <span className="rotation-banner__icon">
            {urgency === 'critical' ? '!' : urgency === 'warning' ? '!' : 'i'}
          </span>
        </div>
        <div className="rotation-banner__content">
          <h3 className="rotation-banner__title">
            Standard Rotation in {days} Day{days !== 1 ? 's' : ''}
          </h3>
          <p className="rotation-banner__subtitle">
            {affectedDecks.length} of your deck{affectedDecks.length !== 1 ? 's' : ''} will lose
            cards on {formattedDate}
          </p>
        </div>
        <div className="rotation-banner__actions">
          <button
            className="rotation-banner__expand"
            onClick={() => setIsExpanded(!isExpanded)}
            aria-label={isExpanded ? 'Collapse details' : 'Expand details'}
          >
            {isExpanded ? 'Hide' : 'Details'}
          </button>
          {onDismiss && (
            <button
              className="rotation-banner__dismiss"
              onClick={onDismiss}
              aria-label="Dismiss notification"
            >
              x
            </button>
          )}
        </div>
      </div>

      {isExpanded && (
        <div className="rotation-banner__details">
          <div className="rotation-banner__section">
            <h4>Rotating Sets</h4>
            <ul className="rotation-banner__sets">
              {rotation.rotatingSets.map((set) => (
                <li key={set.code} className="rotation-banner__set">
                  <span className="rotation-banner__set-code">{set.code}</span>
                  <span className="rotation-banner__set-name">{set.name}</span>
                </li>
              ))}
            </ul>
          </div>

          <div className="rotation-banner__section">
            <h4>Affected Decks</h4>
            <ul className="rotation-banner__decks">
              {affectedDecks.slice(0, 5).map((deck) => (
                <li key={deck.deckId} className="rotation-banner__deck">
                  <Link to={`/decks/${deck.deckId}`} className="rotation-banner__deck-link">
                    {deck.deckName}
                  </Link>
                  <span className="rotation-banner__deck-impact">
                    {deck.rotatingCardCount} card{deck.rotatingCardCount !== 1 ? 's' : ''} (
                    {deck.percentAffected.toFixed(0)}%)
                  </span>
                </li>
              ))}
              {affectedDecks.length > 5 && (
                <li className="rotation-banner__deck rotation-banner__deck--more">
                  <Link to="/decks?filter=rotating">
                    +{affectedDecks.length - 5} more deck{affectedDecks.length - 5 !== 1 ? 's' : ''}
                  </Link>
                </li>
              )}
            </ul>
          </div>
        </div>
      )}
    </div>
  );
}

export default RotationBanner;
