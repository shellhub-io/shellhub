import { useContext } from "react";
import { useNavigate } from "react-router-dom";
import {
  ArrowRightStartOnRectangleIcon,
  Cog6ToothIcon,
  ComputerDesktopIcon,
  LifebuoyIcon,
  MoonIcon,
  SunIcon,
  UserIcon,
} from "@heroicons/react/24/outline";
import { cn } from "@shellhub/design-system/cn";
import { Spinner } from "@shellhub/design-system/primitives";
import { ChatwootContext, type ChatwootHandle } from "@/hooks/useChatwoot";
import { useAuthStore } from "@/stores/authStore";
import { useNamespaces } from "@/hooks/useNamespaces";
import { useThemeStore, type ThemePreference } from "@/stores/themeStore";

const itemClass =
  "w-full flex items-center gap-2.5 px-3 py-2 rounded-md text-left hover:bg-hover-medium transition-colors group";
const itemIconClass =
  "w-4 h-4 text-text-muted group-hover:text-text-primary transition-colors";
const itemLabelClass =
  "text-sm text-text-secondary group-hover:text-text-primary transition-colors";

const ISSUE_URL = "https://github.com/shellhub-io/shellhub/issues/new/choose";

const THEME_OPTIONS: {
  value: ThemePreference;
  label: string;
  Icon: typeof SunIcon;
}[] = [
  { value: "system", label: "Match system", Icon: ComputerDesktopIcon },
  { value: "light", label: "Light", Icon: SunIcon },
  { value: "dark", label: "Dark", Icon: MoonIcon },
];

function GettingHelpItem({
  support,
  onDone,
  onNeedsPlan,
}: {
  support: ChatwootHandle;
  onDone: () => void;
  onNeedsPlan?: () => void;
}) {
  const label = <span className={itemLabelClass}>Getting Help</span>;
  const icon = <LifebuoyIcon className={itemIconClass} />;
  const act = (then: () => void) => () => {
    onDone();
    then();
  };

  switch (support.status) {
    case "unavailable":
      return null;
    case "non-cloud":
      return (
        <a
          href={ISSUE_URL}
          target="_blank"
          rel="noopener noreferrer"
          onClick={onDone}
          className={itemClass}
        >
          {icon}
          {label}
        </a>
      );
    case "loading":
      return (
        <button
          type="button"
          aria-disabled
          className={cn(itemClass, "cursor-wait")}
        >
          <Spinner size="sm" tone="subtle" className="block" />
          {label}
        </button>
      );
    case "no-subscription":
      return (
        <button
          type="button"
          onClick={act(() => onNeedsPlan?.())}
          className={itemClass}
        >
          {icon}
          {label}
        </button>
      );
    case "ready":
      return (
        <button
          type="button"
          onClick={act(support.openWidget)}
          className={itemClass}
        >
          {icon}
          {label}
        </button>
      );
  }
}

/**
 * The account actions every account menu offers: profile, settings (once there is a namespace to
 * set), getting help, the theme (following the system unless fixed), and sign-out. onDone closes
 * the menu holding them. Getting help needs the support widget's provider above it and is left
 * out without one; on a plan without chat support it calls onHelpNeedsPlan instead, since the
 * menu closes before a dialog of its own could show.
 */
export default function AccountMenuItems({
  onDone,
  onHelpNeedsPlan,
}: {
  onDone: () => void;
  onHelpNeedsPlan?: () => void;
}) {
  const navigate = useNavigate();
  const support = useContext(ChatwootContext);
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
        {support && (
          <GettingHelpItem
            support={support}
            onDone={onDone}
            onNeedsPlan={onHelpNeedsPlan}
          />
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
