import { useState, FormEvent } from "react";
import { Link, useSearchParams, useNavigate } from "react-router-dom";
import {
  ExclamationCircleIcon,
  LockClosedIcon,
} from "@heroicons/react/24/outline";
import { updateRecoverPassword } from "../client";
import { validatePassword } from "../utils/validation";
import PasswordField from "@/components/common/fields/PasswordField";

export default function UpdatePassword() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();

  const uid = searchParams.get("id") ?? "";
  const token = searchParams.get("token") ?? "";

  const [password, setPassword] = useState("");
  const [confirm, setConfirm] = useState("");
  const [touched, setTouched] = useState<Record<string, boolean>>({});
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const rawPasswordError = validatePassword(password);
  const passwordError = touched.password ? rawPasswordError : null;
  const confirmError =
    touched.confirm && password !== confirm ? "Passwords do not match" : null;

  const isValid = !rawPasswordError && password === confirm;

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    if (!isValid || loading) return;
    setError("");
    setLoading(true);
    try {
      await updateRecoverPassword({
        path: { uid },
        body: { token, password },
        throwOnError: true,
      });
      void navigate("/login", {
        state: { notice: "Password updated successfully. Please sign in." },
      });
    } catch {
      setError(
        "Failed to update password. The link may have expired. Please request a new one.",
      );
    } finally {
      setLoading(false);
    }
  };

  if (!uid || !token) {
    return (
      <div className="w-full max-w-5xl mx-auto px-8 py-12 flex flex-col items-center">
        <div className="w-full max-w-sm bg-card/80 border border-border rounded-2xl p-8 text-center animate-fade-in">
          <ExclamationCircleIcon
            className="w-10 h-10 text-accent-red mx-auto mb-4"
            strokeWidth={1.5}
          />
          <p className="text-sm font-semibold text-text-primary mb-2">
            Invalid reset link
          </p>
          <p className="text-xs text-text-muted mb-6">
            This password reset link is invalid or has expired.
          </p>
          <Link
            to="/forgot-password"
            className="text-xs text-primary hover:text-primary-400 transition-colors"
          >
            Request a new reset link
          </Link>
        </div>
      </div>
    );
  }

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
          Password Recovery
        </p>
        <h1 className="text-3xl font-bold text-text-primary mb-3">
          Reset your password
        </h1>
        <p className="text-sm text-text-muted max-w-md mx-auto leading-relaxed">
          Choose a new password for your account.
        </p>
      </div>

      {/* Card */}
      <div
        className="w-full max-w-sm bg-card/80 border border-border rounded-2xl p-8 backdrop-blur-sm animate-slide-up"
        style={{ animationDelay: "200ms" }}
      >
        <form onSubmit={(e) => void handleSubmit(e)} className="space-y-5">
          {error && (
            <div
              role="alert"
              className="flex items-start gap-2 bg-accent-red/8 border border-accent-red/20 text-accent-red px-3.5 py-2.5 rounded-md text-xs font-mono animate-slide-down"
            >
              <ExclamationCircleIcon
                className="w-3.5 h-3.5 shrink-0 mt-0.5"
                strokeWidth={2}
              />
              {error}
            </div>
          )}

          <PasswordField
            id="password"
            label="New Password"
            value={password}
            onChange={setPassword}
            onBlur={() => setTouched((prev) => ({ ...prev, password: true }))}
            placeholder="••••••••"
            error={passwordError ?? undefined}
            hint="5–32 characters"
            autoFocus
            required
          />

          <PasswordField
            id="confirm"
            label="Confirm Password"
            value={confirm}
            onChange={setConfirm}
            onBlur={() => setTouched((prev) => ({ ...prev, confirm: true }))}
            placeholder="••••••••"
            error={confirmError ?? undefined}
            required
          />

          <button
            type="submit"
            disabled={loading || !isValid}
            className="w-full bg-primary hover:bg-primary-600 text-white py-3 px-4 rounded-lg text-sm font-semibold disabled:opacity-dim disabled:cursor-not-allowed transition-all duration-200 mt-1"
          >
            {loading ? (
              <span className="flex items-center justify-center gap-2">
                <span className="w-3.5 h-3.5 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                <span className="font-mono text-xs">Updating...</span>
              </span>
            ) : (
              "Update Password"
            )}
          </button>
        </form>
      </div>

      {/* Back to login */}
      <div className="mt-8 animate-fade-in" style={{ animationDelay: "600ms" }}>
        <Link
          to="/login"
          className="text-xs text-text-muted hover:text-text-secondary transition-colors"
        >
          &larr; Back to login
        </Link>
      </div>
    </div>
  );
}
