import { useEffect } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { Callout, Spinner } from "@shellhub/design-system/primitives";
import { useSignUpStore } from "@/stores/signUpStore";
import AuthActions from "@/components/auth/AuthActions";
import ScreenIntro from "@/components/layout/ScreenIntro";

/**
 * Completes an email verification from its link. An expired or reused link is reported
 * separately from a genuine failure, because only the first is worth offering a resend.
 */
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
    <>
      <ScreenIntro
        eyebrow="Account"
        title="Account Verification"
        lead="The link in your email confirms the address and activates the account."
      />

      <div className="mb-6" role="status" aria-live="polite">
        {validationStatus === "processing" || validationStatus === "idle" ? (
          <div className="flex items-center gap-3 text-sm text-text-secondary">
            <Spinner />
            Processing your account activation...
          </div>
        ) : validationStatus === "success" ? (
          <Callout variant="success">
            Congratulations! Your account has been activated successfully.
            Redirecting to login...
          </Callout>
        ) : validationStatus === "failed-token" ? (
          <Callout variant="error">
            Your account activation token has expired. Go to the login page and
            log in to receive another email with the activation link.
          </Callout>
        ) : (
          <Callout variant="error">
            There was a problem activating your account. Go to the login page
            and log in to receive another email with the activation link.
          </Callout>
        )}
      </div>

      <AuthActions links={[{ label: "Back to sign in", to: "/login" }]} />
    </>
  );
}
