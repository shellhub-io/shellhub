import { useState, FormEvent } from "react";
import { useNavigate } from "react-router-dom";
import {
  ExclamationCircleIcon,
  LockClosedIcon,
  BookOpenIcon,
} from "@heroicons/react/24/outline";
import { useAuthStore } from "../stores/authStore";

export default function Login() {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const { login, loading, error } = useAuthStore();
  const navigate = useNavigate();

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    await login(username, password);
    if (useAuthStore.getState().token) {
      navigate("/dashboard");
    }
  };

  return (
    <div className="w-full max-w-5xl mx-auto px-8 py-12 flex flex-col items-center">
      {/* Hero */}
      <div className="text-center mb-12 animate-fade-in">
        <div className="animate-float mb-6 inline-block">
          <div className="w-20 h-20 rounded-2xl bg-primary/15 border border-primary/25 flex items-center justify-center shadow-lg shadow-primary/10">
            <LockClosedIcon
              className="w-10 h-10 text-primary"
              strokeWidth={1.2}
            />
          </div>
        </div>

        <p className="text-2xs font-mono font-semibold uppercase tracking-wide text-primary/80 mb-2">
          Welcome Back
        </p>
        <h1 className="text-3xl font-bold text-text-primary mb-3">
          Sign in to ShellHub
        </h1>
        <p className="text-sm text-text-muted max-w-md mx-auto leading-relaxed">
          Access your devices, sessions, and security rules from a single
          dashboard.
        </p>
      </div>

      {/* Form card */}
      <div
        className="w-full max-w-sm bg-card/80 border border-border rounded-2xl p-8 backdrop-blur-sm animate-slide-up"
        style={{ animationDelay: "200ms" }}
      >
        <form onSubmit={handleSubmit} className="space-y-5">
          {error && (
            <div className="flex items-center gap-2 bg-accent-red/8 border border-accent-red/20 text-accent-red px-3.5 py-2.5 rounded-md text-xs font-mono animate-slide-down">
              <ExclamationCircleIcon
                className="w-3.5 h-3.5 shrink-0"
                strokeWidth={2}
              />
              {error}
            </div>
          )}

          <div>
            <label
              htmlFor="username"
              className="block text-2xs font-mono font-semibold uppercase tracking-label text-text-muted mb-2.5"
            >
              Username
            </label>
            <input
              id="username"
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              required
              autoFocus
              className="w-full px-4 py-3 bg-background border border-border rounded-lg text-sm text-text-primary font-mono placeholder:text-text-secondary focus:outline-none focus:border-primary/50 focus:ring-1 focus:ring-primary/20 transition-all duration-200"
              placeholder="username"
            />
          </div>

          <div>
            <label
              htmlFor="password"
              className="block text-2xs font-mono font-semibold uppercase tracking-label text-text-muted mb-2.5"
            >
              Password
            </label>
            <input
              id="password"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              className="w-full px-4 py-3 bg-background border border-border rounded-lg text-sm text-text-primary font-mono placeholder:text-text-secondary focus:outline-none focus:border-primary/50 focus:ring-1 focus:ring-primary/20 transition-all duration-200"
              placeholder="password"
            />
          </div>

          <button
            type="submit"
            disabled={loading}
            className="w-full bg-primary hover:bg-primary-600 text-white py-3 px-4 rounded-lg text-sm font-semibold disabled:opacity-dim disabled:cursor-not-allowed transition-all duration-200 mt-1"
          >
            {loading ? (
              <span className="flex items-center justify-center gap-2">
                <span className="w-3.5 h-3.5 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                <span className="font-mono text-xs">Authenticating...</span>
              </span>
            ) : (
              "Sign In"
            )}
          </button>
        </form>
      </div>

      {/* Footer links */}
      <div
        className="flex items-center justify-center gap-6 mt-10 animate-fade-in"
        style={{ animationDelay: "800ms" }}
      >
        <a
          href="https://docs.shellhub.io"
          target="_blank"
          rel="noopener noreferrer"
          className="inline-flex items-center gap-1.5 text-xs text-text-muted hover:text-text-secondary transition-colors"
        >
          <BookOpenIcon className="w-3.5 h-3.5" />
          Documentation
        </a>
        <span className="w-px h-3 bg-border" />
        <a
          href="https://github.com/shellhub-io/shellhub"
          target="_blank"
          rel="noopener noreferrer"
          className="inline-flex items-center gap-1.5 text-xs text-text-muted hover:text-text-secondary transition-colors"
        >
          <svg className="w-3.5 h-3.5" fill="currentColor" viewBox="0 0 24 24">
            <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z" />
          </svg>
          Community
        </a>
      </div>
    </div>
  );
}
