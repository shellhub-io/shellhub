import { useState, useEffect, FormEvent } from "react";
import { useForm, useWatch } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { useNavigate } from "react-router-dom";
import { PencilSquareIcon } from "@heroicons/react/24/outline";
import { Button, Callout } from "@shellhub/design-system/primitives";
import { isSdkError } from "@/api/errors";
import { setup } from "@/client";
import { useAuthStore } from "@/stores/authStore";
import {
  FormInputField,
  FormPasswordField,
} from "@/components/common/fields/rhf";
import FirstRunLayout from "@/components/firstRun/FirstRunLayout";
import {
  Trail,
  TrailStep,
  UpcomingDeviceSteps,
} from "@/components/firstRun/Trail";
import { SETUP_STEP_TITLES } from "@/components/firstRun/stepTitles";
import { firstRunEntryState } from "@/components/firstRun/entry";
import { suggestNamespaceName } from "@/utils/validation";
import { setupSchema, type SetupFormValues } from "./setup/schemas";

/**
 * First-run setup for a fresh instance, as the opening step of the first-run trail: the first
 * account, which becomes the instance's admin, with its namespace. It signs that account straight
 * in, since there is nobody yet to sign in as, and hands over to the dashboard, where the trail
 * goes on to the first device.
 */
export default function Setup() {
  const navigate = useNavigate();
  const loginWithToken = useAuthStore((state) => state.loginWithToken);

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const { control, handleSubmit, formState, setValue } =
    useForm<SetupFormValues>({
      resolver: zodResolver(setupSchema),
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
      setValue("namespace", suggestNamespaceName(usernameValue ?? ""), {
        shouldValidate: true,
      });
    }
  }, [usernameValue, namespaceEdited, setValue]);

  const disableCreateAccountButton = loading || !formState.isValid;

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
        state: firstRunEntryState(),
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

  return (
    <FirstRunLayout
      eyebrow="Welcome to ShellHub"
      signedIn={false}
      inConsole={false}
    >
      <Trail>
        <TrailStep number={1} title={SETUP_STEP_TITLES.account} state="active">
          <form onSubmit={handleFormSubmit} className="space-y-4">
            <p className="text-xs text-text-secondary leading-relaxed">
              This instance has no users yet. You become its administrator, and
              the namespace is where your devices live.
            </p>

            {error && <Callout variant="error">{error}</Callout>}

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
        <UpcomingDeviceSteps start={2} />
      </Trail>
    </FirstRunLayout>
  );
}
