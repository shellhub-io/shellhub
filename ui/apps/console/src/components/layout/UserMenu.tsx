import { useState } from "react";
import { useNavigate } from "react-router-dom";
import {
  ChevronDownIcon,
  UserIcon,
  Cog6ToothIcon,
  ArrowRightStartOnRectangleIcon,
  MoonIcon,
  SunIcon,
} from "@heroicons/react/24/outline";
import { Dropdown } from "@shellhub/design-system/primitives";
import { cn } from "@shellhub/design-system/cn";
import { useAuthStore } from "@/stores/authStore";
import { useNamespaces } from "@/hooks/useNamespaces";
import { useThemeStore } from "@/stores/themeStore";
import { getInitials } from "@/utils/string";

export default function UserMenu() {
  const { user, name, email, logout } = useAuthStore();
  const navigate = useNavigate();
  const { namespaces } = useNamespaces();
  const theme = useThemeStore((s) => s.theme);
  const toggleTheme = useThemeStore((s) => s.toggleTheme);
  const isDark = theme === "dark";

  const [open, setOpen] = useState(false);

  // SAML users have no username, so fall back to name/email — the account menu
  // (its only way to log out) must always render in the authenticated layout.
  const display = user || name || email || "Account";

  if (!user && !name && !email) return null;

  return (
    <Dropdown mode="content" placement="bottom-end" open={open} onOpenChange={setOpen}>
      <Dropdown.Trigger>
        <button
          type="button"
          aria-label={`Account menu for ${display}`}
          className="flex items-center gap-2 h-8 pl-1 pr-2.5 rounded-lg border border-transparent hover:border-border hover:bg-hover-subtle transition-all duration-150"
        >
          <span className="w-6 h-6 rounded-md bg-primary/15 border border-primary/20 flex items-center justify-center text-primary text-2xs font-bold font-mono">
            {getInitials(display)}
          </span>
          <span className="hidden sm:inline text-xs font-medium text-text-secondary max-w-[120px] truncate">
            {display}
          </span>
          <ChevronDownIcon
            className={cn(
            "w-3 h-3 text-text-muted transition-transform duration-200",
            open && "rotate-180",
          )}
            strokeWidth={2.5}
          />
        </button>
      </Dropdown.Trigger>

      <Dropdown.Panel className="w-56">
        <div className="p-3.5 border-b border-border">
          <div className="flex items-center gap-3">
            <span className="w-9 h-9 rounded-lg bg-primary/15 border border-primary/20 flex items-center justify-center text-primary text-xs font-bold font-mono shrink-0">
              {getInitials(display)}
            </span>
            <div className="min-w-0">
              <p className="text-sm font-semibold text-text-primary truncate">
                {display}
              </p>
              <p className="text-2xs text-text-muted mt-0.5">Logged in</p>
            </div>
          </div>
        </div>

        <div className="p-1.5">
          <button
            type="button"
            onClick={() => { setOpen(false); void navigate("/profile"); }}
            className="w-full flex items-center gap-2.5 px-3 py-2 rounded-md text-left hover:bg-hover-medium transition-colors group"
          >
            <UserIcon className="w-4 h-4 text-text-muted group-hover:text-text-primary transition-colors" />
            <span className="text-sm text-text-secondary group-hover:text-text-primary transition-colors">
              Profile
            </span>
          </button>
          {namespaces.length > 0 && (
            <button
              type="button"
              onClick={() => { setOpen(false); void navigate("/settings"); }}
              className="w-full flex items-center gap-2.5 px-3 py-2 rounded-md text-left hover:bg-hover-medium transition-colors group"
            >
              <Cog6ToothIcon className="w-4 h-4 text-text-muted group-hover:text-text-primary transition-colors" />
              <span className="text-sm text-text-secondary group-hover:text-text-primary transition-colors">
                Settings
              </span>
            </button>
          )}
          <button
            type="button"
            onClick={toggleTheme}
            aria-label={
              isDark ? "Switch to light theme" : "Switch to dark theme"
            }
            className="w-full flex items-center gap-2.5 px-3 py-2 rounded-md text-left hover:bg-hover-medium transition-colors group"
          >
            {isDark ? (
              <SunIcon className="w-4 h-4 text-text-muted group-hover:text-text-primary transition-colors" />
            ) : (
              <MoonIcon className="w-4 h-4 text-text-muted group-hover:text-text-primary transition-colors" />
            )}
            <span className="text-sm text-text-secondary group-hover:text-text-primary transition-colors">
              {isDark ? "Light theme" : "Dark theme"}
            </span>
          </button>
        </div>

        <div className="p-1.5 border-t border-border">
          <button
            type="button"
            onClick={() => {
              setOpen(false);
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
      </Dropdown.Panel>
    </Dropdown>
  );
}
