import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { SeventeenLandsSection } from './SeventeenLandsSection';

describe('SeventeenLandsSection', () => {
  const defaultProps = {
    setCode: '',
    onSetCodeChange: vi.fn(),
    draftFormat: 'PremierDraft',
    onDraftFormatChange: vi.fn(),
    isFetchingRatings: false,
    isFetchingCards: false,
    isRecalculating: false,
    recalculateMessage: '',
    dataSource: '',
    isClearingCache: false,
    onFetchSetRatings: vi.fn(),
    onRefreshSetRatings: vi.fn(),
    onFetchSetCards: vi.fn(),
    onRefreshSetCards: vi.fn(),
    onRecalculateGrades: vi.fn(),
    onClearDatasetCache: vi.fn(),
  };

  it('renders section title', () => {
    render(<SeventeenLandsSection {...defaultProps} />);
    expect(screen.getByText('17Lands Card Ratings')).toBeInTheDocument();
  });

  describe('set code input', () => {
    it('renders set code input', () => {
      render(<SeventeenLandsSection {...defaultProps} />);
      expect(screen.getByPlaceholderText(/TLA, BLB/)).toBeInTheDocument();
    });

    it('displays current set code', () => {
      render(<SeventeenLandsSection {...defaultProps} setCode="BLB" />);
      expect(screen.getByDisplayValue('BLB')).toBeInTheDocument();
    });

    it('calls onSetCodeChange when input changes', () => {
      const onSetCodeChange = vi.fn();
      render(<SeventeenLandsSection {...defaultProps} onSetCodeChange={onSetCodeChange} />);

      const input = screen.getByPlaceholderText(/TLA, BLB/);
      fireEvent.change(input, { target: { value: 'dsk' } });

      expect(onSetCodeChange).toHaveBeenCalledWith('DSK');
    });
  });

  describe('draft format selector', () => {
    it('renders draft format selector', () => {
      render(<SeventeenLandsSection {...defaultProps} />);
      expect(screen.getByText('Draft Format')).toBeInTheDocument();
    });

    it('calls onDraftFormatChange when format changes', () => {
      const onDraftFormatChange = vi.fn();
      render(<SeventeenLandsSection {...defaultProps} onDraftFormatChange={onDraftFormatChange} />);

      const select = screen.getByRole('combobox');
      fireEvent.change(select, { target: { value: 'QuickDraft' } });

      expect(onDraftFormatChange).toHaveBeenCalledWith('QuickDraft');
    });
  });

  describe('fetch ratings', () => {
    it('renders fetch ratings button', () => {
      render(<SeventeenLandsSection {...defaultProps} />);
      expect(screen.getByRole('button', { name: 'Fetch Ratings' })).toBeInTheDocument();
    });

    it('disables fetch button when no set code', () => {
      render(<SeventeenLandsSection {...defaultProps} setCode="" />);
      expect(screen.getByRole('button', { name: 'Fetch Ratings' })).toBeDisabled();
    });

    it('enables fetch button when set code provided', () => {
      render(<SeventeenLandsSection {...defaultProps} setCode="BLB" />);
      expect(screen.getByRole('button', { name: 'Fetch Ratings' })).not.toBeDisabled();
    });

    it('calls onFetchSetRatings when clicked', () => {
      const onFetchSetRatings = vi.fn();
      render(<SeventeenLandsSection {...defaultProps} setCode="BLB" onFetchSetRatings={onFetchSetRatings} />);

      fireEvent.click(screen.getByRole('button', { name: 'Fetch Ratings' }));

      expect(onFetchSetRatings).toHaveBeenCalled();
    });

    it('shows loading state when fetching ratings', () => {
      render(<SeventeenLandsSection {...defaultProps} setCode="BLB" isFetchingRatings={true} />);
      expect(screen.getByRole('button', { name: 'Fetching...' })).toBeInTheDocument();
    });

    it('renders refresh button', () => {
      render(<SeventeenLandsSection {...defaultProps} />);
      // There are two refresh buttons, get all of them
      const refreshButtons = screen.getAllByRole('button', { name: 'Refresh (Re-download)' });
      expect(refreshButtons.length).toBe(2);
    });

    it('calls onRefreshSetRatings when refresh clicked', () => {
      const onRefreshSetRatings = vi.fn();
      render(<SeventeenLandsSection {...defaultProps} setCode="BLB" onRefreshSetRatings={onRefreshSetRatings} />);

      // First refresh button is for ratings
      const refreshButtons = screen.getAllByRole('button', { name: 'Refresh (Re-download)' });
      fireEvent.click(refreshButtons[0]);

      expect(onRefreshSetRatings).toHaveBeenCalled();
    });
  });

  describe('fetch cards', () => {
    it('renders fetch card data button', () => {
      render(<SeventeenLandsSection {...defaultProps} />);
      expect(screen.getByRole('button', { name: 'Fetch Card Data' })).toBeInTheDocument();
    });

    it('calls onFetchSetCards when clicked', () => {
      const onFetchSetCards = vi.fn();
      render(<SeventeenLandsSection {...defaultProps} setCode="BLB" onFetchSetCards={onFetchSetCards} />);

      fireEvent.click(screen.getByRole('button', { name: 'Fetch Card Data' }));

      expect(onFetchSetCards).toHaveBeenCalled();
    });

    it('shows loading state when fetching cards', () => {
      render(<SeventeenLandsSection {...defaultProps} setCode="BLB" isFetchingCards={true} />);
      // Get all Fetching... buttons
      const fetchingButtons = screen.getAllByRole('button', { name: 'Fetching...' });
      expect(fetchingButtons.length).toBeGreaterThanOrEqual(1);
    });
  });

  describe('recalculate grades', () => {
    it('renders recalculate button', () => {
      render(<SeventeenLandsSection {...defaultProps} />);
      expect(screen.getByRole('button', { name: 'Recalculate All Drafts' })).toBeInTheDocument();
    });

    it('calls onRecalculateGrades when clicked', () => {
      const onRecalculateGrades = vi.fn();
      render(<SeventeenLandsSection {...defaultProps} onRecalculateGrades={onRecalculateGrades} />);

      fireEvent.click(screen.getByRole('button', { name: 'Recalculate All Drafts' }));

      expect(onRecalculateGrades).toHaveBeenCalled();
    });

    it('shows loading state when recalculating', () => {
      render(<SeventeenLandsSection {...defaultProps} isRecalculating={true} />);
      expect(screen.getByRole('button', { name: 'Recalculating...' })).toBeInTheDocument();
    });

    it('shows success message', () => {
      render(<SeventeenLandsSection {...defaultProps} recalculateMessage="✓ Success!" />);
      expect(screen.getByText('✓ Success!')).toBeInTheDocument();
    });

    it('shows error message', () => {
      render(<SeventeenLandsSection {...defaultProps} recalculateMessage="✗ Failed" />);
      expect(screen.getByText('✗ Failed')).toBeInTheDocument();
    });
  });

  describe('clear cache', () => {
    it('renders clear cache button', () => {
      render(<SeventeenLandsSection {...defaultProps} />);
      expect(screen.getByRole('button', { name: 'Clear Dataset Cache' })).toBeInTheDocument();
    });

    it('calls onClearDatasetCache when clicked', () => {
      const onClearDatasetCache = vi.fn();
      render(<SeventeenLandsSection {...defaultProps} onClearDatasetCache={onClearDatasetCache} />);

      fireEvent.click(screen.getByRole('button', { name: 'Clear Dataset Cache' }));

      expect(onClearDatasetCache).toHaveBeenCalled();
    });

    it('shows loading state when clearing', () => {
      render(<SeventeenLandsSection {...defaultProps} isClearingCache={true} />);
      expect(screen.getByRole('button', { name: 'Clearing...' })).toBeInTheDocument();
    });
  });

  describe('data source display', () => {
    it('does not show data source when empty', () => {
      render(<SeventeenLandsSection {...defaultProps} dataSource="" />);
      expect(screen.queryByText(/Current Data Source/)).not.toBeInTheDocument();
    });

    it('shows s3 data source', () => {
      render(<SeventeenLandsSection {...defaultProps} dataSource="s3" />);
      expect(screen.getByText(/S3 Public Datasets/)).toBeInTheDocument();
    });

    it('shows web_api data source', () => {
      render(<SeventeenLandsSection {...defaultProps} dataSource="web_api" />);
      expect(screen.getByText(/Web API/)).toBeInTheDocument();
    });
  });

  describe('common set codes', () => {
    it('shows common set codes info', () => {
      render(<SeventeenLandsSection {...defaultProps} />);
      expect(screen.getByText(/Common Set Codes/)).toBeInTheDocument();
      expect(screen.getByText(/TLA - Avatar/)).toBeInTheDocument();
      expect(screen.getByText(/BLB - Bloomburrow/)).toBeInTheDocument();
    });
  });
});
