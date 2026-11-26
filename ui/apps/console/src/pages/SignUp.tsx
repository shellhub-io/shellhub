import { useState, useEffect, FormEvent } from "react";
import { useNavigate } from "react-router-dom";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";
import { Button, Callout } from "@shellhub/design-system/primitives";
import { registerUserMutation, type UserAuth } from "@/client";
import { apiErrorFields, apiErrorMessage, isSdkError } from "@/api/errors";
import PendingDeviceCallout from "@/components/auth/PendingDeviceCallout";
import { getPendingDeviceCode, hasPendingDeviceCode } from "@/utils/navigation";
import { parsePairingCode } from "@/utils/pairingCode";
import AuthActions from "@/components/auth/AuthActions";
import {
  FormInputField,
  FormPasswordField,
  FormCheckboxField,
} from "@/components/common/fields/rhf";
import CheckboxField from "@/components/common/fields/CheckboxField";
import ScreenIntro from "@/components/layout/ScreenIntro";
import { emailRefusalMessage } from "@/utils/emailRefusal";
import AccountCreated from "@/components/auth/AccountCreated";
import { signUpSchema, type SignUpFormValues } from "./setup/schemas";

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

/**
 * Self-registration, on cloud only. Field errors from the server are shown against the fields
 * they name rather than as one message, since most of them are about a taken username.
 */
export default function SignUp() {
  const navigate = useNavigate();
  const register = useMutation(registerUserMutation());
  const [signUpServerFields, setSignUpServerFields] = useState<
    Record<string, string>
  >({});
  const hasServerFieldErrors = Object.keys(signUpServerFields).length > 0;
  const signUpError =
    register.error && Object.keys(apiErrorFields(register.error)).length === 0
      ? apiErrorMessage(register.error)
      : null;

  const clearSignUpServerField = (field: string) =>
    setSignUpServerFields((fields) =>
      Object.fromEntries(Object.entries(fields).filter(([f]) => f !== field)),
    );

  const [acceptMarketing, setAcceptMarketing] = useState(false);

  const { control, handleSubmit, setError, formState } =
    useForm<SignUpFormValues>({
      resolver: zodResolver(signUpSchema),
      mode: "onTouched",
      defaultValues: {
        name: "",
        username: "",
        email: "",
        password: "",
        confirmPassword: "",
        acceptPrivacyPolicy: false,
      },
    });

  useEffect(() => {
    for (const [field, reason] of Object.entries(signUpServerFields)) {
      const key = SERVER_FIELD_MAP[field];
      const message =
        (field === "email" ? emailRefusalMessage(reason) : undefined) ??
        SERVER_FIELD_MESSAGES[field];

      if (key && message) {
        setError(key, { type: "server", message });
      }
    }
  }, [signUpServerFields, setError]);

  const onSubmit = async (values: SignUpFormValues) => {
    if (hasServerFieldErrors) return;

    const deviceCode = parsePairingCode(getPendingDeviceCode());
    let auth: UserAuth | void;
    try {
      auth = await register.mutateAsync({
        body: {
          name: values.name,
          email: values.email,
          username: values.username,
          password: values.password,
          email_marketing: acceptMarketing,
          ...(deviceCode && { device_code: deviceCode }),
        },
      });
    } catch (error: unknown) {
      const fields = apiErrorFields(error);
      if (Object.keys(fields).length > 0) {
        setSignUpServerFields(fields);
        return;
      }
      if (import.meta.env.DEV && isSdkError(error)) {
        console.warn("Unhandled sign-up error response:", {
          status: error.status,
        });
      }
      return;
    }

    if (!auth?.token) {
      const confirm = new URLSearchParams({ identifier: values.email });
      if (deviceCode) confirm.set("device", deviceCode);
      void navigate(`/confirm-account?${confirm.toString()}`);
    }
  };

  const handleFormSubmit = (e: FormEvent) => {
    void handleSubmit(onSubmit)(e);
  };

  if (register.data?.token) {
    return <AccountCreated auth={register.data} />;
  }

  return (
    <>
      <ScreenIntro
        eyebrow="Account"
        title="Create your account"
        lead={
          hasPendingDeviceCode()
            ? "Next you name a namespace and accept the device that is waiting."
            : "Next you name a namespace and connect your first device to it."
        }
      />

      <div className="flex flex-col gap-3 mb-6 empty:hidden">
        <PendingDeviceCallout />
        {signUpError && <Callout variant="error">{signUpError}</Callout>}
      </div>

      <form
        onSubmit={handleFormSubmit}
        className="space-y-4"
        aria-label="Create account"
      >
        <div className="space-y-4">
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
            autoCapitalize="none"
            onValueChange={() => clearSignUpServerField("username")}
          />
        </div>

        <FormInputField<SignUpFormValues>
          id="email"
          label="Email"
          name="email"
          control={control}
          type="email"
          placeholder="you@example.com"
          autoComplete="email"
          onValueChange={() => clearSignUpServerField("email")}
        />

        <div className="space-y-4">
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
        </div>

        <div className="space-y-2 pt-1">
          <FormCheckboxField<SignUpFormValues>
            id="signup-accept-privacy"
            name="acceptPrivacyPolicy"
            control={control}
            required
            labelSize="sm"
            label={
              <>
                I agree to the{" "}
                <a
                  href="https://www.shellhub.io/privacy-policy"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-text-primary underline decoration-border-light underline-offset-2 hover:decoration-text-secondary transition-colors"
                >
                  Privacy Policy
                </a>
                .
              </>
            }
          />

          <CheckboxField
            id="signup-accept-marketing"
            checked={acceptMarketing}
            onChange={setAcceptMarketing}
            labelSize="sm"
            label="Send me news and updates from ShellHub by email."
          />
        </div>

        <div className="pt-1">
          <AuthActions
            primary={
              <Button
                variant="primary"
                size="lg"
                fullWidth
                type="submit"
                loading={register.isPending}
                disabled={
                  register.isPending || !formState.isValid || hasServerFieldErrors
                }
              >
                {register.isPending ? "Creating account..." : "Create Account"}
              </Button>
            }
            links={[
              { label: "Already have an account? Sign in", to: "/login" },
            ]}
          />
        </div>
      </form>
    </>
  );
}
