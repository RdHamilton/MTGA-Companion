import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import SetSymbol, { clearSetInfoCache } from './SetSymbol';
import { mockWailsApp } from '../test/mocks/wailsApp';
import { gui } from '../../wailsjs/go/models';

// Helper to create mock set info
function createMockSetInfo(overrides: Partial<gui.SetInfo> = {}): gui.SetInfo {
  return new gui.SetInfo({
    code: 'DSK',
    name: 'Duskmourn: House of Horror',
    iconSvgUri: 'https://svgs.scryfall.io/sets/dsk.svg',
    setType: 'expansion',
    releasedAt: '2024-09-27',
    cardCount: 277,
    ...overrides,
  });
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
      mockWailsApp.GetSetInfo.mockImplementation(() => new Promise(() => {}));

      render(<SetSymbol setCode="DSK" />);

      const loadingElement = document.querySelector('.set-symbol-loading');
      expect(loadingElement).toBeInTheDocument();
      expect(loadingElement?.textContent).toBe('DSK');
    });

    it('should show set code in uppercase during loading', () => {
      mockWailsApp.GetSetInfo.mockImplementation(() => new Promise(() => {}));

      render(<SetSymbol setCode="blb" />);

      const loadingElement = document.querySelector('.set-symbol-loading');
      expect(loadingElement?.textContent).toBe('BLB');
    });
  });

  describe('Success State', () => {
    it('should render set symbol image when data loads', async () => {
      const mockSetInfo = createMockSetInfo();
      mockWailsApp.GetSetInfo.mockResolvedValue(mockSetInfo);

      render(<SetSymbol setCode="DSK" />);

      await waitFor(() => {
        const img = screen.getByRole('img');
        expect(img).toBeInTheDocument();
        expect(img).toHaveAttribute('src', 'https://svgs.scryfall.io/sets/dsk.svg');
        expect(img).toHaveAttribute('alt', 'Duskmourn: House of Horror');
      });
    });

    it('should apply correct CSS class for icon', async () => {
      mockWailsApp.GetSetInfo.mockResolvedValue(createMockSetInfo());

      const { container } = render(<SetSymbol setCode="DSK" />);

      await waitFor(() => {
        const icon = container.querySelector('.set-symbol-icon');
        expect(icon).toBeInTheDocument();
      });
    });

    it('should set tooltip with set name by default', async () => {
      mockWailsApp.GetSetInfo.mockResolvedValue(createMockSetInfo());

      render(<SetSymbol setCode="DSK" />);

      await waitFor(() => {
        const img = screen.getByRole('img');
        expect(img).toHaveAttribute('title', 'Duskmourn: House of Horror');
      });
    });

    it('should not show tooltip when showTooltip is false', async () => {
      mockWailsApp.GetSetInfo.mockResolvedValue(createMockSetInfo());

      render(<SetSymbol setCode="DSK" showTooltip={false} />);

      await waitFor(() => {
        const img = screen.getByRole('img');
        expect(img).not.toHaveAttribute('title');
      });
    });
  });

  describe('Error State', () => {
    it('should show text fallback when API call fails', async () => {
      mockWailsApp.GetSetInfo.mockRejectedValue(new Error('Network error'));

      const { container } = render(<SetSymbol setCode="DSK" />);

      await waitFor(() => {
        const textFallback = container.querySelector('.set-symbol-text');
        expect(textFallback).toBeInTheDocument();
        expect(textFallback?.textContent).toBe('DSK');
      });
    });

    it('should show text fallback when set not found', async () => {
      mockWailsApp.GetSetInfo.mockResolvedValue(null);

      const { container } = render(<SetSymbol setCode="XXX" />);

      await waitFor(() => {
        const textFallback = container.querySelector('.set-symbol-text');
        expect(textFallback).toBeInTheDocument();
        expect(textFallback?.textContent).toBe('XXX');
      });
    });

    it('should show text fallback when iconSvgUri is empty', async () => {
      mockWailsApp.GetSetInfo.mockResolvedValue(
        createMockSetInfo({ iconSvgUri: '' })
      );

      const { container } = render(<SetSymbol setCode="DSK" />);

      await waitFor(() => {
        const textFallback = container.querySelector('.set-symbol-text');
        expect(textFallback).toBeInTheDocument();
      });
    });
  });

  describe('Size Variants', () => {
    it('should render small size (16px)', async () => {
      mockWailsApp.GetSetInfo.mockResolvedValue(createMockSetInfo());

      render(<SetSymbol setCode="DSK" size="small" />);

      await waitFor(() => {
        const img = screen.getByRole('img');
        expect(img).toHaveStyle({ width: '16px', height: '16px' });
      });
    });

    it('should render medium size by default (20px)', async () => {
      mockWailsApp.GetSetInfo.mockResolvedValue(createMockSetInfo());

      render(<SetSymbol setCode="DSK" />);

      await waitFor(() => {
        const img = screen.getByRole('img');
        expect(img).toHaveStyle({ width: '20px', height: '20px' });
      });
    });

    it('should render large size (24px)', async () => {
      mockWailsApp.GetSetInfo.mockResolvedValue(createMockSetInfo());

      render(<SetSymbol setCode="DSK" size="large" />);

      await waitFor(() => {
        const img = screen.getByRole('img');
        expect(img).toHaveStyle({ width: '24px', height: '24px' });
      });
    });
  });

  describe('Rarity Colors', () => {
    it('should apply common rarity class', async () => {
      mockWailsApp.GetSetInfo.mockResolvedValue(createMockSetInfo());

      const { container } = render(<SetSymbol setCode="DSK" rarity="common" />);

      await waitFor(() => {
        const icon = container.querySelector('.set-symbol-common');
        expect(icon).toBeInTheDocument();
      });
    });

    it('should apply uncommon rarity class', async () => {
      mockWailsApp.GetSetInfo.mockResolvedValue(createMockSetInfo());

      const { container } = render(<SetSymbol setCode="DSK" rarity="uncommon" />);

      await waitFor(() => {
        const icon = container.querySelector('.set-symbol-uncommon');
        expect(icon).toBeInTheDocument();
      });
    });

    it('should apply rare rarity class', async () => {
      mockWailsApp.GetSetInfo.mockResolvedValue(createMockSetInfo());

      const { container } = render(<SetSymbol setCode="DSK" rarity="rare" />);

      await waitFor(() => {
        const icon = container.querySelector('.set-symbol-rare');
        expect(icon).toBeInTheDocument();
      });
    });

    it('should apply mythic rarity class', async () => {
      mockWailsApp.GetSetInfo.mockResolvedValue(createMockSetInfo());

      const { container } = render(<SetSymbol setCode="DSK" rarity="mythic" />);

      await waitFor(() => {
        const icon = container.querySelector('.set-symbol-mythic');
        expect(icon).toBeInTheDocument();
      });
    });

    it('should not apply rarity class when rarity is not specified', async () => {
      mockWailsApp.GetSetInfo.mockResolvedValue(createMockSetInfo());

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
      mockWailsApp.GetSetInfo.mockResolvedValue(createMockSetInfo());

      const { rerender } = render(<SetSymbol setCode="DSK" />);

      await waitFor(() => {
        expect(screen.getByRole('img')).toBeInTheDocument();
      });

      // Rerender with same set code
      rerender(<SetSymbol setCode="DSK" />);

      // Should still show the image (cached)
      expect(screen.getByRole('img')).toBeInTheDocument();

      // GetSetInfo should only be called once due to caching
      expect(mockWailsApp.GetSetInfo).toHaveBeenCalledTimes(1);
    });

    it('should fetch new set when setCode changes', async () => {
      const dskInfo = createMockSetInfo({ code: 'DSK', name: 'Duskmourn' });
      const blbInfo = createMockSetInfo({
        code: 'BLB',
        name: 'Bloomburrow',
        iconSvgUri: 'https://svgs.scryfall.io/sets/blb.svg',
      });

      mockWailsApp.GetSetInfo.mockImplementation((setCode: string) => {
        if (setCode === 'DSK') return Promise.resolve(dskInfo);
        if (setCode === 'BLB') return Promise.resolve(blbInfo);
        return Promise.resolve(null);
      });

      const { rerender } = render(<SetSymbol setCode="DSK" />);

      await waitFor(() => {
        expect(screen.getByRole('img')).toHaveAttribute('alt', 'Duskmourn');
      });

      // Change to different set
      rerender(<SetSymbol setCode="BLB" />);

      await waitFor(() => {
        expect(screen.getByRole('img')).toHaveAttribute('alt', 'Bloomburrow');
      });
    });
  });

  describe('Image Error Handling', () => {
    it('should show text fallback when image fails to load', async () => {
      mockWailsApp.GetSetInfo.mockResolvedValue(createMockSetInfo());

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
    it('should call GetSetInfo with correct set code', async () => {
      mockWailsApp.GetSetInfo.mockResolvedValue(createMockSetInfo());

      render(<SetSymbol setCode="FDN" />);

      await waitFor(() => {
        expect(mockWailsApp.GetSetInfo).toHaveBeenCalledWith('FDN');
      });
    });

    it('should handle empty setCode gracefully', () => {
      render(<SetSymbol setCode="" />);

      // Should not call API with empty set code
      expect(mockWailsApp.GetSetInfo).not.toHaveBeenCalled();
    });
  });
});
