import { useState, useMemo, FormEvent } from "react";
import { Link, useNavigate, useSearchParams } from "react-router-dom";
import {
  UserPlusIcon,
  EyeIcon,
  EyeSlashIcon,
  ExclamationCircleIcon,
  ExclamationTriangleIcon,
} from "@heroicons/react/24/outline";
import { validate, type FormErrors } from "./setup/validate";
import { useSignUpStore } from "../stores/signUpStore";
import AccountCreated from "../components/auth/AccountCreated";

const SERVER_FIELD_MAP: Record<string, keyof FormErrors> = {
  username: "username",
  email: "email",
  name: "name",
  password: "password",
};

const SERVER_FIELD_MESSAGES: Record<string, string> = {
  username: "This username already exists",
  email: "This email is invalid or already in use",
  name: "This name is invalid",
  password: "This password is invalid",
};

export default function SignUp() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const signUp = useSignUpStore((s) => s.signUp);
  const signUpLoading = useSignUpStore((s) => s.signUpLoading);
  const signUpError = useSignUpStore((s) => s.signUpError);
  const signUpServerFields = useSignUpStore((s) => s.signUpServerFields);
  const clearSignUpServerField = useSignUpStore((s) => s.clearSignUpServerField);
  const resetSignUpErrors = useSignUpStore((s) => s.resetSignUpErrors);

  const emailFromQuery = searchParams.get("email") ?? "";
  const sigFromQuery = searchParams.get("sig") ?? "";
  const isInvite = Boolean(emailFromQuery && sigFromQuery);

  const [name, setName] = useState("");
  const [username, setUsername] = useState("");
  const [email, setEmail] = useState(emailFromQuery);
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [acceptPrivacyPolicy, setAcceptPrivacyPolicy] = useState(false);
  const [acceptMarketing, setAcceptMarketing] = useState(false);
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirm, setShowConfirm] = useState(false);
  const [touched, setTouched] = useState<Record<string, boolean>>({});
  const [accountCreated, setAccountCreated] = useState(false);

  const serverFieldErrors = useMemo(() => {
    const mapped: Partial<FormErrors> = {};
    for (const field of signUpServerFields) {
      const key = SERVER_FIELD_MAP[field];
      if (key) mapped[key] = SERVER_FIELD_MESSAGES[field];
    }
    return mapped;
  }, [signUpServerFields]);

  const validationErrors = useMemo(
    () => validate({ name, username, email, password, confirmPassword }),
    [name, username, email, password, confirmPassword],
  );

  const fieldError = (field: keyof FormErrors): string | undefined => {
    if (serverFieldErrors[field]) return serverFieldErrors[field];
    return touched[field] ? validationErrors[field] : undefined;
  };

  const handleBlur = (field: string) =>
    setTouched((prev) => ({ ...prev, [field]: true }));

  const isFormValid = useMemo(
    () =>
      Object.keys(validationErrors).length === 0 &&
      !Object.values(serverFieldErrors).some(Boolean) &&
      acceptPrivacyPolicy,
    [validationErrors, serverFieldErrors, acceptPrivacyPolicy],
  );

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();

    setTouched({
      name: true,
      username: true,
      email: true,
      password: true,
      confirmPassword: true,
    });

    // Compute validity from source of truth rather than the memoized `isFormValid`,
    // which reflects the previous render and would be stale after `setTouched`.
    const errors = validate({ name, username, email, password, confirmPassword });
    if (Object.keys(errors).length > 0 || !acceptPrivacyPolicy || signUpServerFields.length > 0) return;

    resetSignUpErrors();

    const token = await signUp({
      name,
      email,
      username,
      password,
      email_marketing: acceptMarketing,
      ...(sigFromQuery ? { sig: sigFromQuery } : {}),
    });

    // signUp absorbed any errors into the store; bail out if errors were set
    const { signUpError: err, signUpServerFields: fields } = useSignUpStore.getState();
    if (err !== null || fields.length > 0) return;

    if (!token) {
      navigate(`/confirm-account?username=${encodeURIComponent(username)}`);
      return;
    }

    // Invite flow: token returned — show AccountCreated
    setAccountCreated(true);
  };

  if (accountCreated) {
    return (
      <div className="w-full max-w-5xl mx-auto px-8 py-12 flex flex-col items-center">
        <AccountCreated mode="sig" username={username} />
      </div>
    );
  }

  return (
    <div className="w-full max-w-5xl mx-auto px-8 py-12 flex flex-col items-center">
      {/* Hero */}
      <div className="text-center mb-10 animate-fade-in">
        <div className="animate-float mb-6 inline-block">
          <div className="w-20 h-20 rounded-2xl bg-primary/15 border border-primary/25 flex items-center justify-center shadow-lg shadow-primary/10">
            <UserPlusIcon className="w-10 h-10 text-primary" strokeWidth={1.2} />
          </div>
        </div>

        <p className="text-2xs font-mono font-semibold uppercase tracking-wide text-primary/80 mb-2">
          Get Started
        </p>
        <h1 className="text-3xl font-bold text-text-primary mb-3">
          Create your account
        </h1>
        <p className="text-sm text-text-muted max-w-md mx-auto leading-relaxed">
          Register to start managing your devices securely through ShellHub.
        </p>
      </div>

      <div className="w-full max-w-sm space-y-4">
        {/* Invite alert */}
        {isInvite && (
          <div
            className="flex items-start gap-2.5 bg-accent-yellow/8 border border-accent-yellow/20 text-accent-yellow px-3.5 py-3 rounded-lg text-xs font-mono animate-slide-down"
            role="alert"
          >
            <ExclamationTriangleIcon className="w-4 h-4 shrink-0 mt-0.5" strokeWidth={2} />
            Please create your account before accepting the namespace invitation.
          </div>
        )}

        {/* Form card */}
        <div
          className="bg-card/80 border border-border rounded-2xl p-8 backdrop-blur-sm animate-slide-up"
          style={{ animationDelay: "150ms" }}
        >
          {signUpError && (
            <div role="alert" className="flex items-center gap-2 bg-accent-red/8 border border-accent-red/20 text-accent-red px-3.5 py-2.5 rounded-md text-xs font-mono animate-slide-down mb-5">
              <ExclamationCircleIcon className="w-3.5 h-3.5 shrink-0" strokeWidth={2} />
              {signUpError}
            </div>
          )}

          <form onSubmit={handleSubmit} className="space-y-4" aria-label="Create account">
            <InputField
              id="name"
              label="Name"
              value={name}
              onChange={(v) => { setName(v); clearSignUpServerField("name"); }}
              onBlur={() => handleBlur("name")}
              error={fieldError("name")}
              placeholder="Your name"
              autoComplete="name"
              autoFocus
            />

            <InputField
              id="username"
              label="Username"
              value={username}
              onChange={(v) => { setUsername(v); clearSignUpServerField("username"); }}
              onBlur={() => handleBlur("username")}
              error={fieldError("username")}
              placeholder="username"
              autoComplete="username"
            />

            <InputField
              id="email"
              label="Email"
              type="email"
              value={email}
              onChange={(v) => { setEmail(v); clearSignUpServerField("email"); }}
              onBlur={() => handleBlur("email")}
              error={fieldError("email")}
              placeholder="you@example.com"
              autoComplete="email"
              disabled={isInvite}
            />

            <PasswordField
              id="password"
              label="Password"
              value={password}
              onChange={(v) => { setPassword(v); clearSignUpServerField("password"); }}
              onBlur={() => handleBlur("password")}
              error={fieldError("password")}
              placeholder="Min. 5 characters"
              visible={showPassword}
              onToggle={() => setShowPassword((v) => !v)}
            />

            <PasswordField
              id="confirmPassword"
              label="Confirm Password"
              value={confirmPassword}
              onChange={setConfirmPassword}
              onBlur={() => handleBlur("confirmPassword")}
              error={fieldError("confirmPassword")}
              placeholder="Re-enter password"
              visible={showConfirm}
              onToggle={() => setShowConfirm((v) => !v)}
            />

            {/* Privacy Policy checkbox (required) */}
            <Checkbox
              checked={acceptPrivacyPolicy}
              onChange={setAcceptPrivacyPolicy}
              required
            >
              I agree to the{" "}
              <a
                href="https://www.shellhub.io/privacy-policy"
                target="_blank"
                rel="noopener noreferrer"
                className="text-primary hover:text-primary/80 underline transition-colors"
              >
                Privacy Policy
              </a>
            </Checkbox>

            {/* Marketing checkbox (optional) */}
            <Checkbox
              checked={acceptMarketing}
              onChange={setAcceptMarketing}
            >
              I accept to receive news and updates from ShellHub via email.
            </Checkbox>

            <button
              type="submit"
              disabled={signUpLoading || !isFormValid}
              className="w-full bg-primary hover:bg-primary/90 text-white py-3 px-4 rounded-lg text-sm font-semibold disabled:opacity-dim disabled:cursor-not-allowed transition-all duration-200 mt-2"
            >
              {signUpLoading ? (
                <span className="flex items-center justify-center gap-2">
                  <span className="w-3.5 h-3.5 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                  <span className="font-mono text-xs">Creating account...</span>
                </span>
              ) : (
                "Create Account"
              )}
            </button>
          </form>
        </div>

        {/* Footer link */}
        <div
          className="flex items-center justify-center gap-1.5 text-xs text-text-muted animate-fade-in"
          style={{ animationDelay: "400ms" }}
        >
          Already have an account?
          <Link
            to="/login"
            className="text-primary hover:text-primary/80 font-medium transition-colors"
          >
            Sign In
          </Link>
        </div>
      </div>
    </div>
  );
}

