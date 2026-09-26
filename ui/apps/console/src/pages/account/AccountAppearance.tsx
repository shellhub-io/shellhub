import { useId } from "react";
import SettingsSection from "@/components/settings/SettingsSection";
import SettingsField from "@/components/settings/SettingsField";
import RadioGroupField from "@/components/common/fields/RadioGroupField";
import RadioTile from "@/components/common/fields/RadioTile";
import {
  THEME_PREFERENCES,
  useThemeStore,
  type AppTheme,
  type ThemePreference,
} from "@/stores/themeStore";
import {
  SIDEBAR_PINS,
  useSidebarStore,
  type SidebarPin,
} from "@/stores/sidebarStore";
import TerminalSettingsPreview from "./TerminalSettingsPreview";

function WindowPane({
  scheme,
  part,
}: {
  scheme: AppTheme;
  part: "whole" | "right half";
}) {
  const whole = part === "whole";
  const x = whole ? 0 : 32;
  return (
    <g className={scheme === "light" ? "theme-light" : "theme-dark"}>
      <rect x={x} width={64 - x} height="36" className="fill-surface" />
      {whole && <rect width="14" height="36" className="fill-border" />}
      <path
        d={whole ? "M20 9h24M20 15h32M20 21h18" : "M32 9h12M32 15h20M32 21h6"}
        strokeWidth="2"
        className="stroke-text-muted/50"
      />
    </g>
  );
}

function WindowSketch({ preference }: { preference: ThemePreference }) {
  const clip = useId();
  return (
    <svg viewBox="0 0 64 36" className="w-full h-auto" aria-hidden="true">
      <defs>
        <clipPath id={clip}>
          <rect x="1" y="1" width="62" height="34" rx="3.5" />
        </clipPath>
      </defs>
      <g clipPath={`url(#${clip})`}>
        <WindowPane
          scheme={preference === "dark" ? "dark" : "light"}
          part="whole"
        />
        {preference === "system" && (
          <WindowPane scheme="dark" part="right half" />
        )}
      </g>
      <rect
        x="1"
        y="1"
        width="62"
        height="34"
        rx="3.5"
        fill="none"
        stroke="currentColor"
      />
    </svg>
  );
}

function SidebarSketch({ pin }: { pin: SidebarPin }) {
  const open = pin === "pinned";
  return (
    <svg
      viewBox="0 0 64 36"
      fill="none"
      stroke="currentColor"
      strokeWidth={1}
      className="w-full h-auto"
      aria-hidden="true"
    >
      <rect x="1" y="1" width="62" height="34" rx="3.5" />
      <rect
        x="4"
        y="4"
        width={open ? 16 : 6}
        height="28"
        rx="1.5"
        strokeDasharray={pin === "auto" ? "2.5 2" : undefined}
      />
      {pin === "auto" && (
        <rect x="4" y="4" width="16" height="28" rx="1.5" opacity={0.35} />
      )}
      <path
        d={
          open || pin === "auto"
            ? "M25 10h28M25 16h22M25 22h26"
            : "M15 10h38M15 16h30M15 22h34"
        }
        opacity={0.5}
      />
    </svg>
  );
}

/**
 * The appearance section of the account: the console's colours, how the sidebar sits, the
 * terminal's theme and font, and how the session player shows its controls. Every choice applies
 * at once and is kept in this browser, not in the account, so another browser keeps its own.
 */
export default function AccountAppearance() {
  const preference = useThemeStore((s) => s.preference);
  const setPreference = useThemeStore((s) => s.setPreference);
  const pin = useSidebarStore((s) => s.pin);
  const setPin = useSidebarStore((s) => s.setPin);
  const themeId = useId();
  const sidebarId = useId();

  return (
    <SettingsSection
      wide
      title="Appearance"
      description="How the console looks and behaves here. Kept in this browser, not in your account."
    >
      <SettingsField
        stacked
        titleId={themeId}
        title="Theme"
        description="The console's colours. System follows your operating system."
      >
        <RadioGroupField
          labelledBy={themeId}
          value={preference}
          onChange={setPreference}
          containerClassName="grid grid-cols-3 gap-2 max-w-md"
        >
          {THEME_PREFERENCES.map(({ value, label }) => (
            <RadioTile
              key={value}
              value={value}
              label={label}
              picture={<WindowSketch preference={value} />}
            />
          ))}
        </RadioGroupField>
      </SettingsField>

      <SettingsField
        stacked
        titleId={sidebarId}
        title="Sidebar"
        description="On a desktop window. Automatic keeps it open when the window is wide and folds it to icons when it is narrower."
      >
        <RadioGroupField
          labelledBy={sidebarId}
          value={pin}
          onChange={setPin}
          containerClassName="grid grid-cols-3 gap-2 max-w-md"
        >
          {SIDEBAR_PINS.map(({ value, label }) => (
            <RadioTile
              key={value}
              value={value}
              label={label}
              picture={<SidebarSketch pin={value} />}
            />
          ))}
        </RadioGroupField>
      </SettingsField>

      <SettingsField
        stacked
        title="Terminal"
        description="The theme and font of every terminal, and how the session player shows its controls."
      >
        <TerminalSettingsPreview />
      </SettingsField>
    </SettingsSection>
  );
}
