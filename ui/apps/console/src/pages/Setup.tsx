import { useState, useEffect, useCallback, FormEvent } from "react";
import { useForm, useWatch } from "react-hook-form";
import { isSdkError } from "../api/errors";
import { useNavigate } from "react-router-dom";
import {
  CheckIcon,
  ExclamationCircleIcon,
  PencilSquareIcon,
} from "@heroicons/react/24/outline";
import { setup } from "../client";
import { getConfig } from "../env";
import { useAuthStore } from "@/stores/authStore";
import { setupResolver, type SetupFormValues } from "./setup/setupResolver";
import { suggestNamespace } from "./setup/validate";
import {
  FormInputField,
  FormPasswordField,
} from "@/components/common/fields/rhf";
import { Button, ShellHubLogo } from "@shellhub/design-system/primitives";
import { cn } from "@shellhub/design-system/cn";

const STEP_ONBOARDING = 1;
const STEP_ACCOUNT = 2;

export default function Setup() {
  const navigate = useNavigate();
  const config = getConfig();
  const loginWithToken = useAuthStore((state) => state.loginWithToken);

  const isCommunity = !config.cloud && !config.enterprise;
  const showOnboarding = isCommunity && !!config.onboardingUrl;

  const [step, setStep] = useState(
    showOnboarding ? STEP_ONBOARDING : STEP_ACCOUNT,
  );
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState(false);
  const [surveyCompleted, setSurveyCompleted] = useState(false);

  const { control, handleSubmit, formState, setValue } =
    useForm<SetupFormValues>({
      resolver: setupResolver,
      mode: "onTouched",
      defaultValues: {
        name: "",
        username: "",
        // In development the namespace defaults to "dev" (and does not track the username) so the
        // instance binds to the well-known dev tenant/fixtures.
        namespace: import.meta.env.DEV ? "dev" : "",
        email: "",
        password: "",
        confirmPassword: "",
      },
    });

  // The namespace defaults to a slug of the username and stays in sync until the user opts to
  // edit it (readonly + Edit button), so setup has a sensible name without an extra decision.
  const [namespaceEdited, setNamespaceEdited] = useState(false);
  const usernameValue = useWatch({ control, name: "username" });
  const namespaceValue = useWatch({ control, name: "namespace" });

  useEffect(() => {
    // Skip the username-driven suggestion in development, where it stays fixed at "dev".
    if (!import.meta.env.DEV && !namespaceEdited) {
      setValue("namespace", suggestNamespace(usernameValue ?? ""), {
        shouldValidate: true,
      });
    }
  }, [usernameValue, namespaceEdited, setValue]);

  const disableCreateAccountButton = loading || !formState.isValid;

  const onboardingUrl = (() => {
    if (!config.onboardingUrl) return "";
    const params = new URLSearchParams({
      consent_to_contact: "accepted",
      source: "self-hosted",
      embed: "true",
      instance_domain: window.location.hostname,
    });
    if (import.meta.env.DEV) params.append("preview", "true");
    return `${config.onboardingUrl}?${params.toString()}`;
  })();

  const handleMessage = useCallback(
    (event: MessageEvent) => {
      if (!config.onboardingUrl) return;
      try {
        const origin = new URL(config.onboardingUrl).origin;
        if (event.origin !== origin) return;
      } catch {
        return;
      }
      if (event.data === "formbricksSurveyCompleted") {
        setSurveyCompleted(true);
      }
    },
    [config.onboardingUrl],
  );

  useEffect(() => {
    if (!showOnboarding) return;
    window.addEventListener("message", handleMessage);
    return () => window.removeEventListener("message", handleMessage);
  }, [showOnboarding, handleMessage]);

  useEffect(() => {
    if (success) {
      // Setup already authenticated us (auto-login), so land directly on the app
      // instead of bouncing through the login screen.
      const timer = setTimeout(
        () => void navigate("/", { replace: true }),
        3000,
      );
      return () => clearTimeout(timer);
    }
  }, [success, navigate]);

  const onSubmit = async (values: SetupFormValues) => {
    setLoading(true);
    setError("");

    let token: string | undefined;
    try {
      const { data } = await setup({
        body: {
          name: values.name,
          username: values.username,
          namespace: values.namespace,
          email: values.email,
          password: values.password,
        },
        throwOnError: true,
      });
      token = data.token;
    } catch (err: unknown) {
      setError(
        isSdkError(err) && err.status === 409
          ? "Setup has already been completed."
          : "An error occurred. Please try again.",
      );
      setLoading(false);
      return;
    }

    // Setup is committed at this point. Try to enter the app directly with the session it
    // issued; if the auto-login fails (or no token was issued), setup is still done — route to
    // the login screen with a notice instead of surfacing a misleading "setup failed" error
    // that a retry would only turn into a 409.
    try {
      if (!token) throw new Error("no session issued");
      await loginWithToken(token);
      setSuccess(true);
    } catch {
      void navigate("/login", {
        replace: true,
        state: { notice: "Setup complete. Please sign in." },
      });
    } finally {
      setLoading(false);
    }
  };

  const handleFormSubmit = (e: FormEvent) => {
    void handleSubmit(onSubmit)(e);
  };

  const totalSteps = showOnboarding ? 2 : 1;
  const displayStep = step === STEP_ONBOARDING ? 1 : totalSteps;

  if (success) {
    return (
      <div className="w-full max-w-sm mx-auto animate-fade-in">
        <div className="bg-surface border border-border rounded-lg overflow-hidden">
          <div className="px-8 pt-8 pb-6 border-b border-border bg-card/50">
            <div className="flex justify-center mb-5">
              <ShellHubLogo className="h-7" />
            </div>
            <p className="text-center text-2xs font-mono text-text-muted tracking-wider uppercase">
              Initial Setup
            </p>
          </div>

          <div className="p-8 text-center">
            <div className="inline-flex items-center justify-center w-12 h-12 rounded-full bg-accent-green/10 mb-4">
              <CheckIcon
                className="w-6 h-6 text-accent-green"
                strokeWidth={2}
              />
            </div>
            <h3 className="text-sm font-semibold text-text-primary mb-2">
              Instance ready
            </h3>
            <p className="text-xs text-text-secondary leading-relaxed">
              Taking you to your instance...
            </p>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div
      className={cn("w-full mx-auto animate-fade-in", step === STEP_ONBOARDING ? "max-w-lg" : "max-w-sm")}
    >
      <div className="bg-surface border border-border rounded-lg overflow-hidden">
        <div className="px-8 pt-8 pb-6 border-b border-border bg-card/50">
          <div className="flex justify-center mb-5">
            <ShellHubLogo className="h-7" />
          </div>
          <h1 className="text-center text-sm font-semibold text-text-primary mb-1">
            Welcome to ShellHub
          </h1>
          <p className="text-center text-2xs font-mono text-text-muted tracking-wider uppercase">
            Initial Setup
          </p>

          {showOnboarding && (
            <div className="flex items-center justify-center gap-2 mt-4">
              {Array.from({ length: totalSteps }, (_, i) => (
                <StepIndicator key={i} index={i} current={displayStep} />
              ))}
            </div>
          )}
        </div>

        <div className="p-8">
          {error && (
            <div className="flex items-start gap-2 bg-accent-red/8 border border-accent-red/20 text-accent-red px-3.5 py-2.5 rounded-md text-xs font-mono animate-slide-down mb-5">
              <ExclamationCircleIcon
                className="w-3.5 h-3.5 shrink-0 mt-0.5"
                strokeWidth={2}
              />
              {error}
            </div>
          )}

          {step === STEP_ONBOARDING && (
            <div className="space-y-5">
              <p className="text-xs text-text-secondary leading-relaxed text-center">
                Help us improve ShellHub by sharing your feedback
              </p>

              <div className="relative h-[60dvh] overflow-auto rounded-md border border-border">
                <iframe
                  src={onboardingUrl}
                  title="Onboarding survey"
                  className="absolute inset-0 w-full h-full border-0"
                />
              </div>

              <Button
                fullWidth
                disabled={!surveyCompleted}
                onClick={() => setStep(STEP_ACCOUNT)}
              >
                Continue
              </Button>

              {import.meta.env.DEV && (
                <button
                  type="button"
                  onClick={() => setStep(STEP_ACCOUNT)}
                  className="w-full text-center text-2xs font-mono text-text-muted hover:text-text-secondary transition-colors"
                >
                  Skip survey (dev only)
                </button>
              )}
            </div>
          )}

          {step === STEP_ACCOUNT && (
            <form onSubmit={handleFormSubmit} className="space-y-4">
              <p className="text-xs text-text-secondary leading-relaxed mb-1">
                Set up your ShellHub instance.
              </p>

              <FormInputField<SetupFormValues>
                id="name"
                label="Name"
                name="name"
                control={control}
                placeholder="Your name"
                maxLength={64}
              />

              <FormInputField<SetupFormValues>
                id="username"
                label="Username"
                name="username"
                control={control}
                placeholder="username"
                maxLength={32}
              />

              <FormInputField<SetupFormValues>
                id="email"
                label="Email"
                name="email"
                control={control}
                type="email"
                placeholder="you@example.com"
              />

              <FormInputField<SetupFormValues>
                id="namespace"
                label="Namespace"
                name="namespace"
                control={control}
                variant="mono"
                maxLength={30}
                readOnly={!namespaceEdited}
                // Suppress the error only while the field is pristine/empty. If a non-empty
                // auto-suggestion is invalid (e.g. a username that slugs too short), let the
                // error show so the user knows why submit is disabled and can hit Edit to fix it.
                error={!namespaceEdited && !namespaceValue ? "" : undefined}
                hint={
                  import.meta.env.DEV
                    ? 'Keeping "dev" binds the well-known dev tenant; any other name generates a fresh one.'
                    : undefined
                }
                labelAdornment={
                  !namespaceEdited && (
                    <button
                      type="button"
                      onClick={() => setNamespaceEdited(true)}
                      className="inline-flex items-center gap-1 text-2xs font-medium text-primary hover:text-primary-300 transition-colors"
                    >
                      <PencilSquareIcon className="w-3 h-3" strokeWidth={2} />
                      Edit
                    </button>
                  )
                }
              />

              <FormPasswordField<SetupFormValues>
                id="password"
                label="Password"
                name="password"
                control={control}
                placeholder="Min. 5 characters"
              />

              <FormPasswordField<SetupFormValues>
                id="confirmPassword"
                label="Confirm Password"
                name="confirmPassword"
                control={control}
                placeholder="Re-enter password"
              />

              <div className="flex gap-3 pt-1">
                {showOnboarding && (
                  <Button
                    variant="secondary"
                    className="flex-1"
                    onClick={() => setStep(STEP_ONBOARDING)}
                  >
                    Back
                  </Button>
                )}
                <Button
                  type="submit"
                  loading={loading}
                  disabled={disableCreateAccountButton}
                  className={showOnboarding ? "flex-[2]" : "w-full"}
                >
                  {loading ? "Setting up..." : "Complete setup"}
                </Button>
              </div>
            </form>
          )}
        </div>
      </div>

      <p className="text-center text-2xs font-mono text-text-muted/40 mt-6">
        ShellHub &mdash; Secure Remote Access
      </p>
    </div>
  );
}

function StepIndicator({ index, current }: { index: number; current: number }) {
  const stepNum = index + 1;
  return (
    <>
      {index > 0 && <div className="w-6 h-px bg-border" />}
      <StepDot active={current === stepNum} label={String(stepNum)} />
    </>
  );
}

function StepDot({ active, label }: { active: boolean; label: string }) {
  return (
    <div
      className={cn(
        "w-5 h-5 rounded-full flex items-center justify-center text-3xs font-mono font-bold transition-colors",
        active ? "bg-primary text-white" : "bg-border text-text-muted",
      )}
    >
      {label}
    </div>
  );
}
