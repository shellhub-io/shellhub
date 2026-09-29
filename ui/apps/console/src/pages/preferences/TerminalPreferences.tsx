import SettingsSection from "@/components/settings/SettingsSection";
import TerminalSettingsPreview from "./TerminalSettingsPreview";

/**
 * How every terminal looks in this browser, its theme, font and size, and how the session player
 * shows its controls. A choice shows at once in the preview beside it.
 */
export default function TerminalPreferences() {
  return (
    <SettingsSection
      wide
      fill
      title="Terminal"
      description="The theme and font of every terminal, and how the session player shows its controls."
    >
      <TerminalSettingsPreview />
    </SettingsSection>
  );
}
