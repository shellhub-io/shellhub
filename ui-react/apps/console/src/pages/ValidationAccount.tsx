import { useEffect } from "react";
import { Link, useNavigate, useSearchParams } from "react-router-dom";
import {
  CheckCircleIcon,
  XCircleIcon,
} from "@heroicons/react/24/outline";
import { useSignUpStore } from "../stores/signUpStore";

export default function ValidationAccount() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const validateAccount = useSignUpStore((s) => s.validateAccount);
  const validationStatus = useSignUpStore((s) => s.validationStatus);
  const resetValidation = useSignUpStore((s) => s.resetValidation);
  const setValidationFailed = useSignUpStore((s) => s.setValidationFailed);

  const email = searchParams.get("email") ?? "";
  const token = searchParams.get("token") ?? "";

  useEffect(() => {
    resetValidation();

    if (!email || !token) {
      setValidationFailed();
      return;
    }

    // Pass an AbortController so that the in-flight request is cancelled when
    // the effect cleanup fires (React Strict Mode double-invocation, unmount).
    // Without this, a single-use token would be consumed by the first call and
    // the second call would fail, leaving the user on the "failed" screen even
    // though activation succeeded.
    const controller = new AbortController();
    void validateAccount(email, token, controller.signal);

    return () => controller.abort();
  }, [email, token, validateAccount, resetValidation, setValidationFailed]);

  useEffect(() => {
    if (validationStatus !== "success") return;
    const timer = setTimeout(() => void navigate("/login"), 4000);
    return () => clearTimeout(timer);
  }, [validationStatus, navigate]);

  return (
    <div className="w-full max-w-5xl mx-auto px-8 py-12 flex flex-col items-center">
      <div
        className="w-full max-w-sm bg-card/80 border border-border rounded-2xl p-8 backdrop-blur-sm text-center animate-slide-up"
      >
        <h1 className="text-lg font-semibold text-text-primary mb-5">
          Account Verification
        </h1>

        <div
          className="min-h-32 flex flex-col items-center justify-center"
          role="status"
          aria-live="polite"
        >
          {validationStatus === "processing" || validationStatus === "idle" ? (
            <>
              <span className="w-10 h-10 border-2 border-primary/20 border-t-primary rounded-full animate-spin mb-4" />
              <p className="text-sm text-text-secondary">
                Processing your account activation...
              </p>
            </>
          ) : validationStatus === "success" ? (
            <>
              <div className="inline-flex items-center justify-center w-14 h-14 rounded-full bg-accent-green/10 border border-accent-green/20 mb-4">
                <CheckCircleIcon className="w-7 h-7 text-accent-green" strokeWidth={1.5} />
              </div>
              <p className="text-sm text-text-secondary leading-relaxed">
                Congratulations! Your account has been activated successfully.
                Redirecting to login...
              </p>
            </>
          ) : validationStatus === "failed-token" ? (
            <>
              <div className="inline-flex items-center justify-center w-14 h-14 rounded-full bg-accent-red/10 border border-accent-red/20 mb-4">
                <XCircleIcon className="w-7 h-7 text-accent-red" strokeWidth={1.5} />
              </div>
              <p className="text-sm text-text-secondary leading-relaxed">
                Your account activation token has expired. Go to the login page
                and log in to receive another email with the activation link.
              </p>
            </>
          ) : (
            <>
              <div className="inline-flex items-center justify-center w-14 h-14 rounded-full bg-accent-red/10 border border-accent-red/20 mb-4">
                <XCircleIcon className="w-7 h-7 text-accent-red" strokeWidth={1.5} />
              </div>
              <p className="text-sm text-text-secondary leading-relaxed">
                There was a problem activating your account. Go to the login
                page and log in to receive another email with the activation
                link.
              </p>
            </>
          )}
        </div>

        <p className="text-xs text-text-muted mt-6">
          Back to{" "}
          <Link
            to="/login"
            className="text-primary hover:text-primary/80 font-medium transition-colors"
          >
            Login
          </Link>
        </p>
      </div>
    </div>
  );
}
