import { useEffect } from "react";
import { useNavigate } from "react-router-dom";
import {
  ArrowRightIcon,
  CheckCircleIcon,
} from "@heroicons/react/24/outline";
import { useAuthStore } from "../../stores/authStore";
import { useSignUpStore } from "../../stores/signUpStore";

export default function AccountCreated() {
  const navigate = useNavigate();
  const signUpToken = useSignUpStore((s) => s.signUpToken);
  const signUpTenant = useSignUpStore((s) => s.signUpTenant);
  const setSession = useAuthStore((s) => s.setSession);

  useEffect(() => {
    if (!signUpToken || !signUpTenant) return;

    setSession({ token: signUpToken, tenant: signUpTenant });

    // TODO: replace "/accept-invite" with the real route once it is added to App.tsx.
    const timer = setTimeout(() => { void navigate("/accept-invite"); }, 5000);
    return () => clearTimeout(timer);
  }, [signUpToken, signUpTenant, setSession, navigate]);

  const handleRedirect = () => {
    void navigate("/accept-invite"); // TODO: update when accept-invite route is added
  };

  return (
    <div className="w-full max-w-sm mx-auto animate-fade-in">
      <div className="bg-card/80 border border-border rounded-2xl p-8 backdrop-blur-sm text-center">
        <div className="inline-flex items-center justify-center w-14 h-14 rounded-full bg-accent-green/10 border border-accent-green/20 mb-5">
          <CheckCircleIcon className="w-7 h-7 text-accent-green" strokeWidth={1.5} />
        </div>

        <h2 className="text-lg font-semibold text-text-primary mb-3">
          Account Creation Successful
        </h2>

        <p className="text-sm text-text-secondary leading-relaxed mb-6">
          Thank you for registering an account on ShellHub. You will be
          redirected in 5 seconds. If you weren&apos;t redirected, please
          click the button below.
        </p>

        <button
          type="button"
          onClick={handleRedirect}
          className="inline-flex items-center gap-2 bg-primary hover:bg-primary/90 text-white px-5 py-2.5 rounded-lg text-sm font-semibold transition-all duration-200"
        >
          Redirect
          <ArrowRightIcon className="w-4 h-4" strokeWidth={2} />
        </button>
      </div>
    </div>
  );
}
