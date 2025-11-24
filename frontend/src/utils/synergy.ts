/**
 * Synergy detection utilities for draft card recommendations
 * Implements algorithms from Issue #384
 */

import { models } from '../../wailsjs/go/models';

export interface TypeSynergy {
    type: string;
    count: number;
    percentage: number;
}

export interface ColorSynergy {
    colors: string[];
    count: number;
    percentage: number;
}

export interface CurveSynergy {
    avgCMC: number;
    archetype: 'aggro' | 'midrange' | 'control';
    gaps: number[]; // CMC values that are underrepresented
}

export interface SynergyAnalysis {
    types: TypeSynergy[];
    colors: ColorSynergy;
    curve: CurveSynergy;
    pickedCardsCount: number;
}

/**
 * Analyzes picked cards to detect synergies
 */
export function analyzeSynergies(pickedCards: models.SetCard[]): SynergyAnalysis {
    if (pickedCards.length === 0) {
        return {
            types: [],
            colors: { colors: [], count: 0, percentage: 0 },
            curve: { avgCMC: 0, archetype: 'midrange', gaps: [] },
            pickedCardsCount: 0,
        };
    }

    // Analyze card types
    const types = analyzeTypes(pickedCards);

    // Analyze colors
    const colors = analyzeColors(pickedCards);

    // Analyze mana curve
    const curve = analyzeCurve(pickedCards);

    return {
        types,
        colors,
        curve,
        pickedCardsCount: pickedCards.length,
    };
}

/**
 * Extracts and counts card types from picked cards
 */
function analyzeTypes(pickedCards: models.SetCard[]): TypeSynergy[] {
    const typeCount = new Map<string, number>();

    pickedCards.forEach(card => {
        if (!card.Types || card.Types.length === 0) return;

        card.Types.forEach(type => {
            // Normalize type (capitalize first letter)
            const normalizedType = type.charAt(0).toUpperCase() + type.slice(1).toLowerCase();
            typeCount.set(normalizedType, (typeCount.get(normalizedType) || 0) + 1);
        });
    });

    // Convert to array and sort by count (descending)
    const types: TypeSynergy[] = Array.from(typeCount.entries())
        .map(([type, count]) => ({
            type,
            count,
            percentage: (count / pickedCards.length) * 100,
        }))
        .sort((a, b) => b.count - a.count);

    return types;
}

/**
 * Analyzes color distribution in picked cards
 */
function analyzeColors(pickedCards: models.SetCard[]): ColorSynergy {
    const colorCount = new Map<string, number>();

    pickedCards.forEach(card => {
        if (!card.Colors || card.Colors.length === 0) {
            // Colorless
            colorCount.set('C', (colorCount.get('C') || 0) + 1);
            return;
        }

        card.Colors.forEach(color => {
            colorCount.set(color, (colorCount.get(color) || 0) + 1);
        });
    });

    // Find dominant colors (appearing in 30%+ of cards)
    const dominantColors = Array.from(colorCount.entries())
        .filter(([, count]) => (count / pickedCards.length) >= 0.3)
        .sort((a, b) => b[1] - a[1])
        .map(([color]) => color);

    return {
        colors: dominantColors,
        count: dominantColors.reduce((sum, color) => sum + (colorCount.get(color) || 0), 0),
        percentage: dominantColors.length > 0
            ? (dominantColors.reduce((sum, color) => sum + (colorCount.get(color) || 0), 0) / pickedCards.length) * 100
            : 0,
    };
}

/**
 * Analyzes mana curve and determines archetype
 */
function analyzeCurve(pickedCards: models.SetCard[]): CurveSynergy {
    const creatures = pickedCards.filter(card =>
        card.Types && card.Types.some(t => t.toLowerCase() === 'creature')
    );

    if (creatures.length === 0) {
        return { avgCMC: 0, archetype: 'midrange', gaps: [] };
    }

    // Calculate average CMC for creatures
    const totalCMC = creatures.reduce((sum, card) => sum + card.CMC, 0);
    const avgCMC = totalCMC / creatures.length;

    // Determine archetype based on average CMC
    let archetype: 'aggro' | 'midrange' | 'control';
    if (avgCMC <= 2.5) {
        archetype = 'aggro';
    } else if (avgCMC <= 3.5) {
        archetype = 'midrange';
    } else {
        archetype = 'control';
    }

    // Find gaps in curve (CMC values with 0-1 cards)
    const cmcDistribution = new Map<number, number>();
    creatures.forEach(card => {
        const cmc = Math.min(card.CMC, 7); // Cap at 7+ for curve purposes
        cmcDistribution.set(cmc, (cmcDistribution.get(cmc) || 0) + 1);
    });

    const gaps: number[] = [];
    for (let cmc = 1; cmc <= 6; cmc++) {
        const count = cmcDistribution.get(cmc) || 0;
        if (count <= 1) {
            gaps.push(cmc);
        }
    }

    return { avgCMC, archetype, gaps };
}

