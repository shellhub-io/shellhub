import { useState, useEffect, FormEvent } from "react";
import { useNavigate, Link, useLocation } from "react-router-dom";
import { AxiosError } from "axios";
import {
  ExclamationCircleIcon,
  CheckCircleIcon,
  LockClosedIcon,
} from "@heroicons/react/24/outline";
import { useAuthStore } from "../stores/authStore";
import { getConfig } from "../env";
import AuthFooterLinks from "../components/common/AuthFooterLinks";

interface CountdownState {
  display: string;
  expired: boolean;
  epoch: number | null;
}

function useLoginCountdown(lockoutEndEpoch: number | null) {
  const [state, setState] = useState<CountdownState>({
    display: "",
    expired: false,
    epoch: null,
  });

  useEffect(() => {
    if (lockoutEndEpoch === null) return;

    const interval = setInterval(() => {
      const diff = lockoutEndEpoch - Date.now() / 1000;
      if (diff <= 0) {
        clearInterval(interval);
        setState({ display: "", expired: true, epoch: lockoutEndEpoch });
      } else if (diff < 60) {
        const s = Math.floor(diff);
        setState({
          display: `${s} ${s === 1 ? "second" : "seconds"}`,
          expired: false,
          epoch: lockoutEndEpoch,
        });
      } else {
        const m = Math.floor(diff / 60);
        setState({
          display: `${m} ${m === 1 ? "minute" : "minutes"}`,
          expired: false,
          epoch: lockoutEndEpoch,
        });
      }
    }, 1000);

    return () => clearInterval(interval);
  }, [lockoutEndEpoch]);

  // If the epoch changed (or was cleared), state is stale — return clean defaults
  if (state.epoch !== lockoutEndEpoch) return { display: "", expired: false };

  return { display: state.display, expired: state.expired };
}

export default function Login() {
  const isCloud = getConfig().cloud;
  const location = useLocation();
  const rawState = location.state as Record<string, unknown> | null;
  const notice = typeof rawState?.notice === "string" ? rawState.notice : undefined;

  useEffect(() => {
    if (notice) {
      window.history.replaceState({}, document.title);
    }
  }, [notice]);

  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [lockoutEndEpoch, setLockoutEndEpoch] = useState<number | null>(null);
  const { login, loading } = useAuthStore();
  const navigate = useNavigate();
  const { display: countdownDisplay, expired: lockoutExpired }
    = useLoginCountdown(lockoutEndEpoch);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError(null);
    setLockoutEndEpoch(null);
    try {
      await login(username, password);

      const state = useAuthStore.getState();
      if (state.mfaToken) {
        void navigate("/mfa-login");
      } else {
        void navigate("/dashboard");
      }
    } catch (err) {
      if (err instanceof AxiosError) {
        switch (err.response?.status) {
          case 401:
            setError(
              "Invalid login credentials. Your password is incorrect or this account doesn't exist.",
            );
            break;
          case 403:
            void navigate(
              `/confirm-account?username=${encodeURIComponent(username)}`,
            );
            break;
          case 429: {
            const epoch = Number(err.response.headers["x-account-lockout"]);
            setLockoutEndEpoch(isNaN(epoch) ? null : epoch);
            setError(
              "Too many failed login attempts. Please wait before trying again.",
            );
            break;
          }
          default:
            setError("Something went wrong on our end. Please try again later.");
        }
      } else {
        setError("Something went wrong. Please try again later.");
      }
    }
    // Else: error is already set in store
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
        <form onSubmit={(e) => void handleSubmit(e)} className="space-y-5">
          {lockoutExpired && (
            <div className="flex items-center gap-2 bg-accent-green/8 border border-accent-green/20 text-accent-green px-3.5 py-2.5 rounded-md text-xs font-mono animate-slide-down">
              <CheckCircleIcon className="w-3.5 h-3.5 shrink-0" strokeWidth={2} />
              Your timeout has finished. Please try to log back in.
            </div>
          )}
          {notice && (
            <div role="alert" className="flex items-center gap-2 bg-accent-green/8 border border-accent-green/20 text-accent-green px-3.5 py-2.5 rounded-md text-xs font-mono animate-slide-down">
              <CheckCircleIcon className="w-3.5 h-3.5 shrink-0" strokeWidth={2} />
              {notice}
            </div>
          )}
          {error && !lockoutExpired && (
            <div role="alert" className="flex items-center gap-2 bg-accent-red/8 border border-accent-red/20 text-accent-red px-3.5 py-2.5 rounded-md text-xs font-mono animate-slide-down">
              <ExclamationCircleIcon
                className="w-3.5 h-3.5 shrink-0"
                strokeWidth={2}
              />
              <span>
                {error}
                {countdownDisplay && (
                  <span className="font-semibold">
                    {" "}
                    (
                    {countdownDisplay}
                    )
                  </span>
                )}
              </span>
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

          {isCloud && (
            <div className="flex justify-end">
              <Link
                to="/forgot-password"
                className="text-2xs text-text-muted hover:text-text-secondary transition-colors"
              >
                Forgot password?
              </Link>
            </div>
          )}

          <button
            type="submit"
            disabled={loading}
            className="w-full bg-primary hover:bg-primary-600 text-white py-3 px-4 rounded-lg text-sm font-semibold disabled:opacity-dim disabled:cursor-not-allowed transition-all duration-200 mt-1"
          >
            {loading
              ? (
                <span className="flex items-center justify-center gap-2">
                  <span className="w-3.5 h-3.5 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                  <span className="font-mono text-xs">Authenticating...</span>
                </span>
              )
              : (
                "Sign In"
              )}
          </button>
        </form>
      </div>

      {/* Footer links */}
      <AuthFooterLinks />
    </div>
  );
}
