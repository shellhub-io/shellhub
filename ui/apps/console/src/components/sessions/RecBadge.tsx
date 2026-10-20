import { cn } from "@shellhub/design-system/cn";
import StatePill from "@/components/common/StatePill";

/**
 * The REC mark: a pill whose dot is filled and pulsing while recording and an empty ring while
 * not. With onToggle it is the switch that turns recording on and off, named by label; without it
 * it only states that recording is on or off.
 */
export default function RecBadge({
  on,
  onToggle,
  label,
  describedBy,
  disabled,
}: {
  on: boolean;
  onToggle?: () => void;
  label?: string;
  describedBy?: string;
  disabled?: boolean;
}) {
  return (
    <StatePill
      on={on}
      onClassName="border-accent-red/40 bg-accent-red/10 text-accent-red"
      onToggle={onToggle}
      label={label}
      describedBy={describedBy}
      disabled={disabled}
    >
      <span
        aria-hidden="true"
        className={cn(
          "w-3 h-3 rounded-full border-2 transition-all",
          on
            ? "border-accent-red bg-accent-red animate-pulse-subtle"
            : "border-current bg-transparent",
        )}
      />
      <span className="font-mono text-[10px] font-bold tracking-wider">
        REC
      </span>
    </StatePill>
  );
}
