/**
 * Game Plays API service.
 * Provides access to in-game play tracking data including card plays,
 * attacks, blocks, and turn-by-turn game state snapshots.
 */

import { get } from '../apiClient';

/**
 * Represents a single game play action.
 */
export interface GamePlay {
  id: number;
  game_id: number;
  match_id: string;
  turn_number: number;
  phase: string;
  step?: string;
  player_type: 'player' | 'opponent';
  action_type: 'play_card' | 'attack' | 'block' | 'land_drop' | 'mulligan';
  card_id?: number;
  card_name?: string;
  zone_from?: string;
  zone_to?: string;
  timestamp: string;
  sequence_number: number;
  created_at?: string;
}

/**
 * Represents a snapshot of game state at a specific turn.
 */
export interface GameStateSnapshot {
  id: number;
  game_id: number;
  match_id: string;
  turn_number: number;
  active_player: string;
  player_life?: number;
  opponent_life?: number;
  player_cards_in_hand?: number;
  opponent_cards_in_hand?: number;
  player_lands_in_play?: number;
  opponent_lands_in_play?: number;
  board_state_json?: string;
  timestamp: string;
}

/**
 * Represents a card observed from the opponent.
 */
export interface OpponentCard {
  id: number;
  game_id: number;
  match_id: string;
  card_id: number;
  card_name?: string;
  zone_observed: string;
  turn_first_seen: number;
  times_seen: number;
}

/**
 * Represents a timeline entry for a turn.
 */
export interface PlayTimelineEntry {
  turn: number;
  active_player: string;
  player_plays: GamePlay[];
  opponent_plays: GamePlay[];
  snapshot?: GameStateSnapshot;
}

/**
 * Represents a summary of plays for a match.
 */
export interface GamePlaySummary {
  match_id: string;
  game_id?: number;
  total_plays: number;
  player_plays: number;
  opponent_plays: number;
  card_plays: number;
  attacks: number;
  blocks: number;
  land_drops: number;
  total_turns: number;
  opponent_cards_seen: number;
}

/**
 * Get all plays for a specific match.
 */
export async function getMatchPlays(matchId: string): Promise<GamePlay[]> {
  return get<GamePlay[]>(`/matches/${encodeURIComponent(matchId)}/plays`);
}

/**
 * Get plays organized by turn for a specific match.
 */
export async function getMatchTimeline(matchId: string): Promise<PlayTimelineEntry[]> {
  return get<PlayTimelineEntry[]>(`/matches/${encodeURIComponent(matchId)}/plays/timeline`);
}

/**
 * Get a summary of plays for a match.
 */
export async function getMatchPlaySummary(matchId: string): Promise<GamePlaySummary> {
  return get<GamePlaySummary>(`/matches/${encodeURIComponent(matchId)}/plays/summary`);
}

/**
 * Get opponent cards observed during a match.
 */
export async function getMatchOpponentCards(matchId: string): Promise<OpponentCard[]> {
  return get<OpponentCard[]>(`/matches/${encodeURIComponent(matchId)}/opponent-cards`);
}

/**
 * Get game state snapshots for a match.
 * Optionally filter by game ID.
 */
export async function getMatchSnapshots(
  matchId: string,
  gameId?: number
): Promise<GameStateSnapshot[]> {
  const url = gameId
    ? `/matches/${encodeURIComponent(matchId)}/snapshots?gameID=${gameId}`
    : `/matches/${encodeURIComponent(matchId)}/snapshots`;
  return get<GameStateSnapshot[]>(url);
}

/**
 * Get plays for a specific game within a match.
 */
export async function getPlaysByGame(gameId: number): Promise<GamePlay[]> {
  return get<GamePlay[]>(`/gameplays/game/${gameId}`);
}

/**
 * Constants for player types.
 */
export const PlayerType = {
  Player: 'player',
  Opponent: 'opponent',
} as const;

/**
 * Constants for action types.
 */
export const ActionType = {
  PlayCard: 'play_card',
  Attack: 'attack',
  Block: 'block',
  LandDrop: 'land_drop',
  Mulligan: 'mulligan',
} as const;

/**
 * Constants for game phases.
 */
export const Phase = {
  Beginning: 'Beginning',
  Main1: 'Main1',
  Combat: 'Combat',
  Main2: 'Main2',
  Ending: 'Ending',
} as const;

/**
 * Constants for zone names.
 */
export const Zone = {
  Hand: 'hand',
  Library: 'library',
  Battlefield: 'battlefield',
  Graveyard: 'graveyard',
  Exile: 'exile',
  Stack: 'stack',
  Command: 'command',
} as const;
