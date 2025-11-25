import './SettingToggle.css';

export interface SettingToggleProps {
  /** Label text for the toggle */
  label: string;
  /** Description text shown below the label */
  description?: string;
  /** Whether the toggle is checked */
  checked: boolean;
  /** Callback when toggle state changes */
  onChange: (checked: boolean) => void;
  /** Whether the toggle is disabled */
  disabled?: boolean;
  /** Unique ID for the input (auto-generated if not provided) */
  id?: string;
}

let toggleIdCounter = 0;

const SettingToggle = ({
  label,
  description,
  checked,
  onChange,
  disabled = false,
  id,
}: SettingToggleProps) => {
  const inputId = id || `setting-toggle-${++toggleIdCounter}`;

  return (
    <div className="setting-item">
      <label className="setting-label" htmlFor={inputId}>
        <input
          type="checkbox"
          id={inputId}
          checked={checked}
          onChange={(e) => onChange(e.target.checked)}
          className="checkbox-input"
          disabled={disabled}
        />
        {label}
        {description && <span className="setting-description">{description}</span>}
      </label>
    </div>
  );
};

export default SettingToggle;
