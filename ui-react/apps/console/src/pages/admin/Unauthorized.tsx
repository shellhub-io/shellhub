import { useNavigate } from "react-router-dom";
import {
  ShieldExclamationIcon,
  ArrowLeftIcon,
  ArrowRightStartOnRectangleIcon,
  HomeIcon,
  UserGroupIcon,
  CpuChipIcon,
} from "@heroicons/react/24/outline";
import { useAuthStore } from "@/stores/authStore";

const highlights = [
  {
    icon: <HomeIcon className="w-5 h-5" />,
    title: "Main Application",
    description:
      "Return to the main ShellHub application to manage your resources.",
  },
  {
    icon: <UserGroupIcon className="w-5 h-5" />,
    title: "Request Access",
    description:
      "Contact your system administrator to request admin privileges.",
  },
  {
    icon: <CpuChipIcon className="w-5 h-5" />,
    title: "Your Workspace",
    description:
      "Manage your devices, sessions, and namespaces in the main app.",
  },
];

export default function AdminUnauthorized() {
  const logout = useAuthStore((s) => s.logout);
  const navigate = useNavigate();

  const handleLogout = () => {
    logout();
    void navigate("/login");
  };

  return (
    <section
      aria-labelledby="unauthorized-heading"
      className="relative -mx-8 -mt-8 min-h-[calc(100vh-3.5rem)] flex flex-col"
    >
      {/* Background */}
      <div className="absolute inset-0 overflow-hidden pointer-events-none" aria-hidden="true">
        <div className="absolute -top-32 left-1/3 w-[500px] h-[500px] bg-accent-yellow/5 rounded-full blur-[120px] animate-pulse-subtle" />
        <div
          className="absolute bottom-0 right-1/4 w-[400px] h-[400px] bg-primary/5 rounded-full blur-[100px] animate-pulse-subtle"
          style={{ animationDelay: "1s" }}
        />
        <div className="absolute inset-0 grid-bg opacity-30" />
      </div>

      <div className="flex-1 flex items-center justify-center px-8 py-12">
        <div className="w-full max-w-2xl animate-fade-in">
          {/* Header */}
          <header className="text-center mb-10">
            <div className="w-16 h-16 rounded-2xl bg-accent-yellow/10 border border-accent-yellow/20 flex items-center justify-center mx-auto mb-6 shadow-lg shadow-accent-yellow/5">
              <ShieldExclamationIcon className="w-8 h-8 text-accent-yellow" aria-hidden="true" />
            </div>

            <span className="inline-block text-2xs font-mono font-semibold uppercase tracking-wide text-accent-yellow/80 mb-2">
              Access Restricted
            </span>
            <h1
              id="unauthorized-heading"
              className="text-3xl font-bold text-text-primary mb-3"
            >
              Admin Access Required
            </h1>
            <p className="text-sm text-text-muted max-w-md mx-auto leading-relaxed">
              You don&apos;t have administrator privileges to access the Admin
              Console. This area is restricted to system administrators only.
            </p>
          </header>

          {/* Highlights */}
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-10">
            {highlights.map((h, idx) => (
              <div
                key={h.title}
                className="bg-card/60 border border-border rounded-xl p-5 text-center animate-slide-up"
                style={{ animationDelay: `${150 + idx * 100}ms` }}
              >
                <div className="w-10 h-10 rounded-lg bg-primary/10 border border-primary/20 flex items-center justify-center mx-auto mb-3 text-primary">
                  {h.icon}
                </div>
                <h3 className="text-sm font-semibold text-text-primary mb-1">
                  {h.title}
                </h3>
                <p className="text-xs text-text-muted leading-relaxed">
                  {h.description}
                </p>
              </div>
            ))}
          </div>

          {/* Actions */}
          <footer
            className="flex flex-col sm:flex-row items-center justify-center gap-3 animate-slide-up"
            style={{ animationDelay: "450ms" }}
          >
            <button
              type="button"
              onClick={() => void navigate("/dashboard")}
              className="inline-flex items-center gap-2 px-6 py-3 bg-primary hover:bg-primary-600 text-white rounded-lg text-sm font-semibold transition-all shadow-lg shadow-primary/20"
            >
              <ArrowLeftIcon className="w-4 h-4" aria-hidden="true" />
              Go to ShellHub
            </button>
            <button
              type="button"
              onClick={handleLogout}
              className="inline-flex items-center gap-2 px-6 py-3 border border-border rounded-lg text-sm font-medium text-text-secondary hover:bg-hover-medium transition-colors"
            >
              <ArrowRightStartOnRectangleIcon className="w-4 h-4" aria-hidden="true" />
              Logout
            </button>
          </footer>
          <p className="mt-4 text-center text-2xs text-text-muted">
            If you believe you should have admin access, contact your system administrator.
          </p>
        </div>
      </div>
    </section>
  );
}
