import { useState, useRef } from "react";
import { useNavigate } from "react-router-dom";
import {
  ChevronDownIcon,
  UserIcon,
  Cog6ToothIcon,
  ArrowRightStartOnRectangleIcon,
} from "@heroicons/react/24/outline";
import { useAuthStore } from "../../stores/authStore";
import { useClickOutside } from "../../hooks/useClickOutside";

export default function UserMenu() {
  const { user, logout } = useAuthStore();
  const navigate = useNavigate();
  const [open, setOpen] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  useClickOutside(containerRef, () => setOpen(false));

  const handleLogout = () => {
    setOpen(false);
    logout();
    navigate("/login");
  };

  const initials = (name: string) =>
    name
      .split(/[\s-_@.]+/)
      .slice(0, 2)
      .map((w) => w[0]?.toUpperCase() ?? "")
      .join("");

  if (!user) return null;

  return (
    <div ref={containerRef} className="relative">
      <button
        onClick={() => setOpen(!open)}
        className="flex items-center gap-2 h-8 pl-1 pr-2.5 rounded-lg border border-transparent hover:border-border hover:bg-hover-subtle transition-all duration-150"
      >
        <span className="w-6 h-6 rounded-md bg-primary/15 border border-primary/20 flex items-center justify-center text-primary text-2xs font-bold font-mono">
          {initials(user)}
        </span>
        <span className="text-xs font-medium text-text-secondary max-w-[120px] truncate">
          {user}
        </span>
        <ChevronDownIcon
          className={`w-3 h-3 text-text-muted transition-transform duration-200 ${open ? "rotate-180" : ""}`}
          strokeWidth={2.5}
        />
      </button>

      {open && (
        <div className="absolute top-full right-0 mt-1.5 w-56 bg-surface border border-border rounded-lg shadow-2xl shadow-black/40 z-50 overflow-hidden animate-slide-down">
          {/* User info */}
          <div className="p-3.5 border-b border-border">
            <div className="flex items-center gap-3">
              <span className="w-9 h-9 rounded-lg bg-primary/15 border border-primary/20 flex items-center justify-center text-primary text-xs font-bold font-mono shrink-0">
                {initials(user)}
              </span>
              <div className="min-w-0">
                <p className="text-sm font-semibold text-text-primary truncate">
                  {user}
                </p>
                <p className="text-2xs text-text-muted mt-0.5">Logged in</p>
              </div>
            </div>
          </div>

          {/* Menu items */}
          <div className="p-1.5">
            <button
              onClick={() => {
                setOpen(false);
                navigate("/profile");
              }}
              className="w-full flex items-center gap-2.5 px-3 py-2 rounded-md text-left hover:bg-hover-medium transition-colors group"
            >
              <UserIcon className="w-4 h-4 text-text-muted group-hover:text-text-primary transition-colors" />
              <span className="text-sm text-text-secondary group-hover:text-text-primary transition-colors">
                Profile
              </span>
            </button>
            <button
              onClick={() => {
                setOpen(false);
                navigate("/settings");
              }}
              className="w-full flex items-center gap-2.5 px-3 py-2 rounded-md text-left hover:bg-hover-medium transition-colors group"
            >
              <Cog6ToothIcon className="w-4 h-4 text-text-muted group-hover:text-text-primary transition-colors" />
              <span className="text-sm text-text-secondary group-hover:text-text-primary transition-colors">
                Settings
              </span>
            </button>
          </div>

          {/* Logout */}
          <div className="p-1.5 border-t border-border">
            <button
              onClick={handleLogout}
              className="w-full flex items-center gap-2.5 px-3 py-2 rounded-md text-left hover:bg-accent-red/5 transition-colors group"
            >
              <ArrowRightStartOnRectangleIcon className="w-4 h-4 text-text-muted group-hover:text-accent-red transition-colors" />
              <span className="text-sm text-text-muted group-hover:text-accent-red transition-colors">
                Logout
              </span>
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
