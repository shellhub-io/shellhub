import { useState, FormEvent } from "react";
import { Reveal } from "../landing/components";
import apiClient from "../../api/client";

interface StepSignupProps {
  onBack: () => void;
}

/* ─── Icon helpers ─── */
const UserIcon = () => (
  <svg className="w-4 h-4 text-text-muted" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
    <path strokeLinecap="round" strokeLinejoin="round" d="M15.75 6a3.75 3.75 0 1 1-7.5 0 3.75 3.75 0 0 1 7.5 0ZM4.501 20.118a7.5 7.5 0 0 1 14.998 0A17.933 17.933 0 0 1 12 21.75c-2.676 0-5.216-.584-7.499-1.632Z" />
  </svg>
);
const MailIcon = () => (
  <svg className="w-4 h-4 text-text-muted" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
    <path strokeLinecap="round" strokeLinejoin="round" d="M21.75 6.75v10.5a2.25 2.25 0 0 1-2.25 2.25h-15a2.25 2.25 0 0 1-2.25-2.25V6.75m19.5 0A2.25 2.25 0 0 0 19.5 4.5h-15a2.25 2.25 0 0 0-2.25 2.25m19.5 0v.243a2.25 2.25 0 0 1-1.07 1.916l-7.5 4.615a2.25 2.25 0 0 1-2.36 0L3.32 8.91a2.25 2.25 0 0 1-1.07-1.916V6.75" />
  </svg>
);
const LockIcon = () => (
  <svg className="w-4 h-4 text-text-muted" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
    <path strokeLinecap="round" strokeLinejoin="round" d="M16.5 10.5V6.75a4.5 4.5 0 1 0-9 0v3.75m-.75 11.25h10.5a2.25 2.25 0 0 0 2.25-2.25v-6.75a2.25 2.25 0 0 0-2.25-2.25H6.75a2.25 2.25 0 0 0-2.25 2.25v6.75a2.25 2.25 0 0 0 2.25 2.25Z" />
  </svg>
);

const INPUT_BASE = "w-full pl-10 pr-3.5 py-2.5 bg-surface border border-border rounded-lg text-sm text-text-primary placeholder:text-text-secondary focus:outline-none focus:border-primary/50 focus:ring-1 focus:ring-primary/20 transition-all";

