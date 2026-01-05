/**
 * Standard format API service.
 * Provides functions for Standard legality validation and set management.
 */

import { get, post } from '../apiClient';

// Types

export interface StandardSet {
  code: string;
  name: string;
  releasedAt: string;
  rotationDate?: string;
  isStandardLegal: boolean;
  iconSvgUri: string;
  cardCount: number;
  daysUntilRotation?: number;
  isRotatingSoon: boolean;
}

export interface StandardConfig {
  id: number;
  nextRotationDate: string;
  rotationEnabled: boolean;
  updatedAt: string;
}

export interface CardLegality {
  standard: string;
  historic: string;
  explorer: string;
  pioneer: string;
  modern: string;
  alchemy: string;
  brawl: string;
  commander: string;
}

export interface RotatingCard {
  cardId: number;
  cardName: string;
  setCode: string;
  setName: string;
  rotationDate: string;
  daysUntilRotation: number;
}

export interface DeckSetInfo {
  setCode: string;
  setName: string;
  cardCount: number;
  iconSvgUri: string;
  isRotating: boolean;
}

export interface ValidationError {
  cardId: number;
  cardName: string;
  reason: string;
  details: string;
}

export interface ValidationWarning {
  cardId: number;
  cardName: string;
  type: string;
  details: string;
}

export interface DeckValidationResult {
  isLegal: boolean;
  errors: ValidationError[];
  warnings: ValidationWarning[];
  rotatingCards: RotatingCard[];
  setBreakdown: DeckSetInfo[];
}

export interface RotationAffectedDeck {
  deckId: string;
  deckName: string;
  format: string;
  rotatingCardCount: number;
  totalCards: number;
  percentAffected: number;
  rotatingCards: RotatingCard[];
}

export interface UpcomingRotation {
  nextRotationDate: string;
  daysUntilRotation: number;
  rotatingSets: StandardSet[];
  rotatingCardCount: number;
  affectedDecks: number;
}

// API Functions

/**
 * Get all Standard-legal sets.
 */
export async function getStandardSets(): Promise<StandardSet[]> {
  return get<StandardSet[]>('/standard/sets');
}

/**
 * Get upcoming rotation information.
 */
export async function getUpcomingRotation(): Promise<UpcomingRotation> {
  return get<UpcomingRotation>('/standard/rotation');
}

/**
 * Get decks affected by the upcoming rotation.
 */
export async function getRotationAffectedDecks(): Promise<RotationAffectedDeck[]> {
  return get<RotationAffectedDeck[]>('/standard/rotation/affected-decks');
}

/**
 * Get Standard configuration.
 */
export async function getStandardConfig(): Promise<StandardConfig> {
  return get<StandardConfig>('/standard/config');
}

/**
 * Validate a deck for Standard legality.
 */
export async function validateDeckStandard(deckId: string): Promise<DeckValidationResult> {
  return post<DeckValidationResult>(`/standard/validate/${deckId}`);
}

/**
 * Get the legality of a card.
 */
export async function getCardLegality(arenaId: string): Promise<CardLegality> {
  return get<CardLegality>(`/standard/cards/${arenaId}/legality`);
}
