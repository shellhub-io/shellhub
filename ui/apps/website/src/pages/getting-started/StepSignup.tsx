import { useState, FormEvent } from "react";
import { cn } from "@shellhub/design-system/cn";
import {
  UserIcon,
  EnvelopeIcon,
  LockClosedIcon,
  CheckIcon,
  ExclamationCircleIcon,
  EyeIcon,
  EyeSlashIcon,
  ArrowLeftIcon,
} from "@heroicons/react/24/outline";
import {
  Button,
  Card,
  ShellHubCloudIcon,
} from "@shellhub/design-system/primitives";
import { Reveal } from "../landing/components";
import apiClient from "@/api/client";
import { loginUrl } from "@/links";

interface StepSignupProps {
  onBack: () => void;
}

const INPUT_BASE =
  "w-full pl-10 pr-3.5 py-2.5 bg-surface border border-border rounded-lg text-sm text-text-primary placeholder:text-text-secondary focus:outline-none focus:border-primary/50 focus:ring-1 focus:ring-primary/20 transition-all";

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
    else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email))
      errors.email = "Invalid email";
    if (!password) errors.password = "Password is required";
    else if (password.length < 5) errors.password = "At least 5 characters";
    else if (password.length > 32) errors.password = "At most 32 characters";
    if (password !== confirmPassword)
      errors.confirmPassword = "Passwords do not match";
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
          <Card className="p-8 text-center">
            <div className="w-12 h-12 rounded-full bg-accent-green/10 border border-accent-green/20 flex items-center justify-center mx-auto mb-4">
              <CheckIcon
                className="w-6 h-6 text-accent-green"
                strokeWidth={2}
              />
            </div>
            <h3 className="text-lg font-bold mb-2">Account created</h3>
            <p className="text-sm text-text-secondary leading-relaxed mb-6">
              Check your email to confirm your account, then log in to get
              started.
            </p>
            <Button
              as="a"
              variant="primary"
              size="lg"
              fullWidth
              href={loginUrl}
            >
              Go to Login
            </Button>
          </Card>
        </Reveal>
      </div>
    );
  }

  return (
    <div className="max-w-md mx-auto w-full">
      <Reveal>
        <Card className="overflow-hidden">
          {/* Header */}
          <div className="px-6 pt-6 pb-5 border-b border-border bg-surface/50">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 rounded-xl bg-primary/10 border border-primary/20 flex items-center justify-center">
                <ShellHubCloudIcon className="h-5" />
              </div>
              <div>
                <h3 className="text-sm font-bold">ShellHub Cloud</h3>
                <div className="flex items-center gap-2 mt-0.5">
                  <span className="text-2xs text-text-muted">Free account</span>
                  <span className="inline-flex items-center gap-1 px-1.5 py-px text-2xs font-medium text-accent-green bg-accent-green/8 border border-accent-green/15 rounded">
                    <CheckIcon className="w-2.5 h-2.5" strokeWidth={2.5} />
                    No credit card required
                  </span>
                </div>
              </div>
            </div>
          </div>

          {/* Form */}
          <form
            onSubmit={(e) => {
              void handleSubmit(e);
            }}
            className="p-6 space-y-5"
          >
            {error && (
              <div className="flex items-center gap-2 bg-accent-red/8 border border-accent-red/20 text-accent-red px-3.5 py-2.5 rounded-md text-xs font-mono animate-slide-down">
                <ExclamationCircleIcon
                  className="w-3.5 h-3.5 shrink-0"
                  strokeWidth={2}
                />
                {error}
              </div>
            )}

            {/* Name */}
            <div>
              <label
                htmlFor="signup-name"
                className="block text-2xs font-mono font-semibold uppercase tracking-label text-text-muted mb-1.5"
              >
                Name
              </label>
              <div className="relative">
                <span className="absolute left-3 top-1/2 -translate-y-1/2 pointer-events-none">
                  <UserIcon
                    className="w-4 h-4 text-text-muted"
                    strokeWidth={1.5}
                  />
                </span>
                <input
                  id="signup-name"
                  type="text"
                  value={name}
                  onChange={(e) => {
                    setName(e.target.value);
                    setFieldErrors((p) => ({ ...p, name: "" }));
                  }}
                  placeholder="Your name"
                  className={INPUT_BASE}
                />
              </div>
              {fieldErrors.name && (
                <p className="mt-1 text-2xs text-accent-red">
                  {fieldErrors.name}
                </p>
              )}
            </div>

            {/* Email */}
            <div>
              <label
                htmlFor="signup-email"
                className="block text-2xs font-mono font-semibold uppercase tracking-label text-text-muted mb-1.5"
              >
                Email
              </label>
              <div className="relative">
                <span className="absolute left-3 top-1/2 -translate-y-1/2 pointer-events-none">
                  <EnvelopeIcon
                    className="w-4 h-4 text-text-muted"
                    strokeWidth={1.5}
                  />
                </span>
                <input
                  id="signup-email"
                  type="email"
                  value={email}
                  onChange={(e) => {
                    setEmail(e.target.value);
                    setFieldErrors((p) => ({ ...p, email: "" }));
                  }}
                  placeholder="you@example.com"
                  className={INPUT_BASE}
                />
              </div>
              {fieldErrors.email && (
                <p className="mt-1 text-2xs text-accent-red">
                  {fieldErrors.email}
                </p>
              )}
            </div>

            {/* Divider */}
            <div className="border-t border-border" />

            {/* Password */}
            <div>
              <label
                htmlFor="signup-password"
                className="block text-2xs font-mono font-semibold uppercase tracking-label text-text-muted mb-1.5"
              >
                Password
              </label>
              <div className="relative">
                <span className="absolute left-3 top-1/2 -translate-y-1/2 pointer-events-none">
                  <LockClosedIcon
                    className="w-4 h-4 text-text-muted"
                    strokeWidth={1.5}
                  />
                </span>
                <input
                  id="signup-password"
                  type={showPassword ? "text" : "password"}
                  value={password}
                  onChange={(e) => {
                    setPassword(e.target.value);
                    setFieldErrors((p) => ({
                      ...p,
                      password: "",
                      confirmPassword: "",
                    }));
                  }}
                  placeholder="Min. 5 chars"
                  className={cn(INPUT_BASE, "pr-9")}
                />
                <button
                  type="button"
                  onClick={() => setShowPassword(!showPassword)}
                  className="absolute right-2.5 top-1/2 -translate-y-1/2 text-text-muted hover:text-text-primary transition-colors"
                >
                  {showPassword ? (
                    <EyeSlashIcon className="w-3.5 h-3.5" strokeWidth={1.5} />
                  ) : (
                    <EyeIcon className="w-3.5 h-3.5" strokeWidth={1.5} />
                  )}
                </button>
              </div>
              {fieldErrors.password && (
                <p className="mt-1 text-2xs text-accent-red">
                  {fieldErrors.password}
                </p>
              )}
            </div>

            {/* Confirm password */}
            <div>
              <label
                htmlFor="signup-confirm-password"
                className="block text-2xs font-mono font-semibold uppercase tracking-label text-text-muted mb-1.5"
              >
                Confirm password
              </label>
              <div className="relative">
                <span className="absolute left-3 top-1/2 -translate-y-1/2 pointer-events-none">
                  <LockClosedIcon
                    className="w-4 h-4 text-text-muted"
                    strokeWidth={1.5}
                  />
                </span>
                <input
                  id="signup-confirm-password"
                  type={showPassword ? "text" : "password"}
                  value={confirmPassword}
                  onChange={(e) => {
                    setConfirmPassword(e.target.value);
                    setFieldErrors((p) => ({ ...p, confirmPassword: "" }));
                  }}
                  placeholder="Repeat your password"
                  className={INPUT_BASE}
                />
              </div>
              {fieldErrors.confirmPassword && (
                <p className="mt-1 text-2xs text-accent-red">
                  {fieldErrors.confirmPassword}
                </p>
              )}
            </div>

            {/* Agreements */}
            <div className="space-y-3 pt-1">
              <button
                type="button"
                onClick={() => {
                  setAcceptPrivacy(!acceptPrivacy);
                  setFieldErrors((p) => ({ ...p, privacy: "" }));
                }}
                className="flex items-start gap-3 w-full text-left group"
              >
                <span
                  className={cn("mt-px w-4 h-4 rounded flex items-center justify-center shrink-0 border transition-all duration-200", acceptPrivacy ? "bg-primary border-primary" : "bg-surface border-border group-hover:border-border-light")}
                >
                  {acceptPrivacy && (
                    <CheckIcon
                      className="w-2.5 h-2.5 text-white"
                      strokeWidth={3.5}
                    />
                  )}
                </span>
                <span className="text-xs text-text-secondary leading-relaxed">
                  I agree to the{" "}
                  <a
                    href="https://www.shellhub.io/privacy-policy"
                    target="_blank"
                    rel="noopener noreferrer"
                    onClick={(e) => e.stopPropagation()}
                    className="text-primary hover:underline"
                  >
                    Privacy Policy
                  </a>
                </span>
              </button>
              {fieldErrors.privacy && (
                <p className="ml-7 text-2xs text-accent-red">
                  {fieldErrors.privacy}
                </p>
              )}
              <button
                type="button"
                onClick={() => setAcceptMarketing(!acceptMarketing)}
                className="flex items-start gap-3 w-full text-left group"
              >
                <span
                  className={cn("mt-px w-4 h-4 rounded flex items-center justify-center shrink-0 border transition-all duration-200", acceptMarketing ? "bg-primary border-primary" : "bg-surface border-border group-hover:border-border-light")}
                >
                  {acceptMarketing && (
                    <CheckIcon
                      className="w-2.5 h-2.5 text-white"
                      strokeWidth={3.5}
                    />
                  )}
                </span>
                <span className="text-xs text-text-secondary leading-relaxed">
                  I accept to receive news and updates from ShellHub via email
                </span>
              </button>
            </div>

            {/* Submit */}
            <Button
              type="submit"
              variant="primary"
              size="lg"
              fullWidth
              loading={loading}
              className="mt-1"
            >
              {loading ? "Creating account..." : "Create Account"}
            </Button>

            {/* Back */}
            <Button
              variant="ghost"
              size="sm"
              fullWidth
              className="group"
              onClick={onBack}
              icon={
                <ArrowLeftIcon
                  className="w-3.5 h-3.5 group-hover:-translate-x-0.5 transition-transform duration-300"
                  strokeWidth={2.5}
                />
              }
            >
              Back to options
            </Button>
          </form>
        </Card>
      </Reveal>
    </div>
  );
}
