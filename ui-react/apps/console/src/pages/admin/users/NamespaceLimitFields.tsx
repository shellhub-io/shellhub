import { LABEL, INPUT } from "@/utils/styles";

interface NamespaceLimitFieldsProps {
  idPrefix: string;
  limitEnabled: boolean;
  onLimitEnabledChange: (v: boolean) => void;
  limitDisabled: boolean;
  onLimitDisabledChange: (v: boolean) => void;
  maxNamespaces: number;
  onMaxNamespacesChange: (v: number) => void;
}

export default function NamespaceLimitFields({
  idPrefix,
  limitEnabled,
  onLimitEnabledChange,
  limitDisabled,
  onLimitDisabledChange,
  maxNamespaces,
  onMaxNamespacesChange,
}: NamespaceLimitFieldsProps) {
  return (
    <div className="space-y-3">
      <label className="flex items-center gap-2 cursor-pointer">
        <input
          type="checkbox"
          checked={limitEnabled}
          onChange={(e) => onLimitEnabledChange(e.target.checked)}
          className="w-4 h-4 rounded border-border bg-card text-primary focus:ring-primary/20"
        />
        <span className="text-sm text-text-primary">
          Set namespace creation limit
        </span>
      </label>
      {limitEnabled && (
        <div className="ml-6 space-y-3 animate-fade-in">
          <label className="flex items-center gap-2 cursor-pointer">
            <input
              type="checkbox"
              checked={limitDisabled}
              onChange={(e) => onLimitDisabledChange(e.target.checked)}
              className="w-4 h-4 rounded border-border bg-card text-primary focus:ring-primary/20"
            />
            <span className="text-sm text-text-secondary">
              Disable namespace creation
            </span>
          </label>
          {!limitDisabled && (
            <div>
              <label className={LABEL} htmlFor={`${idPrefix}-max-ns`}>
                Max namespaces
              </label>
              <input
                id={`${idPrefix}-max-ns`}
                type="number"
                min={1}
                value={maxNamespaces}
                onChange={(e) =>
                  onMaxNamespacesChange(parseInt(e.target.value, 10) || 1)}
                className={`${INPUT} w-32`}
              />
            </div>
          )}
        </div>
      )}
    </div>
  );
}
