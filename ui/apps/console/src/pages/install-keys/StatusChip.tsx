import { type ComponentType, type SVGProps } from "react";
import { cn } from "@shellhub/design-system/cn";

type ChipTone = "green" | "red" | "yellow" | "muted" | "primary";

const TONE: Record<ChipTone, string> = {
  green: "bg-accent-green/10 text-accent-green",
  red: "bg-accent-red/10 text-accent-red",
  yellow: "bg-accent-yellow/10 text-accent-yellow",
  muted: "bg-text-muted/10 text-text-muted",
  primary: "bg-primary/10 text-primary",
};

export default function StatusChip({
  icon: Icon,
  label,
  tone,
  mono = false,
}: {
  icon?: ComponentType<SVGProps<SVGSVGElement>>;
  label: string;
  tone: ChipTone;
  mono?: boolean;
}) {
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1.5 rounded-md px-2 py-0.5 text-2xs",
        mono && "font-mono",
        TONE[tone],
      )}
    >
      {Icon && <Icon className="w-3.5 h-3.5 shrink-0" strokeWidth={1.8} />}
      {label}
    </span>
  );
}

export function DeprecatedBadge() {
  return <StatusChip label="Deprecated" tone="yellow" />;
}
