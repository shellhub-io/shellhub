import { useEffect, useRef, useState, type FormEvent } from "react";
import {
  Navigate,
  useLocation,
  useNavigate,
  useSearchParams,
} from "react-router-dom";
import { Button, Callout } from "@shellhub/design-system/primitives";
import { isSdkError } from "@/api/errors";
import { useAuthStore } from "@/stores/authStore";
import { useSignUpStore } from "@/stores/signUpStore";
import { useResendEmail } from "@/hooks/useResendEmail";
import { useLockout } from "@/hooks/useLockout";
import { useOtpInput } from "@/hooks/useOtpInput";
import OtpCells from "@/components/mfa/OtpCells";
import AuthActions from "@/components/auth/AuthActions";
import LockoutCallouts from "@/components/auth/LockoutCallouts";
import ScreenIntro from "@/components/layout/ScreenIntro";
import {
  clearPendingDeviceCode,
  getPendingDeviceCode,
  hasPendingDeviceCode,
  landingAfterSignIn,
  setPendingDeviceCode,
} from "@/utils/navigation";
import { parsePairingCode } from "@/utils/pairingCode";

const wrongCodeMessage =
  "This code is wrong, expired or already used. If you already confirmed your email, sign in; otherwise check the latest email or resend it.";

/**
 * The page a new account confirms its email on: the six cells for the emailed code, exchanged for a
 * session, and the resend. A code in the URL, from the email's button, is submitted on arrival. A
 * device code in the URL, from a sign-up that started on an accept-device link, becomes the pending
 * one once the account is confirmed, so the session lands on accepting that device. A code the
 * server refuses keeps none. When it is the link's own code that is refused, the page also drops
 * that device code if sign-up left it pending, since the link may be one already used on another
 * device; a typed code that is refused leaves it, since it may be a typo. Arriving with sendCode in
 * the route state, from a sign-in that found the account pending, sends a code on arrival, since
 * nothing else on that path has. A wrong code clears the cells; too many lock the form out with a
 * countdown.
 */
export default function ConfirmAccount() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const identifier = searchParams.get("identifier") ?? "";
  const linkedCode = searchParams.get("code") ?? "";
  const linkedDevice = parsePairingCode(searchParams.get("device"));
  const location = useLocation();
  const sendCodeOnArrival =
    (location.state as { sendCode?: unknown } | null)?.sendCode === true;

  const confirmAccount = useAuthStore((s) => s.confirmAccount);
  const resetResendError = useSignUpStore((s) => s.resetResendError);
  const otp = useOtpInput(6);
  const [confirming, setConfirming] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [deviceWaiting, setDeviceWaiting] = useState(hasPendingDeviceCode);
  const {
    lockFrom,
    clearLockout,
    lockedOut,
    countdownDisplay,
    lockoutExpired,
  } = useLockout();
  const linkedCodeSubmitted = useRef(false);
  const codeSentOnArrival = useRef(false);

  const {
    handleResend,
    clearResendNotice,
    resendLoading,
    resendError,
    resendSuccess,
    resendCooldown,
  } = useResendEmail(identifier);

  useEffect(() => {
    resetResendError();
  }, [resetResendError]);

  const confirm = async (code: string, fromLink = false) => {
    setConfirming(true);
    setError(null);
    clearLockout();
    clearResendNotice();
    try {
      const signedIn = await confirmAccount(identifier, code);
      if (linkedDevice) setPendingDeviceCode(linkedDevice);
      if (signedIn) {
        void navigate(landingAfterSignIn("/dashboard"), { replace: true });
      } else {
        void navigate("/login", {
          replace: true,
          state: { notice: "Your email is confirmed. Sign in to continue." },
        });
      }
    } catch (err) {
      setConfirming(false);
      if (isSdkError(err) && err.status === 429) {
        lockFrom(err.headers);
        setError("Too many wrong codes. Please wait before trying again.");
        otp.reset();
      } else if (isSdkError(err) && err.status === 401) {
        if (
          fromLink &&
          linkedDevice &&
          parsePairingCode(getPendingDeviceCode()) === linkedDevice
        ) {
          clearPendingDeviceCode();
          setDeviceWaiting(false);
        }
        setError(wrongCodeMessage);
        otp.reset();
      } else {
        setError("Something went wrong. Please try again.");
      }
    }
  };

  const resend = async () => {
    const outcome = await handleResend();
    if (outcome === "sent") {
      setError((current) => (current === wrongCodeMessage ? null : current));
    }
  };

  useEffect(() => {
    if (!identifier || !sendCodeOnArrival || codeSentOnArrival.current) return;
    codeSentOnArrival.current = true;
    void navigate({ search: location.search }, { replace: true, state: null });
    void handleResend({ unprompted: true });
  });

  useEffect(() => {
    if (!identifier || !linkedCode || linkedCodeSubmitted.current) return;
    linkedCodeSubmitted.current = true;
    const kept = new URLSearchParams({ identifier });
    if (linkedDevice) kept.set("device", linkedDevice);
    void navigate({ search: `?${kept.toString()}` }, { replace: true });
    void confirm(linkedCode, true);
  });

  if (!identifier) return <Navigate to="/login" replace />;

  const handleSubmit = (e: FormEvent) => {
    e.preventDefault();
    if (!otp.isComplete) return;
    void confirm(otp.getValue());
  };

  const destination = identifier.includes("@")
    ? identifier
    : `the email address of ${identifier}`;

  return (
    <>
      <ScreenIntro
        eyebrow="Account"
        title="Confirm your email"
        lead={
          <>
            We sent a six-digit code to {destination}. Enter it below to confirm
            your email and sign in.
            {deviceWaiting &&
              " The link in the email works on any device, your phone included, and the device you are pairing keeps waiting."}
          </>
        }
      />

      <form onSubmit={handleSubmit} className="space-y-6">
        <div className="flex flex-col gap-3 empty:hidden">
          <LockoutCallouts
            error={error}
            countdownDisplay={countdownDisplay}
            lockoutExpired={lockoutExpired}
            expiredMessage="Your timeout has finished. Please enter the code again."
          />
          {resendSuccess && (
            <Callout variant="success">
              We sent the code again. It may take a minute to arrive.
            </Callout>
          )}
          {resendError && <Callout variant="error">{resendError}</Callout>}
        </div>

        <OtpCells otp={otp} label="Confirmation code" size="lg" numeric />

        <AuthActions
          primary={
            <Button
              variant="primary"
              size="lg"
              fullWidth
              type="submit"
              loading={confirming}
              disabled={confirming || lockedOut || !otp.isComplete}
            >
              {confirming ? "Confirming..." : "Confirm"}
            </Button>
          }
          links={[
            {
              label: resendLoading
                ? "Sending..."
                : resendCooldown > 0
                  ? `Resend Email (${resendCooldown}s)`
                  : "Resend Email",
              onClick: () => void resend(),
              disabled: resendLoading || resendCooldown > 0,
            },
            { label: "Back to sign in", to: "/login" },
          ]}
        />
      </form>
    </>
  );
}
