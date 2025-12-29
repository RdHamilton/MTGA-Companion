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
 * Normalizes a queue type to a user-friendly display name.
 *
 * Examples:
 * - 'Play' -> 'Play Queue'
 * - 'Ladder' -> 'Ranked'
 * - 'QuickDraft_TLA_20251127' -> 'QuickDraft'
 * - 'TradDraft_MKM' -> 'Traditional Draft'
 *
 * @param queueType - The raw queue type from MTGA
 * @returns The normalized, user-friendly queue type name
 */
export function normalizeQueueType(queueType: string): string {
  if (!queueType) return queueType;

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
 * - QuickDraft_TLA -> 'QuickDraft'
 *
 * @param match - The match object
 * @returns The event name to display in the Event column
 */
export function getDisplayEventName(match: models.Match): string {
  const queueName = normalizeQueueType(match.EventName || match.Format);
  // If we have a deck format and it's a constructed queue, combine them
  if (match.DeckFormat && ['Play Queue', 'Ranked', 'Traditional Ranked', 'Traditional Play'].includes(queueName)) {
    return `${match.DeckFormat} ${queueName}`;
  }
  return queueName;
}
