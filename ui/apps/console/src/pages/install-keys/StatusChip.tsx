import { type ComponentType, type SVGProps } from "react";

type ChipTone = "green" | "red" | "yellow" | "muted" | "primary";

/** Soft, pill-less chip: a tinted fill in the tone's hue with the label in that hue. */
const TONE: Record<ChipTone, string> = {
  green: "bg-accent-green/10 text-accent-green",
  red: "bg-accent-red/10 text-accent-red",
  yellow: "bg-accent-yellow/10 text-accent-yellow",
  muted: "bg-text-muted/10 text-text-muted",
  primary: "bg-primary/10 text-primary",
};

/**
 * The install-key feature's chip: an optional icon + label on a soft tinted fill. Quieter than a
 * bordered/uppercase pill, more finished than bare text. One shape for every marker in the feature —
 * the Registration kind, the Review verdict, the Deprecated flag, and tags (`mono`) — so they read as
 * one family. Colour carries the meaning.
 */
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
      className={`inline-flex items-center gap-1.5 rounded-md px-2 py-0.5 text-2xs ${
        mono ? "font-mono" : ""
      } ${TONE[tone]}`}
    >
      {Icon && <Icon className="w-3.5 h-3.5 shrink-0" strokeWidth={1.8} />}
      {label}
    </span>
  );
}
