import NumericInput from "@/components/common/NumericInput";
import { LABEL, INPUT } from "@/utils/styles";
import { isMaxNamespacesValid } from "@/utils/validation";

interface NamespaceLimitFieldsProps {
  idPrefix: string;
  limitEnabled: boolean;
  onLimitEnabledChange: (v: boolean) => void;
  limitDisabled: boolean;
  onLimitDisabledChange: (v: boolean) => void;
  maxNamespaces: string;
  onMaxNamespacesChange: (v: string) => void;
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
  const valid = isMaxNamespacesValid(
    limitEnabled,
    limitDisabled,
    maxNamespaces,
  );

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
              <NumericInput
                id={`${idPrefix}-max-ns`}
                value={maxNamespaces}
                onChange={onMaxNamespacesChange}
                className={`${INPUT} w-32`}
              />
              {!valid && (
                <p className="text-2xs text-accent-red mt-1.5">
                  Max namespaces must be a number greater than or equal to 1
                </p>
              )}
            </div>
          )}
        </div>
      )}
    </div>
  );
}
