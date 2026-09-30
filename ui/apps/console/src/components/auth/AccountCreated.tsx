import { useEffect } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { ArrowRightIcon } from "@heroicons/react/24/outline";
import { useAuthStore } from "@/stores/authStore";
import { useSignUpStore } from "@/stores/signUpStore";
import { Button, Callout } from "@shellhub/design-system/primitives";
import AuthActions from "@/components/auth/AuthActions";
import ScreenIntro from "@/components/layout/ScreenIntro";

/**
 * The screen after sign-up: what was sent, where, and how to have it sent again.
 */
export default function AccountCreated() {
  const navigate = useNavigate();
  const location = useLocation();
  const signUpToken = useSignUpStore((s) => s.signUpToken);
  const signUpTenant = useSignUpStore((s) => s.signUpTenant);
  const setSession = useAuthStore((s) => s.setSession);

  const acceptInviteTarget = `/accept-invite${location.search}`;

  useEffect(() => {
    if (!signUpToken || !signUpTenant) return;

    setSession({ token: signUpToken, tenant: signUpTenant });

    const timer = setTimeout(() => {
      void navigate(acceptInviteTarget);
    }, 5000);
    return () => clearTimeout(timer);
  }, [signUpToken, signUpTenant, setSession, navigate, acceptInviteTarget]);

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
