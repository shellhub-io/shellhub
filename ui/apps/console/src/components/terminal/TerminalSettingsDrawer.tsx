import { useId, useState } from "react";
import {
  CheckIcon,
  ChevronDownIcon,
  MinusIcon,
  PlusIcon,
} from "@heroicons/react/24/outline";
import { cn } from "@shellhub/design-system/cn";
import { Dropdown, IconButton } from "@shellhub/design-system/primitives";
import {
  useTerminalThemeStore,
  TERMINAL_FONTS,
  MIN_FONT_SIZE,
  MAX_FONT_SIZE,
  type TerminalFont,
  type TerminalTheme,
} from "@/stores/terminalThemeStore";
import {
  PLAYER_CONTROLS,
  useSessionPlayerStore,
  type PlayerControls,
} from "@/stores/sessionPlayerStore";
import RadioGroupField from "../common/fields/RadioGroupField";
import RadioTile from "../common/fields/RadioTile";
import Drawer from "../common/Drawer";

interface Props {
  open: boolean;
  onClose: () => void;
}

function isLightTheme(bg: string): boolean {
  const hex = bg.replace("#", "");
  if (hex.length !== 6) return false;
  const r = parseInt(hex.slice(0, 2), 16);
  const g = parseInt(hex.slice(2, 4), 16);
  const b = parseInt(hex.slice(4, 6), 16);
  return (0.299 * r + 0.587 * g + 0.114 * b) / 255 > 0.5;
}

/**
 * The terminal's appearance settings, and how the session player shows its controls. They apply
 * to every open terminal and player at once, as a preference rather than per-session state.
 */
export default function TerminalSettingsDrawer({ open, onClose }: Props) {
  const {
    themes,
    themeName,
    fontFamily,
    fontSize,
    setTheme,
    setFontFamily,
    setFontSize,
  } = useTerminalThemeStore();
  const playerControls = useSessionPlayerStore((s) => s.controls);
  const setPlayerControls = useSessionPlayerStore((s) => s.setControls);
  const playerControlsHeading = useId();

  return (
    <Drawer
      open={open}
      onClose={onClose}
      title="Terminal Settings"
      width="sm"
      bodyClassName="flex-1 overflow-y-auto"
    >
      <div className="border-b border-border p-4">
        <div className="mb-2.5 text-2xs font-mono font-semibold uppercase tracking-label text-text-muted">
          Theme
        </div>
        <div className="grid grid-cols-2 gap-1.5">
          {themes.map((t) => (
            <ThemeCard
              key={t.name}
              theme={t}
              selected={t.name === themeName}
              onClick={() => setTheme(t.name)}
            />
          ))}
        </div>
      </div>

      <div className="border-b border-border p-4 flex items-center justify-between gap-3">
        <div className="text-2xs font-mono font-semibold uppercase tracking-label text-text-muted">
          Font Family
        </div>
        <FontPicker value={fontFamily} onChange={setFontFamily} />
      </div>

      <div className="border-b border-border p-4">
        <div className="mb-2.5 text-2xs font-mono font-semibold uppercase tracking-label text-text-muted">
          Font Size
        </div>
        <div className="flex items-center gap-3">
          <IconButton
            aria-label="Decrease font size"
            disabled={fontSize <= MIN_FONT_SIZE}
            onClick={() => setFontSize(fontSize - 1)}
            className="border border-border bg-hover-subtle"
          >
            <MinusIcon className="w-3.5 h-3.5" strokeWidth={2} />
          </IconButton>
          <div className="flex-1">
            <input
              type="range"
              min={MIN_FONT_SIZE}
              max={MAX_FONT_SIZE}
              value={fontSize}
              onChange={(e) => setFontSize(parseInt(e.target.value))}
              className="w-full accent-primary h-1"
            />
          </div>
          <IconButton
            aria-label="Increase font size"
            disabled={fontSize >= MAX_FONT_SIZE}
            onClick={() => setFontSize(fontSize + 1)}
            className="border border-border bg-hover-subtle"
          >
            <PlusIcon className="w-3.5 h-3.5" strokeWidth={2} />
          </IconButton>
          <span className="font-mono text-[13px] text-text-secondary w-7 text-right">
            {fontSize}
          </span>
        </div>
      </div>

      <div className="p-4">
        <div
          id={playerControlsHeading}
          className="mb-2.5 text-2xs font-mono font-semibold uppercase tracking-label text-text-muted"
        >
          Session Player Controls
        </div>
        <RadioGroupField
          labelledBy={playerControlsHeading}
          value={playerControls}
          onChange={setPlayerControls}
          containerClassName="grid grid-cols-3 gap-1.5"
        >
          {PLAYER_CONTROLS.map(({ value, label }) => (
            <RadioTile
              key={value}
              value={value}
              label={label}
              picture={<PlayerControlsSketch controls={value} />}
            />
          ))}
        </RadioGroupField>
      </div>
    </Drawer>
  );
}

