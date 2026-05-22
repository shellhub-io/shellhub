import NumericInput from "@/components/common/fields/NumericInput";
import CheckboxField from "@/components/common/fields/CheckboxField";
import { MAX_NAMESPACES_ERROR, isMaxNamespacesValid } from "@/utils/validation";

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
      <CheckboxField
        id={`${idPrefix}-limit-enabled`}
        label="Set namespace creation limit"
        checked={limitEnabled}
        onChange={onLimitEnabledChange}
      />
      {limitEnabled && (
        <div className="ml-6 space-y-3 animate-fade-in">
          <CheckboxField
            id={`${idPrefix}-limit-disabled`}
            label="Disable namespace creation"
            checked={limitDisabled}
            onChange={onLimitDisabledChange}
          />
          {!limitDisabled && (
            <NumericInput
              id={`${idPrefix}-max-ns`}
              label="Max namespaces"
              value={maxNamespaces}
              onChange={onMaxNamespacesChange}
              error={valid ? undefined : MAX_NAMESPACES_ERROR}
            />
          )}
        </div>
      )}
    </div>
  );
}
