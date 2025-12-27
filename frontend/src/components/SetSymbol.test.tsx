import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import SetSymbol, { clearSetInfoCache } from './SetSymbol';
import { mockCards } from '@/test/mocks/apiMock';
import { gui } from '@/types/models';

// Helper to create mock set info
function createMockSetInfo(overrides: Partial<gui.SetInfo> = {}): gui.SetInfo {
  return Object.assign(new gui.SetInfo({}), {
    code: 'DSK',
    name: 'Duskmourn: House of Horror',
    iconSvgUri: 'https://svgs.scryfall.io/sets/dsk.svg',
    setType: 'expansion',
    releasedAt: '2024-09-27',
    cardCount: 277,
    ...overrides,
  });
}

// Helper to create a list of sets
function createMockSetList(sets: Array<Partial<gui.SetInfo>>): gui.SetInfo[] {
  return sets.map((s) => createMockSetInfo(s));
}

describe('SetSymbol Component', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    // Clear the set info cache before each test to ensure isolation
    clearSetInfoCache();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe('Loading State', () => {
    it('should show loading state initially', () => {
      // Don't resolve the promise immediately
      mockCards.getAllSetInfo.mockImplementation(() => new Promise(() => {}));

      render(<SetSymbol setCode="DSK" />);

      const loadingElement = document.querySelector('.set-symbol-loading');
      expect(loadingElement).toBeInTheDocument();
      expect(loadingElement?.textContent).toBe('DSK');
    });

    it('should show set code in uppercase during loading', () => {
      mockCards.getAllSetInfo.mockImplementation(() => new Promise(() => {}));

      render(<SetSymbol setCode="blb" />);

      const loadingElement = document.querySelector('.set-symbol-loading');
      expect(loadingElement?.textContent).toBe('BLB');
    });
  });

  describe('Success State', () => {
    it('should render set symbol image when data loads', async () => {
      const mockSetInfo = createMockSetInfo();
      mockCards.getAllSetInfo.mockResolvedValue([mockSetInfo]);

      render(<SetSymbol setCode="DSK" />);

      await waitFor(() => {
        const img = screen.getByRole('img');
        expect(img).toBeInTheDocument();
        expect(img).toHaveAttribute('src', 'https://svgs.scryfall.io/sets/dsk.svg');
        expect(img).toHaveAttribute('alt', 'Duskmourn: House of Horror');
      });
    });

    it('should apply correct CSS class for icon', async () => {
      mockCards.getAllSetInfo.mockResolvedValue([createMockSetInfo()]);

      const { container } = render(<SetSymbol setCode="DSK" />);

      await waitFor(() => {
        const icon = container.querySelector('.set-symbol-icon');
        expect(icon).toBeInTheDocument();
      });
    });

    it('should set tooltip with set name by default', async () => {
      mockCards.getAllSetInfo.mockResolvedValue([createMockSetInfo()]);

      render(<SetSymbol setCode="DSK" />);

      await waitFor(() => {
        const img = screen.getByRole('img');
        expect(img).toHaveAttribute('title', 'Duskmourn: House of Horror');
      });
    });

    it('should not show tooltip when showTooltip is false', async () => {
      mockCards.getAllSetInfo.mockResolvedValue([createMockSetInfo()]);

      render(<SetSymbol setCode="DSK" showTooltip={false} />);

      await waitFor(() => {
        const img = screen.getByRole('img');
        expect(img).not.toHaveAttribute('title');
      });
    });
  });

  describe('Error State', () => {
    it('should show text fallback when API call fails', async () => {
      mockCards.getAllSetInfo.mockRejectedValue(new Error('Network error'));

      const { container } = render(<SetSymbol setCode="DSK" />);

      await waitFor(() => {
        const textFallback = container.querySelector('.set-symbol-text');
        expect(textFallback).toBeInTheDocument();
        expect(textFallback?.textContent).toBe('DSK');
      });
    });

    it('should show text fallback when set not found', async () => {
      // Return list without the requested set
      mockCards.getAllSetInfo.mockResolvedValue([createMockSetInfo({ code: 'OTHER' })]);

      const { container } = render(<SetSymbol setCode="XXX" />);

      await waitFor(() => {
        const textFallback = container.querySelector('.set-symbol-text');
        expect(textFallback).toBeInTheDocument();
        expect(textFallback?.textContent).toBe('XXX');
      });
    });

    it('should show text fallback when iconSvgUri is empty', async () => {
      mockCards.getAllSetInfo.mockResolvedValue([createMockSetInfo({ iconSvgUri: '' })]);

      const { container } = render(<SetSymbol setCode="DSK" />);

      await waitFor(() => {
        const textFallback = container.querySelector('.set-symbol-text');
        expect(textFallback).toBeInTheDocument();
      });
    });
  });

  describe('Size Variants', () => {
    it('should render small size (16px)', async () => {
      mockCards.getAllSetInfo.mockResolvedValue([createMockSetInfo()]);

      render(<SetSymbol setCode="DSK" size="small" />);

      await waitFor(() => {
        const img = screen.getByRole('img');
        expect(img).toHaveStyle({ width: '16px', height: '16px' });
      });
    });

    it('should render medium size by default (20px)', async () => {
      mockCards.getAllSetInfo.mockResolvedValue([createMockSetInfo()]);

      render(<SetSymbol setCode="DSK" />);

      await waitFor(() => {
        const img = screen.getByRole('img');
        expect(img).toHaveStyle({ width: '20px', height: '20px' });
      });
    });

    it('should render large size (24px)', async () => {
      mockCards.getAllSetInfo.mockResolvedValue([createMockSetInfo()]);

      render(<SetSymbol setCode="DSK" size="large" />);

      await waitFor(() => {
        const img = screen.getByRole('img');
        expect(img).toHaveStyle({ width: '24px', height: '24px' });
      });
    });
  });

  describe('Rarity Colors', () => {
    it('should apply common rarity class', async () => {
      mockCards.getAllSetInfo.mockResolvedValue([createMockSetInfo()]);

      const { container } = render(<SetSymbol setCode="DSK" rarity="common" />);

      await waitFor(() => {
        const icon = container.querySelector('.set-symbol-common');
        expect(icon).toBeInTheDocument();
      });
    });

    it('should apply uncommon rarity class', async () => {
      mockCards.getAllSetInfo.mockResolvedValue([createMockSetInfo()]);

      const { container } = render(<SetSymbol setCode="DSK" rarity="uncommon" />);

      await waitFor(() => {
        const icon = container.querySelector('.set-symbol-uncommon');
        expect(icon).toBeInTheDocument();
      });
    });

    it('should apply rare rarity class', async () => {
      mockCards.getAllSetInfo.mockResolvedValue([createMockSetInfo()]);

      const { container } = render(<SetSymbol setCode="DSK" rarity="rare" />);

      await waitFor(() => {
        const icon = container.querySelector('.set-symbol-rare');
        expect(icon).toBeInTheDocument();
      });
    });

    it('should apply mythic rarity class', async () => {
      mockCards.getAllSetInfo.mockResolvedValue([createMockSetInfo()]);

      const { container } = render(<SetSymbol setCode="DSK" rarity="mythic" />);

      await waitFor(() => {
        const icon = container.querySelector('.set-symbol-mythic');
        expect(icon).toBeInTheDocument();
      });
    });

    it('should not apply rarity class when rarity is not specified', async () => {
      mockCards.getAllSetInfo.mockResolvedValue([createMockSetInfo()]);

      const { container } = render(<SetSymbol setCode="DSK" />);

      await waitFor(() => {
        const icon = container.querySelector('.set-symbol-icon');
        expect(icon).not.toHaveClass('set-symbol-common');
        expect(icon).not.toHaveClass('set-symbol-uncommon');
        expect(icon).not.toHaveClass('set-symbol-rare');
        expect(icon).not.toHaveClass('set-symbol-mythic');
      });
    });
  });

  describe('Caching', () => {
    it('should cache set info and not call API twice for same set', async () => {
      mockCards.getAllSetInfo.mockResolvedValue([createMockSetInfo()]);

      const { rerender } = render(<SetSymbol setCode="DSK" />);

      await waitFor(() => {
        expect(screen.getByRole('img')).toBeInTheDocument();
      });

      // Rerender with same set code
      rerender(<SetSymbol setCode="DSK" />);

      // Should still show the image (cached)
      expect(screen.getByRole('img')).toBeInTheDocument();

      // getAllSetInfo should only be called once due to caching
      expect(mockCards.getAllSetInfo).toHaveBeenCalledTimes(1);
    });

    it('should fetch new set when setCode changes', async () => {
      const allSets = createMockSetList([
        { code: 'DSK', name: 'Duskmourn' },
        { code: 'BLB', name: 'Bloomburrow', iconSvgUri: 'https://svgs.scryfall.io/sets/blb.svg' },
      ]);
      mockCards.getAllSetInfo.mockResolvedValue(allSets);

      const { rerender } = render(<SetSymbol setCode="DSK" />);

      await waitFor(() => {
        expect(screen.getByRole('img')).toHaveAttribute('alt', 'Duskmourn');
      });

      // Clear cache and change to different set
      clearSetInfoCache();
      rerender(<SetSymbol setCode="BLB" />);

      await waitFor(() => {
        expect(screen.getByRole('img')).toHaveAttribute('alt', 'Bloomburrow');
      });
    });
  });

  describe('Image Error Handling', () => {
    it('should show text fallback when image fails to load', async () => {
      mockCards.getAllSetInfo.mockResolvedValue([createMockSetInfo()]);

      const { container } = render(<SetSymbol setCode="DSK" />);

      await waitFor(() => {
        const img = screen.getByRole('img');
        expect(img).toBeInTheDocument();
      });

      // Simulate image load error
      const img = screen.getByRole('img');
      img.dispatchEvent(new Event('error'));

      await waitFor(() => {
        const textFallback = container.querySelector('.set-symbol-text');
        expect(textFallback).toBeInTheDocument();
      });
    });
  });

  describe('API Integration', () => {
    it('should call getAllSetInfo to fetch set data', async () => {
      mockCards.getAllSetInfo.mockResolvedValue([createMockSetInfo({ code: 'FDN' })]);

      render(<SetSymbol setCode="FDN" />);

      await waitFor(() => {
        expect(mockCards.getAllSetInfo).toHaveBeenCalled();
      });
    });

    it('should handle empty setCode gracefully', () => {
      render(<SetSymbol setCode="" />);

      // Should not call API with empty set code
      expect(mockCards.getAllSetInfo).not.toHaveBeenCalled();
    });
  });
});
