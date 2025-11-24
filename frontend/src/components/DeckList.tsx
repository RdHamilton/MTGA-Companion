import { useState, useEffect, useMemo } from 'react';
import { BarChart, Bar, PieChart, Pie, Cell, XAxis, YAxis, Tooltip, ResponsiveContainer } from 'recharts';
import { GetCardByArenaID } from '../../wailsjs/go/main/App';
import { models, gui } from '../../wailsjs/go/models';
import './DeckList.css';

interface DeckListProps {
  deck: models.Deck;
  cards: models.DeckCard[];
  tags?: models.DeckTag[];
  statistics?: gui.DeckStatistics;
  onRemoveCard?: (cardID: number, board: string) => void;
  onCardHover?: (card: models.SetCard) => void;
}

interface CardWithMetadata {
  deckCard: models.DeckCard;
  metadata?: models.SetCard;
}

interface GroupedCards {
  creatures: CardWithMetadata[];
  instants: CardWithMetadata[];
  sorceries: CardWithMetadata[];
  enchantments: CardWithMetadata[];
  artifacts: CardWithMetadata[];
  planeswalkers: CardWithMetadata[];
  lands: CardWithMetadata[];
  other: CardWithMetadata[];
}

const COLOR_MAP: { [key: string]: string } = {
  white: '#fffbd5',
  blue: '#0e68ab',
  black: '#150b00',
  red: '#d3202a',
  green: '#00733e',
  colorless: '#ccc',
  multicolor: '#888',
};

// Basic land names as fallback for when metadata is missing
const BASIC_LAND_NAMES: { [key: number]: string } = {
  81716: 'Plains',
  81717: 'Island',
  81718: 'Swamp',
  81719: 'Mountain',
  81720: 'Forest',
};

const getCardName = (cardID: number, metadata?: models.SetCard): string => {
  if (metadata?.Name) return metadata.Name;
  if (BASIC_LAND_NAMES[cardID]) return BASIC_LAND_NAMES[cardID];
  return `Unknown Card ${cardID}`;
};

