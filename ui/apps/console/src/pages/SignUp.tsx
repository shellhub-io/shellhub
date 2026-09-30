import { useState, useEffect, FormEvent } from "react";
import { useNavigate } from "react-router-dom";
import { useForm } from "react-hook-form";
import { signUpResolver } from "./setup/signUpResolver";
import type { SignUpFormValues } from "./setup/signUpResolver";
import { useSignUpStore } from "../stores/signUpStore";
import AccountCreated from "../components/auth/AccountCreated";
import { Button, Callout } from "@shellhub/design-system/primitives";
import PendingDeviceCallout from "@/components/auth/PendingDeviceCallout";
import AuthActions from "@/components/auth/AuthActions";
import {
  FormInputField,
  FormPasswordField,
  FormCheckboxField,
} from "@/components/common/fields/rhf";
import CheckboxField from "@/components/common/fields/CheckboxField";
import ScreenIntro from "@/components/layout/ScreenIntro";

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

  const [acceptMarketing, setAcceptMarketing] = useState(false);
  const [accountCreated, setAccountCreated] = useState(false);

  const { control, handleSubmit, setError, formState } =
    useForm<SignUpFormValues>({
      resolver: signUpResolver,
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
    for (const field of signUpServerFields) {
      const key = SERVER_FIELD_MAP[field];
      const message = SERVER_FIELD_MESSAGES[field];

      if (key && message) {
        setError(key, { type: "server", message });
      }
    }
  }, [signUpServerFields, setError]);

  const onSubmit = async (values: SignUpFormValues) => {
    if (signUpServerFields.length > 0) return;

    resetSignUpErrors();

    const token = await signUp({
      name: values.name,
      email: values.email,
      username: values.username,
      password: values.password,
      email_marketing: acceptMarketing,
    });

    const { signUpError: err, signUpServerFields: fields } =
      useSignUpStore.getState();

    if (err !== null || fields.length > 0) return;

    if (!token) {
      void navigate(
        `/confirm-account?username=${encodeURIComponent(values.username)}`,
      );
      return;
    }

    setAccountCreated(true);
  };

  const handleFormSubmit = (e: FormEvent) => {
    void handleSubmit(onSubmit)(e);
  };

  if (accountCreated) {
    return <AccountCreated />;
  }

  return (
    <>
      <ScreenIntro
        eyebrow="Account"
        title="Create your account"
        lead="A namespace is created with it, and your first device goes in there."
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
                loading={signUpLoading}
                disabled={
                  signUpLoading ||
                  !formState.isValid ||
                  signUpServerFields.length > 0
                }
              >
                {signUpLoading ? "Creating account..." : "Create Account"}
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
