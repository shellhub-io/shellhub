import { useState, FormEvent } from "react";
import { useNavigate } from "react-router-dom";
import { ExclamationCircleIcon } from "@heroicons/react/24/outline";
import { useAuthStore } from "../stores/authStore";

export default function Login() {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const { login, loading, error } = useAuthStore();
  const navigate = useNavigate();

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    await login(username, password);
    if (useAuthStore.getState().isLoggedIn) {
      navigate("/dashboard");
    }
  };

  return (
    <div className="w-full max-w-sm mx-auto animate-fade-in">
      <div className="bg-surface border border-border rounded-lg overflow-hidden">
        <div className="px-8 pt-8 pb-6 border-b border-border bg-card/50">
          <div className="flex justify-center mb-5">
            <img src="/v2/logo-inverted.png" alt="ShellHub" className="h-7" />
          </div>
          <p className="text-center text-2xs font-mono text-text-muted tracking-wider uppercase">
            Secure Shell Access
          </p>
        </div>

        <div className="p-8">
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
                className="block text-2xs font-mono font-semibold uppercase tracking-[0.15em] text-text-muted mb-2"
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
                className="w-full px-3.5 py-2.5 bg-card border border-border rounded-md text-sm text-text-primary font-mono placeholder:text-text-secondary focus:outline-none focus:border-primary/50 focus:ring-1 focus:ring-primary/20 transition-all duration-200"
                placeholder="username"
              />
            </div>

            <div>
              <label
                htmlFor="password"
                className="block text-2xs font-mono font-semibold uppercase tracking-[0.15em] text-text-muted mb-2"
              >
                Password
              </label>
              <input
                id="password"
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
                className="w-full px-3.5 py-2.5 bg-card border border-border rounded-md text-sm text-text-primary font-mono placeholder:text-text-secondary focus:outline-none focus:border-primary/50 focus:ring-1 focus:ring-primary/20 transition-all duration-200"
                placeholder="password"
              />
            </div>

            <button
              type="submit"
              disabled={loading}
              className="w-full bg-primary hover:bg-primary-600 text-white py-2.5 px-4 rounded-md text-sm font-semibold disabled:opacity-40 disabled:cursor-not-allowed transition-all duration-200 mt-1"
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
      </div>

      <p className="text-center text-2xs font-mono text-text-muted/40 mt-6">
        ShellHub &mdash; Secure Remote Access
      </p>
    </div>
  );
}