export default function DeckList({
  deck,
  cards,
  tags = [],
  statistics,
  onRemoveCard,
  onCardHover,
}: DeckListProps) {
  const [cardsWithMetadata, setCardsWithMetadata] = useState<CardWithMetadata[]>([]);
  const [loading, setLoading] = useState(cards.length > 0);
  const [showSideboard, setShowSideboard] = useState(false);

  // Load card metadata for all cards
  useEffect(() => {
    let isMounted = true;

    const loadMetadata = async () => {
      // Handle empty cards case
      if (cards.length === 0) {
        if (isMounted) {
          setCardsWithMetadata([]);
          setLoading(false);
        }
        return;
      }

      setLoading(true);
      const withMetadata: CardWithMetadata[] = [];

      for (const card of cards) {
        try {
          const metadata = await GetCardByArenaID(String(card.CardID));
          withMetadata.push({ deckCard: card, metadata });
        } catch (err) {
          console.error(`Failed to load metadata for card ${card.CardID}:`, err);
          withMetadata.push({ deckCard: card });
        }
      }

      if (isMounted) {
        setCardsWithMetadata(withMetadata);
        setLoading(false);
      }
    };

    loadMetadata();

    return () => {
      isMounted = false;
    };
  }, [cards]);

  // Group cards by type (mainboard only)
  const groupedMainboard = useMemo((): GroupedCards => {
    const groups: GroupedCards = {
      creatures: [],
      instants: [],
      sorceries: [],
      enchantments: [],
      artifacts: [],
      planeswalkers: [],
      lands: [],
      other: [],
    };

    cardsWithMetadata
      .filter((c) => c.deckCard.Board === 'main')
      .forEach((card) => {
        // Check if this is a basic land by ID (even without metadata)
        const basicLandIDs = [81716, 81717, 81718, 81719, 81720];
        if (basicLandIDs.includes(card.deckCard.CardID)) {
          groups.lands.push(card);
          return;
        }

        if (!card.metadata) {
          groups.other.push(card);
          return;
        }

        const typeLine = (card.metadata.Types || []).join(' ').toLowerCase();
        if (typeLine.includes('creature')) {
          groups.creatures.push(card);
        } else if (typeLine.includes('instant')) {
          groups.instants.push(card);
        } else if (typeLine.includes('sorcery')) {
          groups.sorceries.push(card);
        } else if (typeLine.includes('enchantment')) {
          groups.enchantments.push(card);
        } else if (typeLine.includes('artifact')) {
          groups.artifacts.push(card);
        } else if (typeLine.includes('planeswalker')) {
          groups.planeswalkers.push(card);
        } else if (typeLine.includes('land')) {
          groups.lands.push(card);
        } else {
          groups.other.push(card);
        }
      });

    // Sort each group by CMC, then alphabetically
    Object.values(groups).forEach((group) => {
      group.sort((a: CardWithMetadata, b: CardWithMetadata) => {
        if (!a.metadata || !b.metadata) return 0;
        if (a.metadata.CMC !== b.metadata.CMC) {
          return a.metadata.CMC - b.metadata.CMC;
        }
        return a.metadata.Name.localeCompare(b.metadata.Name);
      });
    });

    return groups;
  }, [cardsWithMetadata]);

  // Get sideboard cards
  const sideboardCards = useMemo(() => {
    return cardsWithMetadata.filter((c) => c.deckCard.Board === 'sideboard');
  }, [cardsWithMetadata]);

  // Calculate totals
  const mainboardCount = cards.filter((c) => c.Board === 'main').reduce((sum, c) => sum + c.Quantity, 0);
  const sideboardCount = cards.filter((c) => c.Board === 'sideboard').reduce((sum, c) => sum + c.Quantity, 0);

  // Mana curve data
  const manaCurveData = useMemo(() => {
    if (!statistics?.manaCurve) return [];
    return Object.entries(statistics.manaCurve).map(([cmc, count]) => ({
      cmc: cmc === '7' ? '7+' : cmc,
      count,
    }));
  }, [statistics]);

  // Color distribution data
  const colorData = useMemo(() => {
    if (!statistics?.colors) return [];
    const colors = statistics.colors;
    return [
      { name: 'White', value: colors.white, color: COLOR_MAP.white },
      { name: 'Blue', value: colors.blue, color: COLOR_MAP.blue },
      { name: 'Black', value: colors.black, color: COLOR_MAP.black },
      { name: 'Red', value: colors.red, color: COLOR_MAP.red },
      { name: 'Green', value: colors.green, color: COLOR_MAP.green },
      { name: 'Colorless', value: colors.colorless, color: COLOR_MAP.colorless },
      { name: 'Multicolor', value: colors.multicolor, color: COLOR_MAP.multicolor },
    ].filter((item) => item.value > 0);
  }, [statistics]);

  const renderCardGroup = (title: string, cards: CardWithMetadata[], count: number) => {
    if (count === 0) return null;

    return (
      <div className="card-group">
        <div className="group-header">
          <h4>{title}</h4>
          <span className="group-count">({count})</span>
        </div>
        <div className="group-cards">
          {cards.map((card) => (
            <div
              key={`${card.deckCard.CardID}-${card.deckCard.Board}`}
              className="deck-card"
              onMouseEnter={() => card.metadata && onCardHover?.(card.metadata)}
              title={card.metadata?.Name || `Card ${card.deckCard.CardID}`}
            >
              <span className="card-quantity">{card.deckCard.Quantity}x</span>
              <span className="card-name">{getCardName(card.deckCard.CardID, card.metadata)}</span>
              {card.metadata?.ManaCost && <span className="card-mana">{card.metadata.ManaCost}</span>}
              {onRemoveCard && (
                <button
                  className="remove-card-btn"
                  onClick={() => onRemoveCard(card.deckCard.CardID, card.deckCard.Board)}
                  title="Remove card"
                >
                  ×
                </button>
              )}
            </div>
          ))}
        </div>
      </div>
    );
  };

  if (loading) {
    return <div className="deck-list loading">Loading deck...</div>;
  }

  return (
    <div className="deck-list">
      {/* Deck Header */}
      <div className="deck-header">
        <div className="deck-title">
          <h2>{deck.Name}</h2>
          {deck.Source === 'draft' && deck.DraftEventID && (
            <span className="draft-indicator">Draft Deck</span>
          )}
        </div>
        <div className="deck-meta">
          <span className="deck-format">{deck.Format}</span>
          <span className="deck-source">{deck.Source}</span>
          {tags.length > 0 && (
            <div className="deck-tags">
              {tags.map((tag) => (
                <span key={tag.ID} className="deck-tag">
                  {tag.Tag}
                </span>
              ))}
            </div>
          )}
        </div>
        <div className="deck-counts">
          <span className="count-badge mainboard">Mainboard: {mainboardCount}</span>
          <span className="count-badge sideboard">Sideboard: {sideboardCount}</span>
        </div>
      </div>

      {/* Statistics Charts */}
      {statistics && (
        <div className="deck-statistics">
          {/* Mana Curve */}
          {manaCurveData.length > 0 && (
            <div className="stat-chart">
              <h3>Mana Curve</h3>
              <ResponsiveContainer width="100%" height={200}>
                <BarChart data={manaCurveData}>
                  <XAxis dataKey="cmc" stroke="#aaa" />
                  <YAxis stroke="#aaa" />
                  <Tooltip
                    contentStyle={{ background: '#2a2a2a', border: '1px solid #444' }}
                    labelStyle={{ color: '#fff' }}
                  />
                  <Bar dataKey="count" fill="#4a9eff" />
                </BarChart>
              </ResponsiveContainer>
              <div className="avg-cmc">Average CMC: {statistics.averageCMC?.toFixed(2) || 'N/A'}</div>
            </div>
          )}

          {/* Color Distribution */}
          {colorData.length > 0 && (
            <div className="stat-chart">
              <h3>Color Distribution</h3>
              <ResponsiveContainer width="100%" height={250}>
                <PieChart>
                  <Pie
                    data={colorData}
                    cx="50%"
                    cy="55%"
                    outerRadius={75}
                    fill="#8884d8"
                    dataKey="value"
                    label={(entry) => `${entry.name}: ${entry.value}`}
                  >
                    {colorData.map((entry, index) => (
                      <Cell key={`cell-${index}`} fill={entry.color} />
                    ))}
                  </Pie>
                  <Tooltip
                    contentStyle={{ background: '#2a2a2a', border: '1px solid #444' }}
                  />
                </PieChart>
              </ResponsiveContainer>
            </div>
          )}

          {/* Land Recommendation */}
          {statistics.lands && (
            <div className="land-recommendation">
              <div className={`land-status ${statistics.lands.status}`}>
                {statistics.lands.statusMessage}
              </div>
            </div>
          )}
        </div>
      )}

      {/* Mainboard Cards */}
      <div className="mainboard">
        <h3>Mainboard ({mainboardCount})</h3>
        <div className="card-groups">
          {renderCardGroup('Creatures', groupedMainboard.creatures, groupedMainboard.creatures.reduce((sum, c) => sum + c.deckCard.Quantity, 0))}
          {renderCardGroup('Instants', groupedMainboard.instants, groupedMainboard.instants.reduce((sum, c) => sum + c.deckCard.Quantity, 0))}
          {renderCardGroup('Sorceries', groupedMainboard.sorceries, groupedMainboard.sorceries.reduce((sum, c) => sum + c.deckCard.Quantity, 0))}
          {renderCardGroup('Enchantments', groupedMainboard.enchantments, groupedMainboard.enchantments.reduce((sum, c) => sum + c.deckCard.Quantity, 0))}
          {renderCardGroup('Artifacts', groupedMainboard.artifacts, groupedMainboard.artifacts.reduce((sum, c) => sum + c.deckCard.Quantity, 0))}
          {renderCardGroup('Planeswalkers', groupedMainboard.planeswalkers, groupedMainboard.planeswalkers.reduce((sum, c) => sum + c.deckCard.Quantity, 0))}
          {renderCardGroup('Lands', groupedMainboard.lands, groupedMainboard.lands.reduce((sum, c) => sum + c.deckCard.Quantity, 0))}
          {groupedMainboard.other.length > 0 && renderCardGroup('Other', groupedMainboard.other, groupedMainboard.other.reduce((sum, c) => sum + c.deckCard.Quantity, 0))}
        </div>
      </div>

      {/* Sideboard */}
      {sideboardCount > 0 && (
        <div className="sideboard">
          <div className="sideboard-header" onClick={() => setShowSideboard(!showSideboard)}>
            <h3>Sideboard ({sideboardCount})</h3>
            <button className="toggle-sideboard">
              {showSideboard ? '▼' : '▶'}
            </button>
          </div>
          {showSideboard && (
            <div className="sideboard-cards">
              {sideboardCards.map((card) => (
                <div
                  key={`${card.deckCard.CardID}-sideboard`}
                  className="deck-card"
                  onMouseEnter={() => card.metadata && onCardHover?.(card.metadata)}
                  title={card.metadata?.Name || `Card ${card.deckCard.CardID}`}
                >
                  <span className="card-quantity">{card.deckCard.Quantity}x</span>
                  <span className="card-name">{getCardName(card.deckCard.CardID, card.metadata)}</span>
                  {card.metadata?.ManaCost && <span className="card-mana">{card.metadata.ManaCost}</span>}
                  {onRemoveCard && (
                    <button
                      className="remove-card-btn"
                      onClick={() => onRemoveCard(card.deckCard.CardID, card.deckCard.Board)}
                      title="Remove card"
                    >
                      ×
                    </button>
                  )}
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      {cards.length === 0 && (
        <div className="empty-deck">
          <p>No cards in deck yet. Use the card search to add cards.</p>
        </div>
      )}
    </div>
  );
}
