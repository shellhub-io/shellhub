import { useState } from "react";

const TOGGLE_STYLES = {
  primary: {
    on: "bg-primary/15 text-primary border border-primary/25",
    off: "bg-hover-strong text-text-secondary border border-border-light",
  },
  success: {
    on: "bg-accent-green/15 text-accent-green border border-accent-green/25",
    off: "bg-hover-strong text-text-secondary border border-border-light",
  },
  danger: {
    on: "bg-accent-red/15 text-accent-red border border-accent-red/25",
    off: "bg-hover-strong text-text-secondary border border-border-light",
  },
} as const;

export type SettingToggleTone = keyof typeof TOGGLE_STYLES;

interface SettingToggleProps {
  checked: boolean;
  disabled?: boolean;
  tone?: SettingToggleTone;
  onChange: (checked: boolean) => Promise<void> | void;
}

export default function SettingToggle({
  checked,
  disabled = false,
  tone = "primary",
  onChange,
}: SettingToggleProps) {
  const [loading, setLoading] = useState(false);
  const styles = TOGGLE_STYLES[tone];

  const handleToggle = async (value: boolean) => {
    if (loading || disabled) return;
    setLoading(true);
    try {
      await onChange(value);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div
      className={`inline-flex items-center h-7 bg-card border border-border rounded-md p-0.5 ${loading || disabled ? "opacity-40 pointer-events-none" : ""}`}
    >
      <button
        type="button"
        onClick={() => {
          if (checked) void handleToggle(false);
        }}
        className={`h-full px-2.5 text-2xs font-medium rounded transition-all duration-150 ${
          !checked ? styles.off : "text-text-muted hover:text-text-secondary border border-transparent"
        }`}
      >
        Off
      </button>
      <button
        type="button"
        onClick={() => {
          if (!checked) void handleToggle(true);
        }}
        className={`h-full px-2.5 text-2xs font-medium rounded transition-all duration-150 ${
          checked ? styles.on : "text-text-muted hover:text-text-secondary border border-transparent"
        }`}
      >
        On
      </button>
    </div>
  );
}
