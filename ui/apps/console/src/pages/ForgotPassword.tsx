import { useState } from "react";
import { useForm } from "react-hook-form";
import { EnvelopeIcon } from "@heroicons/react/24/outline";
import { Button, Callout } from "@shellhub/design-system/primitives";
import { recoverPassword } from "../client";
import FormInputField from "@/components/common/fields/rhf/FormInputField";
import {
  forgotPasswordResolver,
  type ForgotPasswordFormValues,
} from "./setup/forgotPasswordResolver";
import AuthActions from "@/components/auth/AuthActions";
import ScreenIntro from "@/components/layout/ScreenIntro";
import { ignoreFailure } from "@/utils/failure";

const silenceToPreventAccountEnumeration = ignoreFailure;

/**
 * Requests a password reset. It reports success whatever the outcome, because saying whether the
 * address exists would let the form be used to enumerate accounts.
 */
export default function ForgotPassword() {
  const [loading, setLoading] = useState(false);
  const [sent, setSent] = useState(false);

  const { control, handleSubmit, formState } =
    useForm<ForgotPasswordFormValues>({
      resolver: forgotPasswordResolver,
      mode: "onTouched",
      defaultValues: { account: "" },
    });

  const onSubmit = async (values: ForgotPasswordFormValues) => {
    setLoading(true);

    await recoverPassword({
      body: { username: values.account },
      throwOnError: true,
    }).catch(silenceToPreventAccountEnumeration);

    setLoading(false);
    setSent(true);
  };

  const backToSignIn = [{ label: "Back to sign in", to: "/login" }];

  return (
    <>
      <ScreenIntro
        eyebrow="Password recovery"
        title="Forgot your password?"
        lead="Enter your username or email address and we'll send you a link to reset your password."
      />

      {sent ? (
        <>
          <Callout variant="success" className="mb-6">
            <span>
              <span className="font-semibold">Check your inbox.</span> An email
              with password reset instructions has been sent to your registered
              email address.
            </span>
          </Callout>
          <AuthActions links={backToSignIn} />
        </>
      ) : (
        <form
          onSubmit={(e) => void handleSubmit(onSubmit)(e)}
          className="space-y-6"
        >
          <FormInputField<ForgotPasswordFormValues>
            name="account"
            control={control}
            id="account"
            label="Username or email address"
            placeholder="username or email"
            autoComplete="username"
            required
          />

          <AuthActions
            primary={
              <Button
                variant="primary"
                size="lg"
                fullWidth
                type="submit"
                loading={loading}
                disabled={loading || !formState.isValid}
                icon={<EnvelopeIcon className="w-4 h-4" strokeWidth={2} />}
              >
                {loading ? "Sending..." : "Reset Password"}
              </Button>
            }
            links={backToSignIn}
          />
        </form>
      )}
    </>
  );
}
