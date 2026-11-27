import { useEffect, useState } from "react";
import { useMutation } from "@tanstack/react-query";
import { resendEmailMutation } from "@/client";
import { isSdkError } from "@/api/errors";

type ResendOutcome = "sent" | "throttled" | "failed";

/**
 * How long before the confirmation code can be sent again. Enforced by the server too; this is
 * what makes the button say so rather than fail.
 */
export const RESEND_COOLDOWN_S = 60;

/**
 * Resends the confirmation code, with the cooldown counted down for the button label. A send the
 * server throttles starts the cooldown too. Called with unprompted, for a send the user did not ask
 * for, handleResend keeps a throttle quiet: a code went out within the minute, so there is nothing
 * to tell. handleResend resolves to "sent", to "throttled" when one went out within the last
 * minute, to "failed" on any other error, or to undefined without an identifier.
 * clearResendNotice hides the last resend's success or error.
 */
export function useResendEmail(identifier: string) {
  const resend = useMutation(resendEmailMutation());
  const [resendError, setResendError] = useState<string | null>(null);
  const [resendSuccess, setResendSuccess] = useState(false);
  const [resendCooldown, setResendCooldown] = useState(0);

  useEffect(() => {
    if (resendCooldown <= 0) return;
    const id = setTimeout(() => setResendCooldown((c) => c - 1), 1000);
    return () => clearTimeout(id);
  }, [resendCooldown]);

  const handleResend = async ({
    unprompted = false,
  } = {}): Promise<ResendOutcome | undefined> => {
    if (!identifier) return undefined;
    setResendSuccess(false);
    setResendError(null);
    try {
      await resend.mutateAsync({ body: { identifier } });
    } catch (error: unknown) {
      if (!isSdkError(error) || error.status !== 429) {
        setResendError("Failed to resend email. Please try again.");
        return "failed";
      }
      setResendCooldown(RESEND_COOLDOWN_S);
      if (!unprompted) {
        setResendError("Wait a minute before asking for another email.");
      }
      return "throttled";
    }
    setResendCooldown(RESEND_COOLDOWN_S);
    setResendSuccess(true);
    return "sent";
  };

  const clearResendNotice = () => {
    setResendSuccess(false);
    setResendError(null);
  };

  return {
    handleResend,
    clearResendNotice,
    resendLoading: resend.isPending,
    resendError,
    resendSuccess,
    resendCooldown,
  };
}
