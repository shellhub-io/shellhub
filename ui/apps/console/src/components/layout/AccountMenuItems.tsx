import { useNavigate } from "react-router-dom";
import {
  ArrowRightStartOnRectangleIcon,
  Cog6ToothIcon,
  ComputerDesktopIcon,
  MoonIcon,
  SunIcon,
  UserIcon,
} from "@heroicons/react/24/outline";
import { cn } from "@shellhub/design-system/cn";
import { useAuthStore } from "@/stores/authStore";
import { useNamespaces } from "@/hooks/useNamespaces";
import { useThemeStore, type ThemePreference } from "@/stores/themeStore";

const itemClass =
  "w-full flex items-center gap-2.5 px-3 py-2 rounded-md text-left hover:bg-hover-medium transition-colors group";
const itemIconClass =
  "w-4 h-4 text-text-muted group-hover:text-text-primary transition-colors";
const itemLabelClass =
  "text-sm text-text-secondary group-hover:text-text-primary transition-colors";

const THEME_OPTIONS: {
  value: ThemePreference;
  label: string;
  Icon: typeof SunIcon;
}[] = [
  { value: "system", label: "Match system", Icon: ComputerDesktopIcon },
  { value: "light", label: "Light", Icon: SunIcon },
  { value: "dark", label: "Dark", Icon: MoonIcon },
];

/**
 * The account actions every account menu offers: profile, settings (once there is a namespace to
 * set), the theme (following the system unless fixed), and sign-out. onDone closes the menu
 * holding them.
 */
export default function AccountMenuItems({ onDone }: { onDone: () => void }) {
  const navigate = useNavigate();
  const logout = useAuthStore((s) => s.logout);
  const { namespaces } = useNamespaces();
  const preference = useThemeStore((s) => s.preference);
  const setPreference = useThemeStore((s) => s.setPreference);

  const go = (path: string) => {
    onDone();
    void navigate(path);
  };

  return (
    <>
      <div className="p-1.5">
        <button
          type="button"
          onClick={() => go("/profile")}
          className={itemClass}
        >
          <UserIcon className={itemIconClass} />
          <span className={itemLabelClass}>Profile</span>
        </button>
        {namespaces.length > 0 && (
          <button
            type="button"
            onClick={() => go("/settings")}
            className={itemClass}
          >
            <Cog6ToothIcon className={itemIconClass} />
            <span className={itemLabelClass}>Settings</span>
          </button>
        )}
        <div className="flex items-center justify-between gap-2 px-3 py-1.5">
          <span className="text-sm text-text-secondary">Theme</span>
          <div
            role="radiogroup"
            aria-label="Theme"
            className="flex items-center gap-0.5 p-0.5 rounded-md bg-hover-subtle"
          >
            {THEME_OPTIONS.map(({ value, label, Icon }) => (
              <button
                key={value}
                type="button"
                role="radio"
                aria-checked={preference === value}
                aria-label={label}
                title={label}
                onClick={() => setPreference(value)}
                className={cn(
                  "w-7 h-6 rounded flex items-center justify-center transition-colors",
                  preference === value
                    ? "bg-card text-text-primary shadow-sm"
                    : "text-text-muted hover:text-text-primary",
                )}
              >
                <Icon className="w-3.5 h-3.5" />
              </button>
            ))}
          </div>
        </div>
      </div>

      <div className="p-1.5 border-t border-border">
        <button
          type="button"
          onClick={() => {
            onDone();
            logout();
            void navigate("/login");
          }}
          className="w-full flex items-center gap-2.5 px-3 py-2 rounded-md text-left hover:bg-accent-red/5 transition-colors group"
        >
          <ArrowRightStartOnRectangleIcon className="w-4 h-4 text-text-muted group-hover:text-accent-red transition-colors" />
          <span className="text-sm text-text-muted group-hover:text-accent-red transition-colors">
            Logout
          </span>
        </button>
      </div>
    </>
  );
}
