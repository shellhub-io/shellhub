import { useEffect, useState } from "react";
import { useSignUpStore } from "../stores/signUpStore";

/**
 * How long before a verification mail can be sent again. Enforced by the server too; this is
 * what makes the button say so rather than fail.
 */
export const RESEND_COOLDOWN_S = 60;

/**
 * Resends the verification mail, with the cooldown counted down for the button label.
 */
export function useResendEmail(username: string) {
  const resendEmail = useSignUpStore((s) => s.resendEmail);
  const resendLoading = useSignUpStore((s) => s.resendLoading);
  const resendError = useSignUpStore((s) => s.resendError);
  const [resendSuccess, setResendSuccess] = useState(false);
  const [resendCooldown, setResendCooldown] = useState(0);

  useEffect(() => {
    if (resendCooldown <= 0) return;
    const id = setTimeout(() => setResendCooldown((c) => c - 1), 1000);
    return () => clearTimeout(id);
  }, [resendCooldown]);

  const handleResend = async () => {
    if (!username) return;
    setResendSuccess(false);
    const ok = await resendEmail(username);
    if (ok) {
      setResendSuccess(true);
      setResendCooldown(RESEND_COOLDOWN_S);
    }
  };

  return { handleResend, resendLoading, resendError, resendSuccess, resendCooldown };
}