// ─── Local sub-components ───────────────────────────────────────────────────────────

function InputField({
  id,
  label,
  type = "text",
  value,
  onChange,
  onBlur,
  error,
  placeholder,
  autoComplete,
  autoFocus,
  disabled,
}: {
  id: string;
  label: string;
  type?: string;
  value: string;
  onChange: (v: string) => void;
  onBlur: () => void;
  error?: string;
  placeholder: string;
  autoComplete?: string;
  autoFocus?: boolean;
  disabled?: boolean;
}) {
  return (
    <div>
      <label
        htmlFor={id}
        className="block text-2xs font-mono font-semibold uppercase tracking-label text-text-muted mb-2"
      >
        {label}
      </label>
      <input
        id={id}
        type={type}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        onBlur={onBlur}
        autoFocus={autoFocus}
        autoComplete={autoComplete}
        disabled={disabled}
        aria-invalid={error ? true : undefined}
        aria-describedby={error ? `${id}-error` : undefined}
        className={`w-full px-3.5 py-2.5 bg-card border rounded-md text-sm text-text-primary font-mono placeholder:text-text-secondary focus:outline-none focus:border-primary/50 focus:ring-1 focus:ring-primary/20 transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed ${
          error ? "border-accent-red/50" : "border-border"
        }`}
        placeholder={placeholder}
      />
      {error && (
        <p id={`${id}-error`} className="text-2xs font-mono text-accent-red mt-1.5">
          {error}
        </p>
      )}
    </div>
  );
}

