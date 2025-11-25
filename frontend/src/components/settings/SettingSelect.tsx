import './SettingSelect.css';

export interface SelectOption {
  value: string;
  label: string;
}

export interface SettingSelectProps {
  /** Label text for the select */
  label: string;
  /** Description text shown below the label */
  description?: string;
  /** Currently selected value */
  value: string;
  /** Callback when selection changes */
  onChange: (value: string) => void;
  /** Available options */
  options: SelectOption[];
  /** Whether the select is disabled */
  disabled?: boolean;
  /** Optional hint text */
  hint?: string;
}

const SettingSelect = ({
  label,
  description,
  value,
  onChange,
  options,
  disabled = false,
  hint,
}: SettingSelectProps) => {
  return (
    <div className="setting-item">
      <label className="setting-label">
        {label}
        {description && <span className="setting-description">{description}</span>}
      </label>
      <div className="setting-control">
        <select
          className="select-input"
          value={value}
          onChange={(e) => onChange(e.target.value)}
          disabled={disabled}
        >
          {options.map((option) => (
            <option key={option.value} value={option.value}>
              {option.label}
            </option>
          ))}
        </select>
        {hint && <span className="setting-hint">{hint}</span>}
      </div>
    </div>
  );
};

export default SettingSelect;