/**
 * Calculates synergy score for a card based on picked cards
 * Returns a value 0-100 indicating how well the card synergizes
 */
export function calculateCardSynergyScore(
    card: models.SetCard,
    synergy: SynergyAnalysis
): number {
    if (synergy.pickedCardsCount === 0) return 0;

    let score = 0;

    // Type synergy (max 50 points)
    if (card.Types && card.Types.length > 0) {
        const cardTypes = card.Types.map(t => t.charAt(0).toUpperCase() + t.slice(1).toLowerCase());

        for (const cardType of cardTypes) {
            const typeSynergy = synergy.types.find(t => t.type === cardType);
            if (typeSynergy) {
                // Give more weight to prominent types
                if (typeSynergy.percentage >= 40) {
                    score += 30; // Strong tribal/type synergy
                } else if (typeSynergy.percentage >= 20) {
                    score += 20; // Moderate synergy
                } else if (typeSynergy.percentage >= 10) {
                    score += 10; // Minor synergy
                }
                break; // Only count once per card
            }
        }
    }

    // Color synergy (max 30 points)
    if (synergy.colors.colors.length > 0) {
        if (!card.Colors || card.Colors.length === 0) {
            // Colorless - always fits
            score += 15;
        } else {
            const matchingColors = card.Colors.filter(c => synergy.colors.colors.includes(c));
            if (matchingColors.length === card.Colors.length) {
                // All colors match
                score += 30;
            } else if (matchingColors.length > 0) {
                // Partial match (hybrid card)
                score += 15;
            }
        }
    }

    // Curve synergy (max 20 points)
    const isCreature = card.Types && card.Types.some(t => t.toLowerCase() === 'creature');
    if (isCreature) {
        const cmc = Math.min(card.CMC, 7);

        // If card fills a gap in the curve
        if (synergy.curve.gaps.includes(cmc)) {
            score += 20;
        } else if (synergy.curve.archetype === 'aggro' && cmc <= 3) {
            score += 10; // Low-cost creature for aggro
        } else if (synergy.curve.archetype === 'control' && cmc >= 5) {
            score += 10; // High-cost finisher for control
        } else if (synergy.curve.archetype === 'midrange' && cmc >= 2 && cmc <= 5) {
            score += 10; // Mid-range creature
        }
    }

    return Math.min(score, 100);
}

/**
 * Gets highlighted card types (types appearing in 20%+ of picked cards)
 */
export function getHighlightedTypes(synergy: SynergyAnalysis): string[] {
    return synergy.types
        .filter(t => t.percentage >= 20)
        .map(t => t.type);
}

/**
 * Checks if a card should be highlighted based on synergy
 */
export function shouldHighlightCard(card: models.SetCard, synergy: SynergyAnalysis): boolean {
    const score = calculateCardSynergyScore(card, synergy);
    return score >= 30; // Highlight if synergy score is at least 30
}

/**
 * Gets synergy reason text for display in UI
 */
export function getSynergyReason(card: models.SetCard, synergy: SynergyAnalysis): string {
    const reasons: string[] = [];

    // Check type synergy
    if (card.Types && card.Types.length > 0) {
        const cardTypes = card.Types.map(t => t.charAt(0).toUpperCase() + t.slice(1).toLowerCase());
        for (const cardType of cardTypes) {
            const typeSynergy = synergy.types.find(t => t.type === cardType);
            if (typeSynergy && typeSynergy.percentage >= 20) {
                reasons.push(`${cardType} synergy (${typeSynergy.count} cards)`);
                break;
            }
        }
    }

    // Check color synergy
    if (synergy.colors.colors.length > 0) {
        if (!card.Colors || card.Colors.length === 0) {
            reasons.push('Colorless - fits any deck');
        } else {
            const matchingColors = card.Colors.filter(c => synergy.colors.colors.includes(c));
            if (matchingColors.length === card.Colors.length && card.Colors.length > 0) {
                reasons.push(`Matches your colors (${synergy.colors.colors.join('')})`);
            }
        }
    }

    // Check curve synergy
    const isCreature = card.Types && card.Types.some(t => t.toLowerCase() === 'creature');
    if (isCreature) {
        const cmc = Math.min(card.CMC, 7);
        if (synergy.curve.gaps.includes(cmc)) {
            reasons.push(`Fills ${cmc}-drop gap`);
        } else if (synergy.curve.archetype === 'aggro' && cmc <= 2) {
            reasons.push('Low cost for aggro');
        } else if (synergy.curve.archetype === 'control' && cmc >= 5) {
            reasons.push('Finisher for control');
        }
    }

    return reasons.join(' • ') || 'Good fit for your deck';
}
