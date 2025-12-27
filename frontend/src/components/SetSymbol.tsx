import { useState, useEffect, memo } from 'react';
import { cards } from '@/services/api';
import { gui } from '@/types/models';
import './SetSymbol.css';

// Get set info by code from the full list
async function getSetInfo(setCode: string): Promise<gui.SetInfo> {
  const sets = await cards.getAllSetInfo();
  const set = sets.find((s) => s.code === setCode);
  if (!set) {
    throw new Error(`Set not found: ${setCode}`);
  }
  return set;
}

interface SetSymbolProps {
  setCode: string;
  size?: 'small' | 'medium' | 'large';
  rarity?: 'common' | 'uncommon' | 'rare' | 'mythic';
  showTooltip?: boolean;
}

// Cache for set info to avoid repeated API calls
const setInfoCache = new Map<string, gui.SetInfo | null>();

// Export for testing purposes
// eslint-disable-next-line react-refresh/only-export-components
export function clearSetInfoCache(): void {
  setInfoCache.clear();
}

const SetSymbol = memo(({ setCode, size = 'medium', rarity, showTooltip = true }: SetSymbolProps) => {
  const [setInfo, setSetInfo] = useState<gui.SetInfo | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(false);

  useEffect(() => {
    const fetchSetInfo = async () => {
      // Check cache first
      if (setInfoCache.has(setCode)) {
        setSetInfo(setInfoCache.get(setCode) || null);
        setLoading(false);
        return;
      }

      try {
        setLoading(true);
        const info = await getSetInfo(setCode);
        if (info) {
          const setInfoObj = gui.SetInfo.createFrom(info);
          setInfoCache.set(setCode, setInfoObj);
          setSetInfo(setInfoObj);
        } else {
          setInfoCache.set(setCode, null);
          setSetInfo(null);
        }
      } catch (err) {
        console.error(`Failed to fetch set info for ${setCode}:`, err);
        setError(true);
      } finally {
        setLoading(false);
      }
    };

    if (setCode) {
      fetchSetInfo();
    }
  }, [setCode]);

  // Get size in pixels
  const sizeMap = {
    small: 16,
    medium: 20,
    large: 24,
  };
  const pixelSize = sizeMap[size];

  // Rarity color classes
  const rarityClass = rarity ? `set-symbol-${rarity}` : '';

  if (loading) {
    return (
      <span
        className={`set-symbol set-symbol-loading ${rarityClass}`}
        style={{ width: pixelSize, height: pixelSize }}
        title={showTooltip ? `Loading ${setCode.toUpperCase()}...` : undefined}
      >
        {setCode.toUpperCase()}
      </span>
    );
  }

  if (error || !setInfo || !setInfo.iconSvgUri) {
    // Fallback to text display
    return (
      <span
        className={`set-symbol set-symbol-text ${rarityClass}`}
        title={showTooltip ? setInfo?.name || setCode.toUpperCase() : undefined}
      >
        {setCode.toUpperCase()}
      </span>
    );
  }

  return (
    <img
      src={setInfo.iconSvgUri}
      alt={setInfo.name || setCode}
      className={`set-symbol set-symbol-icon ${rarityClass}`}
      style={{ width: pixelSize, height: pixelSize }}
      title={showTooltip ? setInfo.name : undefined}
      onError={() => setError(true)}
    />
  );
});

SetSymbol.displayName = 'SetSymbol';

export default SetSymbol;
