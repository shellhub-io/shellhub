import { useMutation } from "@tanstack/react-query";
import { useNavigate } from "react-router-dom";
import { Button, Callout } from "@shellhub/design-system/primitives";
import { linkSamlAccountMutation } from "@/client";
import { isSdkError } from "@/api/errors";
import AuthActions from "@/components/auth/AuthActions";
import ScreenIntro from "@/components/layout/ScreenIntro";
import { landingAfterSignIn } from "@/utils/navigation";

/**
 * Confirms the link an SSO sign-in left waiting when its email matched this account. The user
 * reaches it right after signing in to the account by its own credentials, and the link happens
 * only when the button is pressed.
 */
export default function SsoLink() {
  const navigate = useNavigate();
  const link = useMutation({ ...linkSamlAccountMutation() });

  const failure = link.isError
    ? isSdkError(link.error) && link.error.status === 403
      ? "There is no single sign-on link waiting for this account. It expires after ten minutes and belongs to the account whose email your identity provider sent. Sign in with SSO again to start over."
      : "Failed to link this account to single sign-on. Please try again."
    : "";

  const proceed = () => void navigate(landingAfterSignIn("/dashboard"));

  if (link.isSuccess) {
    return (
      <>
        <ScreenIntro
          eyebrow="Single sign-on"
          title="Account linked"
          lead="This account now signs in through your identity provider."
        />
        <AuthActions
          primary={
            <Button size="lg" fullWidth onClick={proceed}>
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
        eyebrow="Single sign-on"
        title="Link this account"
        lead="Sign in to this account through your identity provider from now on."
      />
      <div className="space-y-6">
        {failure && <Callout variant="error">{failure}</Callout>}
        <AuthActions
          primary={
            <Button
              variant="primary"
              size="lg"
              fullWidth
              loading={link.isPending}
              disabled={link.isPending}
              onClick={() => link.mutate({})}
            >
              {link.isPending ? "Linking..." : "Link Account"}
            </Button>
          }
          links={[{ label: "Not now", onClick: proceed }]}
        />
      </div>
    </>
  );
}
