import { useId } from "react";
import { Card } from "@shellhub/design-system/primitives";
import CopyButton from "@/components/common/CopyButton";
import { LABEL } from "@/utils/styles";

/**
 * A value with a copy button beside it. The label names the card holding the value, so it can be
 * found by its label, and the value selects as a whole on click.
 */
export default function CopyableValue({
  label,
  value,
  showLabel,
}: {
  label: string;
  value: string;
  showLabel?: boolean;
}) {
  const labelId = useId();

  return (
    <div>
      <span id={labelId} className={LABEL}>
        {label}
      </span>
      <Card
        aria-labelledby={labelId}
        className="rounded-lg px-3.5 py-2.5 flex items-center gap-2"
      >
        <code className="flex-1 text-xs font-mono text-accent-cyan break-all select-all">
          {value}
        </code>
        <CopyButton text={value} size="md" showLabel={showLabel} />
      </Card>
    </div>
  );
}
