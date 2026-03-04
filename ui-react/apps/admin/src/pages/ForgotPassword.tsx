import { useState, FormEvent } from "react";
import { Link } from "react-router-dom";
import {
  EnvelopeIcon,
  CheckCircleIcon,
  LockClosedIcon,
} from "@heroicons/react/24/outline";
import { recoverPassword } from "../api/auth";

export default function ForgotPassword() {
  const [account, setAccount] = useState("");
  const [loading, setLoading] = useState(false);
  const [sent, setSent] = useState(false);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setLoading(true);
    try {
      await recoverPassword(account.trim());
    } catch {
      // Silently ignore to prevent user enumeration.
    } finally {
      setLoading(false);
      setSent(true);
    }
  };

  return (
    <div className="w-full max-w-5xl mx-auto px-8 py-12 flex flex-col items-center">
      {/* Hero */}
      <div className="text-center mb-12 animate-fade-in">
        <div className="animate-float mb-6 inline-block">
          <div className="w-20 h-20 rounded-2xl bg-primary/15 border border-primary/25 flex items-center justify-center shadow-lg shadow-primary/10">
            <LockClosedIcon className="w-10 h-10 text-primary" strokeWidth={1.2} />
          </div>
        </div>

        <p className="text-2xs font-mono font-semibold uppercase tracking-wide text-primary/80 mb-2">
          Password Recovery
        </p>
        <h1 className="text-3xl font-bold text-text-primary mb-3">
          Forgot your password?
        </h1>
        <p className="text-sm text-text-muted max-w-md mx-auto leading-relaxed">
          Enter your username or email address and we&apos;ll send you a link to
          reset your password.
        </p>
      </div>

      {/* Card */}
      <div
        className="w-full max-w-sm bg-card/80 border border-border rounded-2xl p-8 backdrop-blur-sm animate-slide-up"
        style={{ animationDelay: "200ms" }}
      >
        {sent ? (
          <div role="alert" className="flex flex-col items-center text-center gap-4">
            <div className="w-12 h-12 rounded-full bg-accent-green/15 border border-accent-green/25 flex items-center justify-center">
              <CheckCircleIcon className="w-6 h-6 text-accent-green" strokeWidth={1.5} />
            </div>
            <div>
              <p className="text-sm font-semibold text-text-primary mb-1">Check your inbox</p>
              <p className="text-xs text-text-muted leading-relaxed">
                An email with password reset instructions has been sent to your
                registered email address.
              </p>
            </div>
          </div>
        ) : (
          <form onSubmit={handleSubmit} className="space-y-5">
            <div>
              <label
                htmlFor="account"
                className="block text-2xs font-mono font-semibold uppercase tracking-label text-text-muted mb-2.5"
              >
                Username or email address
              </label>
              <input
                id="account"
                type="text"
                value={account}
                onChange={(e) => setAccount(e.target.value)}
                required
                autoFocus
                autoComplete="username"
                className="w-full px-4 py-3 bg-background border border-border rounded-lg text-sm text-text-primary font-mono placeholder:text-text-secondary focus:outline-none focus:border-primary/50 focus:ring-1 focus:ring-primary/20 transition-all duration-200"
                placeholder="username or email"
              />
            </div>

            <button
              type="submit"
              disabled={loading || !account.trim()}
              className="w-full bg-primary hover:bg-primary-600 text-white py-3 px-4 rounded-lg text-sm font-semibold disabled:opacity-dim disabled:cursor-not-allowed transition-all duration-200 mt-1 flex items-center justify-center gap-2"
            >
              {loading ? (
                <>
                  <span className="w-3.5 h-3.5 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                  <span className="font-mono text-xs">Sending...</span>
                </>
              ) : (
                <>
                  <EnvelopeIcon className="w-4 h-4" strokeWidth={2} />
                  Reset Password
                </>
              )}
            </button>
          </form>
        )}
      </div>

      {/* Back to login */}
      <div
        className="mt-8 animate-fade-in"
        style={{ animationDelay: "600ms" }}
      >
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
