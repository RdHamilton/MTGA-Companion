import { useState } from 'react';
import type { ValidationError, ValidationWarning } from '@/services/api/standard';
import './LegalityBanner.css';

interface LegalityBannerProps {
  isLegal: boolean;
  errors: ValidationError[];
  warnings: ValidationWarning[];
  format: string;
  onDismiss?: () => void;
  compact?: boolean;
}

export function LegalityBanner({
  isLegal,
  errors,
  warnings,
  format,
  onDismiss,
  compact = false,
}: LegalityBannerProps) {
  const [isExpanded, setIsExpanded] = useState(false);

  // Group errors by type
  const bannedCards = errors.filter((e) => e.reason === 'banned');
  const notLegalCards = errors.filter((e) => e.reason === 'not_legal');
  const tooManyCopies = errors.filter((e) => e.reason === 'too_many_copies');
  const deckSizeErrors = errors.filter((e) => e.reason === 'deck_size');

  const hasErrors = errors.length > 0;
  const hasWarnings = warnings.length > 0;

  // Don't render if deck is legal and has no warnings
  if (isLegal && !hasWarnings) {
    return null;
  }

  // Determine urgency based on error types
  const urgency = bannedCards.length > 0 ? 'critical' : hasErrors ? 'warning' : 'info';

  const formatName = format.charAt(0).toUpperCase() + format.slice(1);

  if (compact) {
    const issueCount = errors.length + warnings.length;
    return (
      <div className={`legality-banner legality-banner--${urgency} legality-banner--compact`}>
        <span className="legality-banner__icon">
          {urgency === 'critical' ? '!' : urgency === 'warning' ? '!' : 'i'}
        </span>
        <span className="legality-banner__text">
          {issueCount} legality issue{issueCount !== 1 ? 's' : ''} in {formatName}
        </span>
        <button
          className="legality-banner__link"
          onClick={() => setIsExpanded(!isExpanded)}
        >
          {isExpanded ? 'Hide' : 'View'}
        </button>
      </div>
    );
  }

  return (
    <div className={`legality-banner legality-banner--${urgency}`}>
      <div className="legality-banner__header">
        <div className="legality-banner__icon-container">
          <span className="legality-banner__icon">
            {urgency === 'critical' ? '!' : urgency === 'warning' ? '!' : 'i'}
          </span>
        </div>
        <div className="legality-banner__content">
          <h3 className="legality-banner__title">
            {bannedCards.length > 0
              ? `Deck Contains Banned Cards`
              : !isLegal
                ? `Deck Not Legal in ${formatName}`
                : `${formatName} Legality Warnings`}
          </h3>
          <p className="legality-banner__subtitle">
            {bannedCards.length > 0 && (
              <span>
                {bannedCards.length} banned card{bannedCards.length !== 1 ? 's' : ''}
              </span>
            )}
            {notLegalCards.length > 0 && (
              <span>
                {bannedCards.length > 0 ? ', ' : ''}
                {notLegalCards.length} card{notLegalCards.length !== 1 ? 's' : ''} not legal
              </span>
            )}
            {tooManyCopies.length > 0 && (
              <span>
                {(bannedCards.length > 0 || notLegalCards.length > 0) ? ', ' : ''}
                {tooManyCopies.length} card{tooManyCopies.length !== 1 ? 's' : ''} exceed 4-copy limit
              </span>
            )}
            {deckSizeErrors.length > 0 && (
              <span>
                {(bannedCards.length > 0 || notLegalCards.length > 0 || tooManyCopies.length > 0) ? ', ' : ''}
                deck size issue
              </span>
            )}
            {hasWarnings && !hasErrors && (
              <span>{warnings.length} warning{warnings.length !== 1 ? 's' : ''}</span>
            )}
          </p>
        </div>
        <div className="legality-banner__actions">
          <button
            className="legality-banner__expand"
            onClick={() => setIsExpanded(!isExpanded)}
            aria-label={isExpanded ? 'Collapse details' : 'Expand details'}
          >
            {isExpanded ? 'Hide' : 'Details'}
          </button>
          {onDismiss && (
            <button
              className="legality-banner__dismiss"
              onClick={onDismiss}
              aria-label="Dismiss notification"
            >
              ×
            </button>
          )}
        </div>
      </div>

      {isExpanded && (
        <div className="legality-banner__details">
          {bannedCards.length > 0 && (
            <div className="legality-banner__section">
              <h4>Banned Cards</h4>
              <ul className="legality-banner__cards">
                {bannedCards.map((card) => (
                  <li key={card.cardId} className="legality-banner__card legality-banner__card--banned">
                    <span className="legality-banner__card-name">{card.cardName || `Card #${card.cardId}`}</span>
                    <span className="legality-banner__card-badge">Banned</span>
                  </li>
                ))}
              </ul>
            </div>
          )}

          {notLegalCards.length > 0 && (
            <div className="legality-banner__section">
              <h4>Not Legal in {formatName}</h4>
              <ul className="legality-banner__cards">
                {notLegalCards.map((card) => (
                  <li key={card.cardId} className="legality-banner__card legality-banner__card--not-legal">
                    <span className="legality-banner__card-name">{card.cardName || `Card #${card.cardId}`}</span>
                    <span className="legality-banner__card-details">{card.details}</span>
                  </li>
                ))}
              </ul>
            </div>
          )}

          {tooManyCopies.length > 0 && (
            <div className="legality-banner__section">
              <h4>Too Many Copies</h4>
              <ul className="legality-banner__cards">
                {tooManyCopies.map((card) => (
                  <li key={card.cardId} className="legality-banner__card legality-banner__card--copies">
                    <span className="legality-banner__card-name">{card.cardName || `Card #${card.cardId}`}</span>
                    <span className="legality-banner__card-details">{card.details}</span>
                  </li>
                ))}
              </ul>
            </div>
          )}

          {deckSizeErrors.length > 0 && (
            <div className="legality-banner__section">
              <h4>Deck Size</h4>
              <ul className="legality-banner__cards">
                {deckSizeErrors.map((error, i) => (
                  <li key={i} className="legality-banner__card legality-banner__card--size">
                    <span className="legality-banner__card-details">{error.details}</span>
                  </li>
                ))}
              </ul>
            </div>
          )}

          {warnings.length > 0 && (
            <div className="legality-banner__section">
              <h4>Warnings</h4>
              <ul className="legality-banner__cards">
                {warnings.map((warning, i) => (
                  <li key={i} className="legality-banner__card legality-banner__card--warning">
                    <span className="legality-banner__card-name">
                      {warning.cardName || (warning.cardId ? `Card #${warning.cardId}` : 'Unknown')}
                    </span>
                    <span className="legality-banner__card-details">{warning.details}</span>
                  </li>
                ))}
              </ul>
            </div>
          )}
        </div>
      )}
    </div>
  );
}

export default LegalityBanner;
