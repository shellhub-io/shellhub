import { useEffect } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { ArrowRightIcon, CheckCircleIcon } from "@heroicons/react/24/outline";
import { useAuthStore } from "@/stores/authStore";
import { useSignUpStore } from "@/stores/signUpStore";
import { Button } from "@shellhub/design-system/primitives";
import LoginLayoutCard from "@/components/layout/LoginLayoutCard";

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
    <LoginLayoutCard className="text-center">
      <div className="inline-flex items-center justify-center w-14 h-14 rounded-full bg-accent-green/10 border border-accent-green/20 mb-5">
        <CheckCircleIcon
          className="w-7 h-7 text-accent-green"
          strokeWidth={1.5}
        />
      </div>

      <h2 className="text-lg font-semibold text-text-primary mb-3">
        Account Creation Successful
      </h2>

      <p className="text-sm text-text-secondary leading-relaxed mb-6">
        Thank you for registering an account on ShellHub. You will be redirected
        in 5 seconds. If you weren&apos;t redirected, please click the button
        below.
      </p>

      <Button
        iconRight={<ArrowRightIcon className="w-4 h-4" strokeWidth={2} />}
        onClick={handleRedirect}
      >
        Redirect
      </Button>
    </LoginLayoutCard>
  );
}
