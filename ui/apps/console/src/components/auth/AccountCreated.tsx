import { useEffect } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { ArrowRightIcon } from "@heroicons/react/24/outline";
import { useAuthStore } from "@/stores/authStore";
import { signedInSession } from "@/stores/authSession";
import { Button, Callout } from "@shellhub/design-system/primitives";
import AuthActions from "@/components/auth/AuthActions";
import ScreenIntro from "@/components/layout/ScreenIntro";
import type { UserAuth } from "@/client";

/**
 * The screen after a sign-up that returned a session (auth). A session with a tenant signs the user
 * in and moves on to the invitation after five seconds, or at once through the button. One without
 * a tenant signs no one in and only moves on through the button.
 */
export default function AccountCreated({ auth: signUpAuth }: { auth: UserAuth }) {
  const navigate = useNavigate();
  const location = useLocation();

  const acceptInviteTarget = `/accept-invite${location.search}`;

  useEffect(() => {
    if (!signUpAuth.tenant) return;

    useAuthStore.setState(signedInSession(signUpAuth));

    const timer = setTimeout(() => {
      void navigate(acceptInviteTarget);
    }, 5000);
    return () => clearTimeout(timer);
  }, [signUpAuth, navigate, acceptInviteTarget]);

  const handleRedirect = () => {
    void navigate(acceptInviteTarget);
  };

  return (
    <>
      <ScreenIntro
        eyebrow="Account"
        title="Account Creation Successful"
        lead="Thank you for registering an account on ShellHub."
      />

      <Callout variant="success" className="mb-6">
        You will be redirected in 5 seconds. If you weren&apos;t redirected, use
        the button below.
      </Callout>

      <AuthActions
        primary={
          <Button
            size="lg"
            fullWidth
            iconRight={<ArrowRightIcon className="w-4 h-4" strokeWidth={2} />}
            onClick={handleRedirect}
          >
            Redirect
          </Button>
        }
      />
    </>
  );
}
