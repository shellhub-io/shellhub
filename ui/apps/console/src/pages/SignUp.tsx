import { useState, useMemo, FormEvent, useEffect } from "react";
import { Link, useNavigate, useSearchParams } from "react-router-dom";
import { UserPlusIcon } from "@heroicons/react/24/outline";
import Alert from "@/components/common/Alert";
import { validate, type FormErrors } from "./setup/validate";
import { useSignUpStore } from "../stores/signUpStore";
import AccountCreated from "../components/auth/AccountCreated";
import { Button } from "@shellhub/design-system/primitives";
import InputField from "@/components/common/fields/InputField";
import PasswordField from "@/components/common/fields/PasswordField";
import CheckboxField from "@/components/common/fields/CheckboxField";

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
  const clearSignUpServerField = useSignUpStore(
    (s) => s.clearSignUpServerField,
  );
  const resetSignUpErrors = useSignUpStore((s) => s.resetSignUpErrors);

  useEffect(() => {
    resetSignUpErrors();
  }, [resetSignUpErrors]);

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
    const errors = validate({
      name,
      username,
      email,
      password,
      confirmPassword,
    });
    if (
      Object.keys(errors).length > 0 ||
      !acceptPrivacyPolicy ||
      signUpServerFields.length > 0
    )
      return;

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
    const { signUpError: err, signUpServerFields: fields } =
      useSignUpStore.getState();
    if (err !== null || fields.length > 0) return;

    if (!token) {
      void navigate(
        `/confirm-account?username=${encodeURIComponent(username)}`,
      );
      return;
    }

    // Invite flow: token returned — show AccountCreated
    setAccountCreated(true);
  };

  if (accountCreated) {
    return (
      <div className="w-full max-w-5xl mx-auto px-8 py-12 flex flex-col items-center">
        <AccountCreated />
      </div>
    );
  }

  return (
    <div className="w-full max-w-5xl mx-auto px-8 py-12 flex flex-col items-center">
      {/* Hero */}
      <div className="text-center mb-10 animate-fade-in">
        <div className="animate-float mb-6 inline-block">
          <div className="w-20 h-20 rounded-2xl bg-primary/15 border border-primary/25 flex items-center justify-center shadow-lg shadow-primary/10">
            <UserPlusIcon
              className="w-10 h-10 text-primary"
              strokeWidth={1.2}
            />
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
          <Alert variant="warning">
            Please create your account before accepting the namespace
            invitation.
          </Alert>
        )}

        {/* Form card */}
        <div
          className="bg-card/80 border border-border rounded-2xl p-8 backdrop-blur-sm animate-slide-up"
          style={{ animationDelay: "150ms" }}
        >
          {signUpError && (
            <Alert variant="error" className="mb-5">
              {signUpError}
            </Alert>
          )}

          <form
            onSubmit={(e) => void handleSubmit(e)}
            className="space-y-4"
            aria-label="Create account"
          >
            <InputField
              id="name"
              label="Name"
              value={name}
              onChange={(v) => {
                setName(v);
                clearSignUpServerField("name");
              }}
              onBlur={() => handleBlur("name")}
              error={fieldError("name")}
              placeholder="Your name"
              autoComplete="name"

            />

            <InputField
              id="username"
              label="Username"
              value={username}
              onChange={(v) => {
                setUsername(v);
                clearSignUpServerField("username");
              }}
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
              onChange={(v) => {
                setEmail(v);
                clearSignUpServerField("email");
              }}
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
              onChange={(v) => {
                setPassword(v);
                clearSignUpServerField("password");
              }}
              onBlur={() => handleBlur("password")}
              error={fieldError("password")}
              placeholder="Min. 5 characters"
            />

            <PasswordField
              id="confirmPassword"
              label="Confirm Password"
              value={confirmPassword}
              onChange={setConfirmPassword}
              onBlur={() => handleBlur("confirmPassword")}
              error={fieldError("confirmPassword")}
              placeholder="Re-enter password"
            />

            {/* Privacy Policy checkbox (required) */}
            <CheckboxField
              id="signup-accept-privacy"
              checked={acceptPrivacyPolicy}
              onChange={setAcceptPrivacyPolicy}
              required
              label={
                <>
                  I agree to the{" "}
                  <a
                    href="https://www.shellhub.io/privacy-policy"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-primary hover:text-primary/80 underline transition-colors"
                  >
                    Privacy Policy
                  </a>
                </>
              }
            />

            {/* Marketing checkbox (optional) */}
            <CheckboxField
              id="signup-accept-marketing"
              checked={acceptMarketing}
              onChange={setAcceptMarketing}
              label="I accept to receive news and updates from ShellHub via email."
            />

            <Button
              variant="primary"
              size="lg"
              fullWidth
              type="submit"
              className="px-4"
              loading={signUpLoading}
              disabled={signUpLoading || !isFormValid}
            >
              {signUpLoading ? "Creating account..." : "Create Account"}
            </Button>
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
