import { useEffect, useState } from "react";
import { useSignUpStore } from "../stores/signUpStore";

export const RESEND_COOLDOWN_S = 60;

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