export function StepSignup({ onBack }: StepSignupProps) {
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [acceptPrivacy, setAcceptPrivacy] = useState(false);
  const [acceptMarketing, setAcceptMarketing] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});
  const [success, setSuccess] = useState(false);

  const validate = () => {
    const errors: Record<string, string> = {};
    if (!name.trim()) errors.name = "Name is required";
    if (name.length > 64) errors.name = "Name must be at most 64 characters";
    if (!email.trim()) errors.email = "Email is required";
    else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) errors.email = "Invalid email";
    if (!password) errors.password = "Password is required";
    else if (password.length < 5) errors.password = "At least 5 characters";
    else if (password.length > 32) errors.password = "At most 32 characters";
    if (password !== confirmPassword) errors.confirmPassword = "Passwords do not match";
    if (!acceptPrivacy) errors.privacy = "You must accept the Privacy Policy";
    return errors;
  };

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    const errors = validate();
    setFieldErrors(errors);
    if (Object.keys(errors).length > 0) return;

    setLoading(true);
    setError(null);
    try {
      await apiClient.post("/api/register", {
        name: name.trim(),
        email: email.trim(),
        password,
        email_marketing: acceptMarketing,
      });
      setSuccess(true);
    } catch (err: unknown) {
      if (
        err &&
        typeof err === "object" &&
        "response" in err &&
        (err as { response?: { data?: unknown } }).response?.data
      ) {
        const data = (err as { response: { data: unknown } }).response.data;
        if (Array.isArray(data)) {
          const fe: Record<string, string> = {};
          if (data.includes("email")) fe.email = "This email is already in use";
          if (data.includes("name")) fe.name = "Invalid name";
          if (data.includes("password")) fe.password = "Invalid password";
          setFieldErrors(fe);
        } else {
          setError("Something went wrong. Please try again.");
        }
      } else {
        setError("Something went wrong. Please try again.");
      }
    } finally {
      setLoading(false);
    }
  };

  if (success) {
    return (
      <div className="max-w-md mx-auto w-full">
        <Reveal>
          <div className="bg-card border border-border rounded-xl p-8 text-center">
            <div className="w-12 h-12 rounded-full bg-accent-green/10 border border-accent-green/20 flex items-center justify-center mx-auto mb-4">
              <svg className="w-6 h-6 text-accent-green" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                <path strokeLinecap="round" strokeLinejoin="round" d="m4.5 12.75 6 6 9-13.5" />
              </svg>
            </div>
            <h3 className="text-lg font-bold mb-2">Account created</h3>
            <p className="text-sm text-text-secondary leading-relaxed mb-6">
              Check your email to confirm your account, then log in to get started.
            </p>
            <a
              href="/v2/ui/login"
              className="inline-flex items-center justify-center gap-2 w-full px-6 py-3 text-sm font-semibold bg-primary border border-primary-400/40 text-white rounded-xl hover:brightness-110 hover:scale-[1.02] active:scale-[0.98] transition-all duration-300"
            >
              Go to Login
            </a>
          </div>
        </Reveal>
      </div>
    );
  }

  return (
    <div className="max-w-md mx-auto w-full">
      <Reveal>
        <div className="bg-card border border-border rounded-xl overflow-hidden">
          {/* Header */}
          <div className="px-6 pt-6 pb-5 border-b border-border bg-surface/50">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 rounded-xl bg-primary/10 border border-primary/20 flex items-center justify-center">
                <img src="/v2/cloud-icon.svg" alt="" className="h-5" />
              </div>
              <div>
                <h3 className="text-sm font-bold">ShellHub Cloud</h3>
                <div className="flex items-center gap-2 mt-0.5">
                  <span className="text-2xs text-text-muted">Free account</span>
                  <span className="inline-flex items-center gap-1 px-1.5 py-px text-2xs font-medium text-accent-green bg-accent-green/8 border border-accent-green/15 rounded">
                    <svg className="w-2.5 h-2.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2.5}><path strokeLinecap="round" strokeLinejoin="round" d="m4.5 12.75 6 6 9-13.5" /></svg>
                    No credit card required
                  </span>
                </div>
              </div>
            </div>
          </div>

          {/* Form */}
          <form onSubmit={handleSubmit} className="p-6 space-y-5">
            {error && (
              <div className="flex items-center gap-2 bg-accent-red/8 border border-accent-red/20 text-accent-red px-3.5 py-2.5 rounded-md text-xs font-mono animate-slide-down">
                <svg className="w-3.5 h-3.5 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="M12 9v3.75m9-.75a9 9 0 1 1-18 0 9 9 0 0 1 18 0Zm-9 3.75h.008v.008H12v-.008Z" />
                </svg>
                {error}
              </div>
            )}

            {/* Name */}
            <div>
              <label className="block text-2xs font-mono font-semibold uppercase tracking-[0.15em] text-text-muted mb-1.5">Name</label>
              <div className="relative">
                <span className="absolute left-3 top-1/2 -translate-y-1/2 pointer-events-none"><UserIcon /></span>
                <input
                  type="text"
                  value={name}
                  onChange={(e) => { setName(e.target.value); setFieldErrors((p) => ({ ...p, name: "" })); }}
                  placeholder="Your name"
                  className={INPUT_BASE}
                  autoFocus
                />
              </div>
              {fieldErrors.name && <p className="mt-1 text-2xs text-accent-red">{fieldErrors.name}</p>}
            </div>

            {/* Email */}
            <div>
              <label className="block text-2xs font-mono font-semibold uppercase tracking-[0.15em] text-text-muted mb-1.5">Email</label>
              <div className="relative">
                <span className="absolute left-3 top-1/2 -translate-y-1/2 pointer-events-none"><MailIcon /></span>
                <input
                  type="email"
                  value={email}
                  onChange={(e) => { setEmail(e.target.value); setFieldErrors((p) => ({ ...p, email: "" })); }}
                  placeholder="you@example.com"
                  className={INPUT_BASE}
                />
              </div>
              {fieldErrors.email && <p className="mt-1 text-2xs text-accent-red">{fieldErrors.email}</p>}
            </div>

            {/* Divider */}
            <div className="border-t border-border" />

            {/* Password */}
            <div>
              <label className="block text-2xs font-mono font-semibold uppercase tracking-[0.15em] text-text-muted mb-1.5">Password</label>
              <div className="relative">
                <span className="absolute left-3 top-1/2 -translate-y-1/2 pointer-events-none"><LockIcon /></span>
                <input
                  type={showPassword ? "text" : "password"}
                  value={password}
                  onChange={(e) => { setPassword(e.target.value); setFieldErrors((p) => ({ ...p, password: "", confirmPassword: "" })); }}
                  placeholder="Min. 5 chars"
                  className={`${INPUT_BASE} pr-9`}
                />
                <button
                  type="button"
                  onClick={() => setShowPassword(!showPassword)}
                  className="absolute right-2.5 top-1/2 -translate-y-1/2 text-text-muted hover:text-text-primary transition-colors"
                >
                  {showPassword ? (
                    <svg className="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}><path strokeLinecap="round" strokeLinejoin="round" d="M3.98 8.223A10.477 10.477 0 0 0 1.934 12C3.226 16.338 7.244 19.5 12 19.5c.993 0 1.953-.138 2.863-.395M6.228 6.228A10.451 10.451 0 0 1 12 4.5c4.756 0 8.773 3.162 10.065 7.498a10.522 10.522 0 0 1-4.293 5.774M6.228 6.228 3 3m3.228 3.228 3.65 3.65m7.894 7.894L21 21m-3.228-3.228-3.65-3.65m0 0a3 3 0 1 0-4.243-4.243m4.242 4.242L9.88 9.88" /></svg>
                  ) : (
                    <svg className="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}><path strokeLinecap="round" strokeLinejoin="round" d="M2.036 12.322a1.012 1.012 0 0 1 0-.639C3.423 7.51 7.36 4.5 12 4.5c4.638 0 8.573 3.007 9.963 7.178.07.207.07.431 0 .639C20.577 16.49 16.64 19.5 12 19.5c-4.638 0-8.573-3.007-9.963-7.178Z" /><path strokeLinecap="round" strokeLinejoin="round" d="M15 12a3 3 0 1 1-6 0 3 3 0 0 1 6 0Z" /></svg>
                  )}
                </button>
              </div>
              {fieldErrors.password && <p className="mt-1 text-2xs text-accent-red">{fieldErrors.password}</p>}
            </div>

            {/* Confirm password */}
            <div>
              <label className="block text-2xs font-mono font-semibold uppercase tracking-[0.15em] text-text-muted mb-1.5">Confirm password</label>
              <div className="relative">
                <span className="absolute left-3 top-1/2 -translate-y-1/2 pointer-events-none"><LockIcon /></span>
                <input
                  type={showPassword ? "text" : "password"}
                  value={confirmPassword}
                  onChange={(e) => { setConfirmPassword(e.target.value); setFieldErrors((p) => ({ ...p, confirmPassword: "" })); }}
                  placeholder="Repeat your password"
                  className={INPUT_BASE}
                />
              </div>
              {fieldErrors.confirmPassword && <p className="mt-1 text-2xs text-accent-red">{fieldErrors.confirmPassword}</p>}
            </div>

            {/* Agreements */}
            <div className="space-y-3 pt-1">
              <button
                type="button"
                onClick={() => { setAcceptPrivacy(!acceptPrivacy); setFieldErrors((p) => ({ ...p, privacy: "" })); }}
                className="flex items-start gap-3 w-full text-left group"
              >
                <span className={`mt-px w-4 h-4 rounded flex items-center justify-center shrink-0 border transition-all duration-200 ${acceptPrivacy ? "bg-primary border-primary" : "bg-surface border-border group-hover:border-border-light"}`}>
                  {acceptPrivacy && (
                    <svg className="w-2.5 h-2.5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={3.5}>
                      <path strokeLinecap="round" strokeLinejoin="round" d="m4.5 12.75 6 6 9-13.5" />
                    </svg>
                  )}
                </span>
                <span className="text-xs text-text-secondary leading-relaxed">
                  I agree to the{" "}
                  <a href="https://www.shellhub.io/privacy-policy" target="_blank" rel="noopener noreferrer" onClick={(e) => e.stopPropagation()} className="text-primary hover:underline">Privacy Policy</a>
                </span>
              </button>
              {fieldErrors.privacy && <p className="ml-7 text-2xs text-accent-red">{fieldErrors.privacy}</p>}
              <button
                type="button"
                onClick={() => setAcceptMarketing(!acceptMarketing)}
                className="flex items-start gap-3 w-full text-left group"
              >
                <span className={`mt-px w-4 h-4 rounded flex items-center justify-center shrink-0 border transition-all duration-200 ${acceptMarketing ? "bg-primary border-primary" : "bg-surface border-border group-hover:border-border-light"}`}>
                  {acceptMarketing && (
                    <svg className="w-2.5 h-2.5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={3.5}>
                      <path strokeLinecap="round" strokeLinejoin="round" d="m4.5 12.75 6 6 9-13.5" />
                    </svg>
                  )}
                </span>
                <span className="text-xs text-text-secondary leading-relaxed">
                  I accept to receive news and updates from ShellHub via email
                </span>
              </button>
            </div>

            {/* Submit */}
            <button
              type="submit"
              disabled={loading}
              className="w-full px-6 py-3 text-sm font-semibold bg-primary border border-primary-400/40 text-white rounded-xl hover:brightness-110 hover:border-primary-400/60 hover:scale-[1.02] active:scale-[0.98] disabled:opacity-40 disabled:cursor-not-allowed transition-all duration-300 mt-1"
            >
              {loading ? (
                <span className="flex items-center justify-center gap-2">
                  <span className="w-3.5 h-3.5 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                  Creating account...
                </span>
              ) : "Create Account"}
            </button>

            {/* Back */}
            <button
              type="button"
              onClick={onBack}
              className="w-full inline-flex items-center justify-center gap-2 px-5 py-1.5 text-xs font-medium text-text-muted hover:text-text-secondary transition-colors group"
            >
              <svg className="w-3.5 h-3.5 group-hover:-translate-x-0.5 transition-transform duration-300" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2.5}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M10.5 19.5 3 12m0 0 7.5-7.5M3 12h18" />
              </svg>
              Back to options
            </button>
          </form>
        </div>
      </Reveal>
    </div>
  );
}
