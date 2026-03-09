import { useState } from "react";
import { Link, useNavigate, useSearchParams } from "react-router-dom";
import {
  EnvelopeIcon,
  ExclamationCircleIcon,
  CheckCircleIcon,
} from "@heroicons/react/24/outline";
import { resendEmail } from "../api/auth";

export default function ConfirmAccount() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const username = searchParams.get("username") ?? "";

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleResend = async () => {
    if (!username) return;
    setLoading(true);
    setError(null);
    try {
      await resendEmail(username);
      navigate("/login");
    } catch {
      setError(
        "An error occurred while sending the email. Please try again.",
      );
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="w-full max-w-5xl mx-auto px-8 py-12 flex flex-col items-center">
      {/* Hero */}
      <div className="text-center mb-12 animate-fade-in">
        <div className="animate-float mb-6 inline-block">
          <div className="w-20 h-20 rounded-2xl bg-primary/15 border border-primary/25 flex items-center justify-center shadow-lg shadow-primary/10">
            <EnvelopeIcon
              className="w-10 h-10 text-primary"
              strokeWidth={1.2}
            />
          </div>
        </div>

        <p className="text-2xs font-mono font-semibold uppercase tracking-wide text-primary/80 mb-2">
          Activation Required
        </p>
        <h1 className="text-3xl font-bold text-text-primary mb-3">
          Account Activation Required
        </h1>
        <p className="text-sm text-text-muted max-w-md mx-auto leading-relaxed">
          Thank you for registering an account on ShellHub. An email was sent
          with a confirmation link. You need to click on the link to activate
          your account. If you haven&apos;t received the email, click on Resend
          Email below.
        </p>
      </div>

      {/* Card */}
      <div
        className="w-full max-w-sm bg-card/80 border border-border rounded-2xl p-8 backdrop-blur-sm animate-slide-up"
        style={{ animationDelay: "200ms" }}
      >
        <div className="space-y-5">
          {error && (
            <div className="flex items-center gap-2 bg-accent-red/8 border border-accent-red/20 text-accent-red px-3.5 py-2.5 rounded-md text-xs font-mono animate-slide-down">
              <ExclamationCircleIcon
                className="w-3.5 h-3.5 shrink-0"
                strokeWidth={2}
              />
              {error}
            </div>
          )}

          {!username && (
            <div className="flex items-center gap-2 bg-accent-yellow/8 border border-accent-yellow/20 text-accent-yellow px-3.5 py-2.5 rounded-md text-xs font-mono">
              <CheckCircleIcon
                className="w-3.5 h-3.5 shrink-0"
                strokeWidth={2}
              />
              No username provided. Please go back to login and try again.
            </div>
          )}

          <button
            type="button"
            onClick={handleResend}
            disabled={loading || !username}
            className="w-full bg-primary hover:bg-primary-600 text-white py-3 px-4 rounded-lg text-sm font-semibold disabled:opacity-dim disabled:cursor-not-allowed transition-all duration-200"
          >
            {loading ? (
              <span className="flex items-center justify-center gap-2">
                <span className="w-3.5 h-3.5 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                <span className="font-mono text-xs">Sending...</span>
              </span>
            ) : (
              "Resend Email"
            )}
          </button>

          <p className="text-center text-xs text-text-muted">
            Back to{" "}
            <Link
              to="/login"
              className="text-primary hover:text-primary/80 transition-colors"
            >
              Login
            </Link>
          </p>
        </div>
      </div>
    </div>
  );
}
