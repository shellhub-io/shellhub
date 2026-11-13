import { useMutation } from "@tanstack/react-query";
import { Link, useSearchParams } from "react-router-dom";
import { Button, Callout } from "@shellhub/design-system/primitives";
import { confirmEmailChangeMutation } from "@/client";
import { isSdkError } from "@/api/errors";
import AuthActions from "@/components/auth/AuthActions";
import ScreenIntro from "@/components/layout/ScreenIntro";

const failureMessages: Record<number, string> = {
  400: "This address can no longer be used. Change your email again from your account.",
  409: "Another account took this address. Change your email again from your account.",
};

/**
 * Applies an email change from the link mailed to the new address. The user ID and the token are
 * in the query string; the token works once, and only when the button is pressed.
 */
export default function ConfirmEmail() {
  const [searchParams] = useSearchParams();

  const uid = searchParams.get("id") ?? "";
  const token = searchParams.get("token") ?? "";

  const confirm = useMutation({ ...confirmEmailChangeMutation() });

  const backToSignIn = [{ label: "Back to sign in", to: "/login" }];

  const failure = confirm.isError
    ? (failureMessages[isSdkError(confirm.error) ? confirm.error.status : 0] ??
      "This link is invalid, expired or already used. Change your email again from your account.")
    : "";

  if (!uid || !token) {
    return (
      <>
        <ScreenIntro
          eyebrow="Email change"
          title="Invalid confirmation link"
          lead="This confirmation link is incomplete. Open the link from the email again."
        />
        <AuthActions links={backToSignIn} />
      </>
    );
  }

  if (confirm.isSuccess) {
    return (
      <>
        <ScreenIntro
          eyebrow="Email change"
          title="Email confirmed"
          lead="Your account now uses this address. Sign in with it from now on."
        />
        <AuthActions
          primary={
            <Button as={Link} to="/" size="lg" fullWidth>
              Continue
            </Button>
          }
        />
      </>
    );
  }

  return (
    <>
      <ScreenIntro
        eyebrow="Email change"
        title="Confirm your new email"
        lead="Make this address the email of your ShellHub account."
      />
      <div className="space-y-6">
        {failure && <Callout variant="error">{failure}</Callout>}
        <AuthActions
          primary={
            <Button
              variant="primary"
              size="lg"
              fullWidth
              loading={confirm.isPending}
              disabled={confirm.isPending}
              onClick={() => confirm.mutate({ path: { uid }, body: { token } })}
            >
              {confirm.isPending ? "Confirming..." : "Confirm Email"}
            </Button>
          }
          links={backToSignIn}
        />
      </div>
    </>
  );
}
