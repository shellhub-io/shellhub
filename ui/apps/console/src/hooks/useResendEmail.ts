import { useEffect, useState } from "react";
import { useSignUpStore } from "../stores/signUpStore";

/**
 * How long before the confirmation code can be sent again. Enforced by the server too; this is
 * what makes the button say so rather than fail.
 */
export const RESEND_COOLDOWN_S = 60;

/**
 * Resends the confirmation code, with the cooldown counted down for the button label. A send the
 * server throttles starts the cooldown too. Called with unprompted, for a send the user did not ask
 * for, handleResend keeps a throttle quiet: a code went out within the minute, so there is nothing
 * to tell. handleResend resolves to the outcome, or undefined without an identifier, and
 * clearResendNotice hides the last resend's success or error.
 */
export function useResendEmail(identifier: string) {
  const resendEmail = useSignUpStore((s) => s.resendEmail);
  const resetResendError = useSignUpStore((s) => s.resetResendError);
  const resendLoading = useSignUpStore((s) => s.resendLoading);
  const resendError = useSignUpStore((s) => s.resendError);
  const [resendSuccess, setResendSuccess] = useState(false);
  const [resendCooldown, setResendCooldown] = useState(0);

  useEffect(() => {
    if (resendCooldown <= 0) return;
    const id = setTimeout(() => setResendCooldown((c) => c - 1), 1000);
    return () => clearTimeout(id);
  }, [resendCooldown]);

  const handleResend = async ({ unprompted = false } = {}) => {
    if (!identifier) return undefined;
    setResendSuccess(false);
    const outcome = await resendEmail(identifier);
    if (outcome === "failed") return outcome;
    setResendCooldown(RESEND_COOLDOWN_S);
    if (outcome === "sent") setResendSuccess(true);
    else if (unprompted) resetResendError();
    return outcome;
  };

  const clearResendNotice = () => {
    setResendSuccess(false);
    resetResendError();
  };

  return {
    handleResend,
    clearResendNotice,
    resendLoading,
    resendError,
    resendSuccess,
    resendCooldown,
  };
}
