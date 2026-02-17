import { useEffect } from "react";
import { CheckIcon, MinusIcon, PlusIcon } from "@heroicons/react/24/outline";
import {
  useTerminalThemeStore,
  TERMINAL_FONTS,
  type TerminalTheme,
} from "../../stores/terminalThemeStore";
import Drawer from "../common/Drawer";
import { getConfig } from "../../env";

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

export default function TerminalSettingsDrawer({ open, onClose }: Props) {
  const {
    themes,
    themeName,
    theme,
    fontFamily,
    fontSize,
    setTheme,
    setFontFamily,
    setFontSize,
    loadThemes,
  } = useTerminalThemeStore();

  useEffect(() => {
    if (open) loadThemes();
  }, [open, loadThemes]);

  return (
    <Drawer
      open={open}
      onClose={onClose}
      title="Terminal Settings"
      width="sm"
      bodyClassName="flex-1 overflow-y-auto"
      footer={
        <>
          <span className="text-2xs font-mono text-text-muted mr-auto">
            {themeName}
          </span>
          <span className="text-2xs font-mono text-text-muted/60">
            {fontFamily} {fontSize}px
          </span>
        </>
      }
    >
      {/* Theme Picker */}
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

      {/* Font Family */}
      <div className="border-b border-border p-4">
        <div className="mb-2.5 text-2xs font-mono font-semibold uppercase tracking-label text-text-muted">
          Font Family
        </div>
        <div className="space-y-0.5">
          {TERMINAL_FONTS.map((font) => (
            <button
              key={font}
              onClick={() => setFontFamily(font)}
              className={`flex w-full items-center gap-2 rounded-md px-2.5 py-1.5 transition-all duration-150 ${
                font === fontFamily
                  ? "bg-primary/10 border border-primary/20"
                  : "hover:bg-hover-subtle border border-transparent"
              }`}
            >
              <span
                className={`text-[13px] ${font === fontFamily ? "text-primary" : "text-text-secondary"}`}
                style={{ fontFamily: `"${font}", monospace` }}
              >
                {font}
              </span>
              {font === fontFamily && (
                <CheckIcon
                  className="ml-auto w-3.5 h-3.5 text-primary shrink-0"
                  strokeWidth={2}
                />
              )}
            </button>
          ))}
        </div>
      </div>

      {/* Font Size */}
      <div className="border-b border-border p-4">
        <div className="mb-2.5 text-2xs font-mono font-semibold uppercase tracking-label text-text-muted">
          Font Size
        </div>
        <div className="flex items-center gap-3">
          <button
            onClick={() => setFontSize(fontSize - 1)}
            disabled={fontSize <= 8}
            className="rounded-md border border-border bg-hover-subtle p-1.5 text-text-muted hover:text-text-primary hover:bg-hover-medium disabled:opacity-faint transition-colors"
          >
            <MinusIcon className="w-3.5 h-3.5" strokeWidth={2} />
          </button>
          <div className="flex-1">
            <input
              type="range"
              min={8}
              max={24}
              value={fontSize}
              onChange={(e) => setFontSize(parseInt(e.target.value))}
              className="w-full accent-primary h-1"
            />
          </div>
          <button
            onClick={() => setFontSize(fontSize + 1)}
            disabled={fontSize >= 24}
            className="rounded-md border border-border bg-hover-subtle p-1.5 text-text-muted hover:text-text-primary hover:bg-hover-medium disabled:opacity-faint transition-colors"
          >
            <PlusIcon className="w-3.5 h-3.5" strokeWidth={2} />
          </button>
          <span className="font-mono text-[13px] text-text-secondary w-7 text-right">
            {fontSize}
          </span>
        </div>
      </div>

      {/* Preview */}
      <div className="p-4">
        <div className="mb-2.5 text-2xs font-mono font-semibold uppercase tracking-label text-text-muted">
          Preview
        </div>
        <div
          className="rounded-lg border border-border p-3 overflow-hidden"
          style={{
            backgroundColor: theme.colors.background,
            fontFamily: `"${fontFamily}", monospace`,
            fontSize: `${fontSize}px`,
            lineHeight: 1.5,
          }}
        >
          <div style={{ color: theme.colors.green }}>$ ssh root@device</div>
          <div style={{ color: theme.colors.foreground }}>
            <span style={{ color: theme.colors.cyan }}>ShellHub</span>{" "}
            <span style={{ color: theme.colors.yellow }}>
              {getConfig().version || "v0.0.0"}
            </span>{" "}
            <span style={{ color: theme.colors.green }}>connected</span>
          </div>
          <div style={{ color: theme.colors.foreground }}>
            <span style={{ color: theme.colors.brightBlack }}>~</span>{" "}
            <span style={{ color: theme.colors.red }}>3</span> devices online
          </div>
          <div className="mt-1">
            <span style={{ color: theme.colors.green }}>$</span>
            <span
              style={{
                color: theme.colors.cursor,
                backgroundColor: theme.colors.cursor,
                marginLeft: "4px",
                display: "inline-block",
                width: "8px",
              }}
            >
              &nbsp;
            </span>
          </div>
        </div>
      </div>
    </Drawer>
  );
}

/* -- Theme Card -- */

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
      onClick={onClick}
      className={`relative rounded-lg border p-2 text-left transition-all duration-150 ${
        selected
          ? "border-primary/40 bg-primary/[0.06] ring-1 ring-primary/10"
          : "border-border hover:border-border-light bg-hover-subtle"
      }`}
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
        className={`font-mono text-[11px] truncate ${
          selected
            ? "text-primary"
            : light
              ? "text-text-muted"
              : "text-text-secondary"
        }`}
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
