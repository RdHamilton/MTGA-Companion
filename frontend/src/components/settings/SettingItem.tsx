import './SettingItem.css';

export interface SettingItemProps {
  /** Label text for the setting */
  label: string;
  /** Description text shown below the label */
  description?: string;
  /** Optional hint text (e.g., for showing URL preview) */
  hint?: string;
  /** Whether this setting is indented (for nested settings) */
  indented?: boolean;
  /** Whether this is a danger/destructive setting */
  danger?: boolean;
  /** The control element(s) for this setting */
  children: React.ReactNode;
}

const SettingItem = ({
  label,
  description,
  hint,
  indented = false,
  danger = false,
  children,
}: SettingItemProps) => {
  const classNames = ['setting-item'];
  if (indented) classNames.push('indented');
  if (danger) classNames.push('danger');

  return (
    <div className={classNames.join(' ')}>
      <label className="setting-label">
        {label}
        {description && <span className="setting-description">{description}</span>}
      </label>
      <div className="setting-control">
        {children}
        {hint && <span className="setting-hint">{hint}</span>}
      </div>
    </div>
  );
};

export default SettingItem;
