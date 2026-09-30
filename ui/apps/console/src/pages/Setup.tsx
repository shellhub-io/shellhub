import { useState, useEffect, useCallback, FormEvent } from "react";
import { useForm, useWatch } from "react-hook-form";
import { useNavigate } from "react-router-dom";
import {
  ExclamationCircleIcon,
  PencilSquareIcon,
} from "@heroicons/react/24/outline";
import { Button } from "@shellhub/design-system/primitives";
import { isSdkError } from "@/api/errors";
import { setup } from "@/client";
import { getConfig, isCommunity } from "@/env";
import { useAuthStore } from "@/stores/authStore";
import {
  FormInputField,
  FormPasswordField,
} from "@/components/common/fields/rhf";
import FirstRunLayout from "@/components/firstRun/FirstRunLayout";
import {
  SETUP_STEP_TITLES,
  Trail,
  TrailStep,
  UpcomingDeviceSteps,
} from "@/components/firstRun/Trail";
import { firstRunEntryState } from "@/components/firstRun/entry";
import { setupResolver, type SetupFormValues } from "./setup/setupResolver";
import { suggestNamespace } from "./setup/validate";

const STEP_ONBOARDING = 1;
const STEP_ACCOUNT = 2;

/**
 * First-run setup for a fresh instance, as the opening steps of the first-run trail: the optional
 * survey on community, then the first account, which becomes the instance's admin, with its
 * namespace. It signs that account straight in, since there is nobody yet to sign in as, and hands
 * over to the dashboard, where the trail goes on to the first device.
 */
export default function Setup() {
  const navigate = useNavigate();
  const config = getConfig();
  const loginWithToken = useAuthStore((state) => state.loginWithToken);

  const showOnboarding = isCommunity() && !!config.onboardingUrl;

  const [step, setStep] = useState(
    showOnboarding ? STEP_ONBOARDING : STEP_ACCOUNT,
  );
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [surveyCompleted, setSurveyCompleted] = useState(false);

  const { control, handleSubmit, formState, setValue } =
    useForm<SetupFormValues>({
      resolver: setupResolver,
      mode: "onTouched",
      defaultValues: {
        name: "",
        username: "",
        namespace: import.meta.env.DEV ? "dev" : "",
        email: "",
        password: "",
        confirmPassword: "",
      },
    });

  const [namespaceEdited, setNamespaceEdited] = useState(false);
  const usernameValue = useWatch({ control, name: "username" });
  const namespaceValue = useWatch({ control, name: "namespace" });

  useEffect(() => {
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

    try {
      if (!token) throw new Error("no session issued");
      await loginWithToken(token);
      void navigate("/dashboard", {
        replace: true,
        state: firstRunEntryState(showOnboarding),
      });
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

  const surveyState = step === STEP_ONBOARDING ? "active" : "done";
  const accountStep = showOnboarding ? 2 : 1;

  return (
    <FirstRunLayout
      eyebrow="Welcome to ShellHub"
      signedIn={false}
      inConsole={false}
    >
      <Trail>
        {showOnboarding && (
          <TrailStep
            number={1}
            title={SETUP_STEP_TITLES.survey}
            state={surveyState}
            summary={surveyCompleted ? "Thanks" : "Skipped"}
          >
            <div className="space-y-4">
              <p className="text-xs text-text-secondary">
                Help us improve ShellHub by sharing your feedback.
              </p>
              <div className="relative h-[60dvh] overflow-auto rounded-lg border border-border">
                <iframe
                  src={onboardingUrl}
                  title="Onboarding survey"
                  className="absolute inset-0 w-full h-full border-0"
                />
              </div>
              <div className="flex items-center justify-end gap-3">
                {import.meta.env.DEV && (
                  <button
                    type="button"
                    onClick={() => setStep(STEP_ACCOUNT)}
                    className="text-2xs font-mono text-text-muted hover:text-text-secondary transition-colors"
                  >
                    Skip survey (dev only)
                  </button>
                )}
                <Button
                  disabled={!surveyCompleted}
                  onClick={() => setStep(STEP_ACCOUNT)}
                >
                  Continue
                </Button>
              </div>
            </div>
          </TrailStep>
        )}
        <TrailStep
          number={accountStep}
          title={SETUP_STEP_TITLES.account}
          state={step === STEP_ACCOUNT ? "active" : "upcoming"}
        >
          <form onSubmit={handleFormSubmit} className="space-y-4">
            <p className="text-xs text-text-secondary leading-relaxed">
              This instance has no users yet. You become its administrator, and
              the namespace is where your devices live.
            </p>

            {error && (
              <div className="flex items-start gap-2 bg-accent-red/8 border border-accent-red/20 text-accent-red px-3.5 py-2.5 rounded-md text-xs font-mono animate-slide-down">
                <ExclamationCircleIcon
                  className="w-3.5 h-3.5 shrink-0 mt-0.5"
                  strokeWidth={2}
                />
                {error}
              </div>
            )}

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
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
            </div>

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

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
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
            </div>

            <div className="flex items-center justify-end gap-3 pt-1">
              {showOnboarding && (
                <Button
                  variant="secondary"
                  onClick={() => setStep(STEP_ONBOARDING)}
                >
                  Back
                </Button>
              )}
              <Button
                type="submit"
                loading={loading}
                disabled={disableCreateAccountButton}
              >
                {loading ? "Setting up..." : "Create and continue"}
              </Button>
            </div>
          </form>
        </TrailStep>
        <UpcomingDeviceSteps start={accountStep + 1} />
      </Trail>
    </FirstRunLayout>
  );
}
