import {
  AdjustmentsHorizontalIcon,
  ClockIcon,
  CommandLineIcon,
  KeyIcon,
  SwatchIcon,
  VideoCameraIcon,
} from "@heroicons/react/24/outline";
import SectionedLayout from "@/components/settings/SectionedLayout";
import { PREFERENCES_PATH } from "@/utils/preferencesRoute";

const PREFERENCES_SECTIONS = [
  { to: "appearance", label: "Appearance", icon: SwatchIcon },
  { to: "terminal", label: "Terminal", icon: CommandLineIcon },
  { to: "recordings", label: "Local recordings", icon: VideoCameraIcon },
  { to: "browser-identity", label: "Browser identity", icon: KeyIcon },
  { to: "recent-devices", label: "Recent devices", icon: ClockIcon },
];

/**
 * This browser's preferences, one section per URL under /preferences. Everything here is kept in
 * the browser, not in the account: it belongs to no namespace, and another browser keeps its own.
 */
export default function PreferencesLayout() {
  return (
    <SectionedLayout
      base={PREFERENCES_PATH}
      icon={<AdjustmentsHorizontalIcon className="w-6 h-6" />}
      title="Preferences"
      description="How the console works in this browser. Kept here, not in your account."
      sections={PREFERENCES_SECTIONS}
    />
  );
}
