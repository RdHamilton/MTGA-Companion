import LoadingButton from '../../LoadingButton';
import { SettingItem, SettingSelect } from '../';

export interface SeventeenLandsSectionProps {
  setCode: string;
  onSetCodeChange: (code: string) => void;
  draftFormat: string;
  onDraftFormatChange: (format: string) => void;
  isFetchingRatings: boolean;
  isFetchingCards: boolean;
  isRecalculating: boolean;
  recalculateMessage: string;
  dataSource: string;
  isClearingCache: boolean;
  onFetchSetRatings: () => void;
  onRefreshSetRatings: () => void;
  onFetchSetCards: () => void;
  onRefreshSetCards: () => void;
  onRecalculateGrades: () => void;
  onClearDatasetCache: () => void;
}

export function SeventeenLandsSection({
  setCode,
  onSetCodeChange,
  draftFormat,
  onDraftFormatChange,
  isFetchingRatings,
  isFetchingCards,
  isRecalculating,
  recalculateMessage,
  dataSource,
  isClearingCache,
  onFetchSetRatings,
  onRefreshSetRatings,
  onFetchSetCards,
  onRefreshSetCards,
  onRecalculateGrades,
  onClearDatasetCache,
}: SeventeenLandsSectionProps) {
  return (
    <div className="settings-section">
      <h2 className="section-title">17Lands Card Ratings</h2>
      <div className="setting-description settings-section-description">
        Download card ratings and tier lists from 17Lands for draft assistance.
        Ratings are used in the Active Draft page to show pick quality and recommendations.
      </div>

      <div className="setting-item">
        <label className="setting-label">
          Set Code
          <span className="setting-description">
            The MTG set code (e.g., TLA = Avatar: The Last Airbender, BLB = Bloomburrow, DSK = Duskmourn)
          </span>
        </label>
        <div className="setting-control">
          <input
            type="text"
            value={setCode}
            onChange={(e) => onSetCodeChange(e.target.value.toUpperCase())}
            placeholder="e.g., TLA, BLB, DSK, FDN"
            className="text-input input-width-200"
            maxLength={5}
          />
        </div>
      </div>

      <SettingSelect
        label="Draft Format"
        description="Choose between Premier Draft (BO1) or Quick Draft ratings"
        value={draftFormat}
        onChange={onDraftFormatChange}
        options={[
          { value: 'PremierDraft', label: 'Premier Draft (BO1)' },
          { value: 'QuickDraft', label: 'Quick Draft' },
          { value: 'TradDraft', label: 'Traditional Draft (BO3)' },
        ]}
      />

      <SettingItem
        label="Fetch Ratings"
        description="Download and cache 17Lands ratings for the selected set and format"
      >
        <LoadingButton
          loading={isFetchingRatings}
          loadingText="Fetching..."
          onClick={onFetchSetRatings}
          disabled={!setCode}
          variant="primary"
          className="button-margin-right"
        >
          Fetch Ratings
        </LoadingButton>
        <button
          className="action-button"
          onClick={onRefreshSetRatings}
          disabled={isFetchingRatings || !setCode}
        >
          Refresh (Re-download)
        </button>
      </SettingItem>

      <SettingItem
        label="Fetch Card Data (Scryfall)"
        description="Download and cache card details (names, images, text) from Scryfall for the selected set"
      >
        <LoadingButton
          loading={isFetchingCards}
          loadingText="Fetching..."
          onClick={onFetchSetCards}
          disabled={!setCode}
          variant="primary"
          className="button-margin-right"
        >
          Fetch Card Data
        </LoadingButton>
        <button
          className="action-button"
          onClick={onRefreshSetCards}
          disabled={isFetchingCards || !setCode}
        >
          Refresh (Re-download)
        </button>
      </SettingItem>

      <SettingItem
        label="Recalculate Draft Grades"
        description="Update all draft grades and predictions with the latest 17Lands card ratings"
      >
        <LoadingButton
          loading={isRecalculating}
          loadingText="Recalculating..."
          onClick={onRecalculateGrades}
          variant="recalculate"
        >
          Recalculate All Drafts
        </LoadingButton>
        {recalculateMessage && (
          <div className={`recalculate-message ${recalculateMessage.startsWith('✓') ? 'success' : 'error'}`}>
            {recalculateMessage}
          </div>
        )}
      </SettingItem>

      <SettingItem
        label="Clear Dataset Cache"
        description="Remove cached 17Lands CSV files to free up disk space (ratings in database are preserved)"
      >
        <LoadingButton
          loading={isClearingCache}
          loadingText="Clearing..."
          onClick={onClearDatasetCache}
          variant="clear-cache"
        >
          Clear Dataset Cache
        </LoadingButton>
      </SettingItem>

      {dataSource && (
        <div className="setting-hint settings-success-box">
          <strong>📊 Current Data Source:</strong> {
            dataSource === 's3' ? 'S3 Public Datasets (recommended, uses locally cached game data)' :
            dataSource === 'web_api' ? 'Web API (fallback for new sets like TLA)' :
            dataSource === 'legacy_api' ? 'Legacy API (older mode)' :
            dataSource
          }
        </div>
      )}

      <div className="setting-hint settings-info-box">
        <strong>💡 Common Set Codes:</strong>
        <div className="set-codes-grid">
          <div>• TLA - Avatar: The Last Airbender</div>
          <div>• FDN - Foundations</div>
          <div>• DSK - Duskmourn</div>
          <div>• BLB - Bloomburrow</div>
          <div>• OTJ - Outlaws of Thunder Junction</div>
          <div>• MKM - Murders at Karlov Manor</div>
        </div>
      </div>
    </div>
  );
}