function PasswordField({
  id,
  label,
  value,
  onChange,
  onBlur,
  error,
  placeholder,
  visible,
  onToggle,
}: {
  id: string;
  label: string;
  value: string;
  onChange: (v: string) => void;
  onBlur: () => void;
  error?: string;
  placeholder: string;
  visible: boolean;
  onToggle: () => void;
}) {
  return (
    <div>
      <label
        htmlFor={id}
        className="block text-2xs font-mono font-semibold uppercase tracking-label text-text-muted mb-2"
      >
        {label}
      </label>
      <div className="relative">
        <input
          id={id}
          type={visible ? "text" : "password"}
          value={value}
          onChange={(e) => onChange(e.target.value)}
          onBlur={onBlur}
          autoComplete="new-password"
          aria-invalid={error ? true : undefined}
          aria-describedby={error ? `${id}-error` : undefined}
          className={`w-full px-3.5 py-2.5 pr-10 bg-card border rounded-md text-sm text-text-primary font-mono placeholder:text-text-secondary focus:outline-none focus:border-primary/50 focus:ring-1 focus:ring-primary/20 transition-all duration-200 ${
            error ? "border-accent-red/50" : "border-border"
          }`}
          placeholder={placeholder}
        />
        <button
          type="button"
          onClick={onToggle}
          aria-label={visible ? "Hide password" : "Show password"}
          className="absolute right-2.5 top-1/2 -translate-y-1/2 text-text-muted hover:text-text-secondary transition-colors"
          tabIndex={-1}
        >
          {visible ? (
            <EyeSlashIcon className="w-4 h-4" />
          ) : (
            <EyeIcon className="w-4 h-4" />
          )}
        </button>
      </div>
      {error && (
        <p id={`${id}-error`} className="text-2xs font-mono text-accent-red mt-1.5">
          {error}
        </p>
      )}
    </div>
  );
}

function Checkbox({
  checked,
  onChange,
  required,
  children,
}: {
  checked: boolean;
  onChange: (v: boolean) => void;
  required?: boolean;
  children: React.ReactNode;
}) {
  return (
    <label className="flex items-start gap-3 cursor-pointer select-none pt-1">
      <div className="relative mt-0.5 shrink-0">
        <input
          type="checkbox"
          checked={checked}
          onChange={(e) => onChange(e.target.checked)}
          aria-required={required ? "true" : undefined}
          className="sr-only peer"
        />
        <div className="w-4 h-4 rounded border border-border peer-checked:bg-primary peer-checked:border-primary peer-focus-visible:ring-2 peer-focus-visible:ring-primary/30 transition-colors" />
        <svg
          className="absolute inset-0 w-4 h-4 text-white opacity-0 peer-checked:opacity-100 transition-opacity pointer-events-none"
          viewBox="0 0 16 16"
          fill="none"
        >
          <path d="M3 8l3.5 3.5 6.5-7" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
        </svg>
      </div>
      <span className="text-xs text-text-secondary leading-relaxed">{children}</span>
    </label>
  );
}