function PlayerControlsSketch({ controls }: { controls: PlayerControls }) {
  return (
    <svg
      viewBox="0 0 64 32"
      fill="none"
      stroke="currentColor"
      strokeWidth={1}
      strokeLinecap="round"
      className="w-full h-auto"
      aria-hidden="true"
    >
      <rect x="1" y="1" width="62" height="30" rx="3.5" />
      <path d="M7 7h22M7 11h30M7 15h16" opacity={0.5} />
      {controls !== "hidden" && (
        <>
          <rect
            x="14"
            y="21"
            width="36"
            height="6"
            rx="3"
            strokeDasharray={controls === "auto" ? "2.5 2.5" : undefined}
          />
          <circle
            cx="18"
            cy="24"
            r="1.25"
            fill="currentColor"
            stroke="none"
            opacity={controls === "auto" ? 0.5 : 1}
          />
        </>
      )}
    </svg>
  );
}

function FontPicker({
  value,
  onChange,
}: {
  value: TerminalFont;
  onChange: (font: TerminalFont) => void;
}) {
  const [open, setOpen] = useState(false);

  return (
    <Dropdown placement="bottom-end" open={open} onOpenChange={setOpen}>
      <Dropdown.Trigger>
        <button
          type="button"
          aria-label={`Font family: ${value}`}
          className="flex items-center gap-1.5 px-2.5 py-1.5 text-sm text-text-primary bg-card border border-border rounded-md hover:border-border-light transition-colors"
          style={{ fontFamily: `"${value}", monospace` }}
        >
          {value}
          <ChevronDownIcon
            className={cn(
              "w-3.5 h-3.5 text-text-muted transition-transform",
              open && "rotate-180",
            )}
            strokeWidth={2.5}
          />
        </button>
      </Dropdown.Trigger>

      <Dropdown.Panel aria-label="Font families" className="w-48">
        {TERMINAL_FONTS.map((font) => (
          <Dropdown.Item
            key={font}
            label={font}
            role="menuitemradio"
            aria-checked={font === value}
            onSelect={() => onChange(font)}
            className={cn(
              "px-3 py-2 text-sm",
              font === value && "text-primary bg-primary/10",
            )}
          >
            <span style={{ fontFamily: `"${font}", monospace` }}>{font}</span>
            {font === value && (
              <CheckIcon
                className="ml-auto w-3.5 h-3.5 shrink-0"
                strokeWidth={2}
              />
            )}
          </Dropdown.Item>
        ))}
      </Dropdown.Panel>
    </Dropdown>
  );
}

function ThemeCard({
  theme,
  selected,
  onClick,
}: {
  theme: TerminalTheme;
  selected: boolean;
  onClick: () => void;
}) {
  const light = isLightTheme(theme.colors.background);
  const swatches = [
    theme.colors.red,
    theme.colors.green,
    theme.colors.yellow,
    theme.colors.blue,
    theme.colors.magenta,
    theme.colors.cyan,
  ];

  return (
    <button
      type="button"
      onClick={onClick}
      className={cn(
        "relative rounded-lg border p-2 text-left transition-all duration-150",
        selected
          ? "border-primary/40 bg-primary/[0.06] ring-1 ring-primary/10"
          : "border-border hover:border-border-light bg-hover-subtle",
      )}
    >
      {/* Color preview */}
      <div
        className="rounded mb-1.5 px-1.5 py-1"
        style={{ backgroundColor: theme.colors.background }}
      >
        <div
          className="font-mono text-2xs mb-0.5 truncate"
          style={{ color: theme.colors.green }}
        >
          $ ssh root@dev
        </div>
        <div className="flex gap-[2px]">
          {swatches.map((color, i) => (
            <div
              key={i}
              className="h-1.5 flex-1 rounded-sm"
              style={{ backgroundColor: color }}
            />
          ))}
        </div>
      </div>

      {/* Name */}
      <div
        className={cn(
          "font-mono text-[11px] truncate",
          selected
            ? "text-primary"
            : light
              ? "text-text-muted"
              : "text-text-secondary",
        )}
      >
        {theme.name}
      </div>

      {/* Selected indicator */}
      {selected && (
        <div className="absolute top-1.5 right-1.5">
          <CheckIcon className="w-3 h-3 text-primary" strokeWidth={2.5} />
        </div>
      )}
    </button>
  );
}
