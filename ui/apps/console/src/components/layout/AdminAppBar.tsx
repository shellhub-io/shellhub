import { useNavigate } from "react-router-dom";
import {
  ArrowRightStartOnRectangleIcon,
  Bars3Icon,
} from "@heroicons/react/24/outline";
import { useAuthStore } from "@/stores/authStore";
import { getInitials } from "@/utils/string";
import NamespaceSelector from "./NamespaceSelector";

interface AdminAppBarProps {
  onMenuToggle?: () => void;
}

export default function AdminAppBar({ onMenuToggle }: AdminAppBarProps) {
  const user = useAuthStore((s) => s.user);
  const email = useAuthStore((s) => s.email);
  const logout = useAuthStore((s) => s.logout);
  const navigate = useNavigate();

  const handleLogout = () => {
    logout();
    void navigate("/login");
  };

  return (
    <header className="h-14 bg-surface border-b border-border px-3 sm:px-5 flex items-center justify-between shrink-0">
      <div className="flex items-center gap-1">
        {onMenuToggle && (
          <button
            type="button"
            onClick={onMenuToggle}
            className="lg:hidden p-2 -ml-1 rounded-md text-text-muted hover:text-text-primary hover:bg-hover-subtle transition-colors"
            aria-label="Open navigation menu"
          >
            <Bars3Icon className="w-5 h-5" />
          </button>
        )}
        <NamespaceSelector isAdminContext />
      </div>
      <div className="flex items-center gap-3">
        <div className="flex items-center gap-2">
          <span className="w-6 h-6 rounded-md bg-accent-red/15 border border-accent-red/20 flex items-center justify-center text-accent-red text-2xs font-bold font-mono">
            {getInitials(user ?? "A")}
          </span>
          <span className="hidden md:inline text-xs font-medium text-text-secondary max-w-[160px] truncate">
            {email ?? user ?? "Admin"}
          </span>
        </div>
        <button
          type="button"
          onClick={handleLogout}
          className="flex items-center gap-1.5 px-2.5 py-1.5 rounded-md text-xs text-text-muted hover:text-accent-red hover:bg-accent-red/5 transition-colors"
          aria-label="Log out"
        >
          <ArrowRightStartOnRectangleIcon className="w-4 h-4" />
          <span className="hidden sm:inline">Logout</span>
        </button>
      </div>
    </header>
  );
}
