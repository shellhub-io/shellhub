import { useState, useEffect, FormEvent } from "react";
import { Link, useNavigate, useSearchParams } from "react-router-dom";
import { UserPlusIcon } from "@heroicons/react/24/outline";
import { useForm } from "react-hook-form";
import { signUpResolver } from "./setup/signUpResolver";
import type { SignUpFormValues } from "./setup/signUpResolver";
import { useSignUpStore } from "../stores/signUpStore";
import AccountCreated from "../components/auth/AccountCreated";
import { Button, Callout } from "@shellhub/design-system/primitives";
import {
  FormInputField,
  FormPasswordField,
  FormCheckboxField,
} from "@/components/common/fields/rhf";
import CheckboxField from "@/components/common/fields/CheckboxField";

const SERVER_FIELD_MAP: Record<string, keyof SignUpFormValues> = {
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

  // acceptMarketing is optional and not part of the validated form values
  const [acceptMarketing, setAcceptMarketing] = useState(false);
  const [accountCreated, setAccountCreated] = useState(false);

  const { control, handleSubmit, setError, formState } =
    useForm<SignUpFormValues>({
      resolver: signUpResolver,
      mode: "onTouched",
      defaultValues: {
        name: "",
        username: "",
        email: emailFromQuery,
        password: "",
        confirmPassword: "",
        acceptPrivacyPolicy: false,
      },
    });

  // Sync server-side field errors into RHF field state whenever the store changes
  useEffect(() => {
    for (const field of signUpServerFields) {
      const key = SERVER_FIELD_MAP[field];
      const message = SERVER_FIELD_MESSAGES[field];

      if (key && message) {
        setError(key, { type: "server", message });
      }
    }
  }, [signUpServerFields, setError]);

  const onSubmit = async (values: SignUpFormValues) => {
    // Guard: still block if server field errors haven't been cleared yet
    if (signUpServerFields.length > 0) return;

    resetSignUpErrors();

    const token = await signUp({
      name: values.name,
      email: values.email,
      username: values.username,
      password: values.password,
      email_marketing: acceptMarketing,
      ...(sigFromQuery ? { sig: sigFromQuery } : {}),
    });

    // signUp absorbed any errors into the store; bail out if errors were set
    const { signUpError: err, signUpServerFields: fields } =
      useSignUpStore.getState();

    if (err !== null || fields.length > 0) return;

    if (!token) {
      void navigate(
        `/confirm-account?username=${encodeURIComponent(values.username)}`,
      );
      return;
    }

    // Invite flow: token returned — show AccountCreated
    setAccountCreated(true);
  };

  const handleFormSubmit = (e: FormEvent) => {
    void handleSubmit(onSubmit)(e);
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
          <Callout variant="warning">
            Please create your account before accepting the namespace
            invitation.
          </Callout>
        )}

        {/* Form card */}
        <div
          className="bg-card/80 border border-border rounded-2xl p-8 backdrop-blur-sm animate-slide-up"
          style={{ animationDelay: "150ms" }}
        >
          {signUpError && (
            <Callout variant="error" className="mb-5">
              {signUpError}
            </Callout>
          )}

          <form
            onSubmit={handleFormSubmit}
            className="space-y-4"
            aria-label="Create account"
          >
            <FormInputField<SignUpFormValues>
              id="name"
              label="Name"
              name="name"
              control={control}
              placeholder="Your name"
              autoComplete="name"
              onValueChange={() => clearSignUpServerField("name")}
            />

            <FormInputField<SignUpFormValues>
              id="username"
              label="Username"
              name="username"
              control={control}
              placeholder="username"
              autoComplete="username"
              onValueChange={() => clearSignUpServerField("username")}
            />

            <FormInputField<SignUpFormValues>
              id="email"
              label="Email"
              name="email"
              control={control}
              type="email"
              placeholder="you@example.com"
              autoComplete="email"
              disabled={isInvite}
              onValueChange={() => clearSignUpServerField("email")}
            />

            <FormPasswordField<SignUpFormValues>
              id="password"
              label="Password"
              name="password"
              control={control}
              placeholder="Min. 5 characters"
              onValueChange={() => clearSignUpServerField("password")}
            />

            <FormPasswordField<SignUpFormValues>
              id="confirmPassword"
              label="Confirm Password"
              name="confirmPassword"
              control={control}
              placeholder="Re-enter password"
            />

            {/* Privacy Policy checkbox (required) */}
            <FormCheckboxField<SignUpFormValues>
              id="signup-accept-privacy"
              name="acceptPrivacyPolicy"
              control={control}
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

            {/* Marketing checkbox (optional) — not validated, kept as local state */}
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
              disabled={
                signUpLoading ||
                !formState.isValid ||
                signUpServerFields.length > 0
              }
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
