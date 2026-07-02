import type { CSSProperties } from "react";
import { cn } from "../primitives/cn";

export type GlowOrbsPreset = "hero" | "duo" | "section" | "corner" | "ambient";

/** Solid-fill tones for the static presets (hero/duo/section/corner). */
export type GlowOrbsTone = "primary" | "cyan" | "blue" | "green" | "yellow";

/** Ambient recolors all three pulsing orbs together as a named set. */
export type GlowOrbsAmbientTone = "brand" | "error" | "warning";

export type GlowOrbsProps =
  | { preset: "hero"; tone?: GlowOrbsTone; className?: string }
  | {
      preset: "duo" | "section" | "corner";
      tone: GlowOrbsTone;
      className?: string;
    }
  | { preset: "ambient"; tone: GlowOrbsAmbientTone; className?: string };

type Layer = { className: string; style?: CSSProperties };

// Tailwind only emits classes it finds spelled out in source, so each tone's color utilities are
// listed as complete literals here; the geometry shared across tones lives in the layer builders.
const HERO: Record<GlowOrbsTone, { wash: string; lg: string; sm: string }> = {
  primary: { wash: "from-primary/10", lg: "bg-primary/8", sm: "bg-primary/5" },
  cyan: {
    wash: "from-accent-cyan/10",
    lg: "bg-accent-cyan/8",
    sm: "bg-accent-cyan/5",
  },
  blue: {
    wash: "from-accent-blue/10",
    lg: "bg-accent-blue/8",
    sm: "bg-accent-blue/5",
  },
  green: {
    wash: "from-accent-green/10",
    lg: "bg-accent-green/8",
    sm: "bg-accent-green/5",
  },
  yellow: {
    wash: "from-accent-yellow/10",
    lg: "bg-accent-yellow/8",
    sm: "bg-accent-yellow/5",
  },
};

const DUO: Record<GlowOrbsTone, { wash: string; orb: string }> = {
  primary: { wash: "from-primary/8", orb: "bg-primary/8" },
  cyan: { wash: "from-accent-cyan/8", orb: "bg-accent-cyan/8" },
  blue: { wash: "from-accent-blue/8", orb: "bg-accent-blue/8" },
  green: { wash: "from-accent-green/8", orb: "bg-accent-green/8" },
  yellow: { wash: "from-accent-yellow/8", orb: "bg-accent-yellow/8" },
};

const SECTION: Record<GlowOrbsTone, { wash: string; orb: string }> = {
  primary: { wash: "from-primary/8", orb: "bg-primary/6" },
  cyan: { wash: "from-accent-cyan/8", orb: "bg-accent-cyan/6" },
  blue: { wash: "from-accent-blue/8", orb: "bg-accent-blue/6" },
  green: { wash: "from-accent-green/8", orb: "bg-accent-green/6" },
  yellow: { wash: "from-accent-yellow/8", orb: "bg-accent-yellow/6" },
};

const CORNER: Record<GlowOrbsTone, { from: string; via: string; orb: string }> =
  {
    primary: {
      from: "from-primary/[0.08]",
      via: "via-primary/[0.02]",
      orb: "bg-primary/[0.08]",
    },
    cyan: {
      from: "from-accent-cyan/[0.08]",
      via: "via-accent-cyan/[0.02]",
      orb: "bg-accent-cyan/[0.08]",
    },
    blue: {
      from: "from-accent-blue/[0.08]",
      via: "via-accent-blue/[0.02]",
      orb: "bg-accent-blue/[0.08]",
    },
    green: {
      from: "from-accent-green/[0.08]",
      via: "via-accent-green/[0.02]",
      orb: "bg-accent-green/[0.08]",
    },
    yellow: {
      from: "from-accent-yellow/[0.08]",
      via: "via-accent-yellow/[0.02]",
      orb: "bg-accent-yellow/[0.08]",
    },
  };

// Ambient recolors all three orbs together; opacity descends with orb size.
const AMBIENT: Record<GlowOrbsAmbientTone, [string, string, string]> = {
  brand: ["bg-primary/10", "bg-accent-cyan/8", "bg-accent-blue/5"],
  error: ["bg-accent-red/[0.06]", "bg-primary/[0.04]", "bg-accent-red/[0.03]"],
  warning: ["bg-accent-yellow/10", "bg-primary/8", "bg-accent-blue/5"],
};

const WASH =
  "absolute inset-0 bg-gradient-radial via-transparent to-transparent";
const ORB = "absolute rounded-full blur-3xl";

// Ambient orbs use pixel blurs and stagger their pulse so the set doesn't breathe in lockstep.
const AMBIENT_ORBS = [
  {
    geometry: "-top-32 -left-32 w-[500px] h-[500px] blur-[120px]",
    delay: "0s",
  },
  {
    geometry: "-bottom-48 -right-32 w-[400px] h-[400px] blur-[100px]",
    delay: "1s",
  },
  {
    geometry: "top-1/3 right-1/4 w-[300px] h-[300px] blur-[80px]",
    delay: "2s",
  },
] as const;

function layersFor(props: GlowOrbsProps): Layer[] {
  switch (props.preset) {
    case "hero": {
      const c = HERO[props.tone ?? "primary"];
      return [
        { className: cn(WASH, c.wash) },
        { className: cn(ORB, "top-16 left-1/4 w-[500px] h-[500px]", c.lg) },
        { className: cn(ORB, "bottom-0 right-1/4 w-96 h-96 bg-accent-cyan/6") },
        { className: cn(ORB, "top-1/3 right-[10%] w-72 h-72", c.sm) },
      ];
    }
    case "duo": {
      const c = DUO[props.tone];
      return [
        { className: cn(WASH, c.wash) },
        { className: cn(ORB, "top-1/4 left-1/4 w-96 h-96", c.orb) },
        { className: cn(ORB, "bottom-0 right-1/4 w-72 h-72 bg-accent-cyan/6") },
      ];
    }
    case "section": {
      const c = SECTION[props.tone];
      return [
        { className: cn(WASH, c.wash) },
        { className: cn(ORB, "top-1/3 right-1/4 w-96 h-96", c.orb) },
      ];
    }
    case "corner": {
      const c = CORNER[props.tone];
      return [
        {
          className: cn(
            "absolute inset-0 bg-gradient-to-br to-transparent",
            c.from,
            c.via,
          ),
        },
        {
          className: cn(
            ORB,
            "top-0 right-0 w-40 h-40 -translate-y-1/2 translate-x-1/2",
            c.orb,
          ),
        },
      ];
    }
    case "ambient": {
      const colors = AMBIENT[props.tone];
      return AMBIENT_ORBS.map(({ geometry, delay }, i) => ({
        className: cn(
          "absolute rounded-full animate-pulse-subtle",
          geometry,
          colors[i],
        ),
        style: { animationDelay: delay },
      }));
    }
  }
}

/**
 * Decorative blurred-orb backdrop. Each preset owns a fixed orb count, size, blur, and position;
 * `tone` is the only per-site variable and resolves to design tokens (never a raw color). Purely
 * ornamental — renders an `aria-hidden`, `pointer-events-none` layer meant to sit inside a
 * `relative` container, alongside any `ConnectionGrid`/`.grid-bg` the consumer composes.
 */
export function GlowOrbs(props: GlowOrbsProps) {
  return (
    <div
      aria-hidden
      className={cn(
        "absolute inset-0 overflow-hidden pointer-events-none",
        props.className,
      )}
    >
      {layersFor(props).map((layer) => (
        <div
          key={layer.className}
          className={layer.className}
          style={layer.style}
        />
      ))}
    </div>
  );
}
