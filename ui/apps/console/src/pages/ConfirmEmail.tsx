import { useMutation } from "@tanstack/react-query";
import { Link, useSearchParams } from "react-router-dom";
import { Button, Callout } from "@shellhub/design-system/primitives";
import { confirmEmailChangeMutation } from "@/client";
import { isSdkError } from "@/api/errors";
import AuthActions from "@/components/auth/AuthActions";
import ScreenIntro from "@/components/layout/ScreenIntro";

const copy = {
  email: {
    eyebrow: "Email change",
    title: "Confirm your new email",
    lead: "Make this address the email of your ShellHub account.",
    button: "Confirm Email",
    doneTitle: "Email confirmed",
    doneLead:
      "Your account now uses this address. Sign in with it from now on.",
    again: "Change your email again from your account.",
  },
  recovery_email: {
    eyebrow: "Recovery email change",
    title: "Confirm your new recovery email",
    lead: "Make this address the recovery email of your ShellHub account.",
    button: "Confirm Recovery Email",
    doneTitle: "Recovery email confirmed",
    doneLead: "Your account now uses this address for account recovery.",
    again: "Change your recovery email again from your account.",
  },
};

/**
 * Applies an email or recovery email change from the link mailed to the new address. The user ID
 * and the token are in the query string, and field=recovery_email marks a recovery email link,
 * which changes only the wording: the server finds the change from the token. The token works
 * once, and only when the button is pressed.
 */
export default function ConfirmEmail() {
  const [searchParams] = useSearchParams();

  const uid = searchParams.get("id") ?? "";
  const token = searchParams.get("token") ?? "";
  const text =
    searchParams.get("field") === "recovery_email"
      ? copy.recovery_email
      : copy.email;
  const failureMessages: Record<number, string> = {
    400: `This address can no longer be used. ${text.again}`,
    409: `Another account took this address. ${text.again}`,
  };

  const confirm = useMutation({ ...confirmEmailChangeMutation() });

  const backToSignIn = [{ label: "Back to sign in", to: "/login" }];

  const failure = confirm.isError
    ? (failureMessages[isSdkError(confirm.error) ? confirm.error.status : 0] ??
      `This link is invalid, expired or already used. ${text.again}`)
    : "";

  if (!uid || !token) {
    return (
      <>
        <ScreenIntro
          eyebrow={text.eyebrow}
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
          eyebrow={text.eyebrow}
          title={text.doneTitle}
          lead={text.doneLead}
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
      <ScreenIntro eyebrow={text.eyebrow} title={text.title} lead={text.lead} />
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
              {confirm.isPending ? "Confirming..." : text.button}
            </Button>
          }
          links={backToSignIn}
        />
      </div>
    </>
  );
}
