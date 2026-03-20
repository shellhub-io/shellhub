import { useNavigate } from "react-router-dom";
import { ArrowRightStartOnRectangleIcon } from "@heroicons/react/24/outline";
import { useAuthStore } from "../../stores/authStore";
import { getInitials } from "../../utils/string";

export default function AdminAppBar() {
  const user = useAuthStore((s) => s.user);
  const email = useAuthStore((s) => s.email);
  const logout = useAuthStore((s) => s.logout);
  const navigate = useNavigate();

  const handleLogout = () => {
    logout();
    void navigate("/login");
  };

  return (
    <header className="h-14 bg-surface border-b border-border px-5 flex items-center justify-between shrink-0">
      <div className="flex items-center gap-2">
        <span className="text-2xs font-mono font-semibold uppercase tracking-label text-text-muted/60">
          Admin Panel
        </span>
      </div>
      <div className="flex items-center gap-3">
        <div className="flex items-center gap-2">
          <span className="w-6 h-6 rounded-md bg-accent-red/15 border border-accent-red/20 flex items-center justify-center text-accent-red text-2xs font-bold font-mono">
            {getInitials(user ?? "A")}
          </span>
          <span className="text-xs font-medium text-text-secondary max-w-[160px] truncate">
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
