import CheckboxField from "@/components/common/fields/CheckboxField";
import StepperField from "@/components/common/fields/StepperField";
import OptionCard from "./OptionCard";

/**
 * The ephemeral toggle: whether devices enrolled with this key are removed when they go offline.
 */
export default function EphemeralField({
  id,
  enabled,
  onEnabledChange,
  timeout,
  onTimeoutChange,
}: {
  id: string;
  enabled: boolean;
  onEnabledChange: (value: boolean) => void;
  timeout: number;
  onTimeoutChange: (value: number) => void;
}) {
  return (
    <OptionCard active={enabled}>
      <CheckboxField
        id={id}
        label="Ephemeral devices"
        description="Automatically remove registered devices once they stay offline past the timeout."
        checked={enabled}
        onChange={onEnabledChange}
      />
      {enabled && (
        <div className="flex items-center gap-3 mt-4 pt-4 border-t border-border/70">
          <span className="text-xs text-text-secondary">Remove after</span>
          <StepperField
            id={`${id}-timeout`}
            label="Timeout in minutes"
            value={timeout}
            onChange={onTimeoutChange}
            min={1}
            max={10}
            readOnly
            size="sm"
            className="[&_input]:w-9"
          />
          <span className="text-2xs text-text-muted">min offline</span>
        </div>
      )}
    </OptionCard>
  );
}
