import { models } from '@/types/models';

/**
 * Queue type mapping from MTGA event IDs to user-friendly names.
 */
const QUEUE_TYPE_MAP: Record<string, string> = {
  'Play': 'Play Queue',
  'Ladder': 'Ranked',
  'Traditional_Ladder': 'Traditional Ranked',
  'Traditional_Play': 'Traditional Play',
};

/**
 * Known draft format prefixes that should be normalized.
 */
const DRAFT_PREFIXES = ['QuickDraft', 'PremierDraft', 'TradDraft', 'SealedDeck'];

/**
 * Format-specific event IDs that contain the format name.
 * These map to their queue type suffix.
 */
const FORMAT_EVENT_PATTERNS: Record<string, string> = {
  // Alchemy events
  'Alchemy': 'Play Queue',
  'Alchemy_Play': 'Play Queue',
  'Alchemy_Ladder': 'Ranked',
  // Historic Brawl events
  'HistoricBrawl': 'Play Queue',
  'HistoricBrawl_Play': 'Play Queue',
  // Brawl events
  'Brawl': 'Play Queue',
  'Brawl_Play': 'Play Queue',
  // Explorer events
  'Explorer': 'Play Queue',
  'Explorer_Play': 'Play Queue',
  'Explorer_Ladder': 'Ranked',
  // Historic events (when not using generic Play/Ladder)
  'Historic': 'Play Queue',
  'Historic_Play': 'Play Queue',
  'Historic_Ladder': 'Ranked',
  // Timeless events
  'Timeless': 'Play Queue',
  'Timeless_Play': 'Play Queue',
  'Timeless_Ladder': 'Ranked',
};

/**
 * Normalizes a queue type to a user-friendly display name.
 *
 * Examples:
 * - 'Play' -> 'Play Queue'
 * - 'Ladder' -> 'Ranked'
 * - 'QuickDraft_TLA_20251127' -> 'QuickDraft'
 * - 'TradDraft_MKM' -> 'Traditional Draft'
 * - 'Alchemy' -> 'Play Queue'
 * - 'HistoricBrawl_Play' -> 'Play Queue'
 *
 * @param queueType - The raw queue type from MTGA
 * @returns The normalized, user-friendly queue type name
 */
export function normalizeQueueType(queueType: string): string {
  if (!queueType) return queueType;

  // Check for format-specific event patterns first
  if (FORMAT_EVENT_PATTERNS[queueType]) {
    return FORMAT_EVENT_PATTERNS[queueType];
  }

  // Check if it's a draft format (contains underscore with set code pattern)
  const underscoreIndex = queueType.indexOf('_');
  if (underscoreIndex !== -1) {
    const prefix = queueType.substring(0, underscoreIndex);
    // Known draft prefixes
    if (DRAFT_PREFIXES.includes(prefix)) {
      return prefix
        .replace('TradDraft', 'Traditional Draft')
        .replace('SealedDeck', 'Sealed');
    }
    // Check if it's a mapped queue type with underscore
    if (QUEUE_TYPE_MAP[queueType]) {
      return QUEUE_TYPE_MAP[queueType];
    }
    // Otherwise just return the prefix
    return prefix;
  }

  return QUEUE_TYPE_MAP[queueType] || queueType;
}

/**
 * Gets the display format for a match.
 * Prefers the deck format (Standard, Historic, etc.) over the queue type.
 *
 * @param match - The match object
 * @returns The format to display in the Format column
 */
export function getDisplayFormat(match: models.Match): string {
  // If we have a deck format, use it
  if (match.DeckFormat) {
    return match.DeckFormat;
  }
  // Fall back to normalized queue type
  return normalizeQueueType(match.Format);
}

/**
 * Gets the display event name for a match.
 * Combines deck format with queue type for constructed matches.
 *
 * Examples:
 * - Standard deck + Ladder -> 'Standard Ranked'
 * - Standard deck + Play -> 'Standard Play Queue'
 * - Alchemy deck + Alchemy -> 'Alchemy Play Queue'
 * - HistoricBrawl deck + HistoricBrawl -> 'HistoricBrawl Play Queue'
 * - QuickDraft_TLA -> 'QuickDraft'
 *
 * @param match - The match object
 * @returns The event name to display in the Event column
 */
export function getDisplayEventName(match: models.Match): string {
  const rawEvent = match.EventName || match.Format;
  const queueName = normalizeQueueType(rawEvent);

  // If we have a deck format, combine with normalized queue type
  if (match.DeckFormat && ['Play Queue', 'Ranked', 'Traditional Ranked', 'Traditional Play'].includes(queueName)) {
    return `${match.DeckFormat} ${queueName}`;
  }

  // For draft formats or when no deck format, just return the normalized queue name
  return queueName;
}
