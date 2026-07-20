import { useCallback, useState } from "react";
import { Link } from "react-router-dom";
import {
  XMarkIcon,
  ArrowRightIcon,
  ArrowLeftIcon,
  AdjustmentsHorizontalIcon,
} from "@heroicons/react/24/outline";
import { cn } from "@shellhub/design-system/cn";
import { Button, IconButton } from "@shellhub/design-system/primitives";
import { useResetOnOpen } from "@/hooks/useResetOnOpen";
import { useOtpInput } from "@/hooks/useOtpInput";
import { useAcceptDeviceByCode } from "@/hooks/useAcceptDeviceByCode";
import BaseDialog from "@/components/common/BaseDialog";
import WizardStepInstall from "./WizardStepInstall";
import WizardStepCode from "./WizardStepCode";
import WizardStepComplete from "./WizardStepComplete";
import WizardAcceptedWatcher from "./WizardAcceptedWatcher";
import { isWizardDemo, DEMO_DEVICE } from "./demo";

interface WelcomeWizardProps {
  open: boolean;
  /** Close for now: the wizard reappears next time (no accepted device yet). */
  onClose: () => void;
  /** Dismiss for good: skip explicitly or finish, so it never shows again. */
  onDismiss: () => void;
}

const PAIRING_CODE_LENGTH = 8;
const TOTAL_STEPS = 2;

export default function WelcomeWizard({
  open,
  onClose,
  onDismiss,
}: WelcomeWizardProps) {
  const [step, setStep] = useState(1);
  // Within step 1, the code-entry face (the "not on that machine" path). Not a
  // numbered step; the progress still reads step 1 while it shows.
  const [showCode, setShowCode] = useState(false);
  const [device, setDevice] = useState<{ uid: string; name: string } | null>(
    null,
  );

  // Code entry lives here (not in WizardStepCode) so the footer's Pair button
  // can read the value and drive the accept.
  const otp = useOtpInput(PAIRING_CODE_LENGTH, true);
  const {
    submit: acceptByCode,
    isPending: accepting,
    error: codeError,
    clearError,
  } = useAcceptDeviceByCode();

  useResetOnOpen(open, () => {
    setStep(1);
    setShowCode(false);
    setDevice(null);
    otp.reset();
    clearError();
  });

  const connect = (d: { uid: string; name: string }) => {
    setDevice(d);
    setStep(2);
  };

  const backToInstall = () => {
    setShowCode(false);
    otp.reset();
    clearError();
  };

  const acceptCode = async () => {
    const accepted = await acceptByCode(otp.getValue());
    if (accepted) connect(accepted);
  };

  // Closing the header X (or ESC/backdrop) defers on step 1 but, on the final
  // step, onboarding is already done (a device is connected), so it dismisses
  // for good, same as Finish. Memoized so BaseDialog's cancel-event listener
  // doesn't re-attach on every render (its dep-array includes it).
  const handleClose = useCallback(
    () => (step < TOTAL_STEPS ? onClose() : onDismiss()),
    [step, onClose, onDismiss],
  );

  return (
    <BaseDialog
      open={open}
      onClose={handleClose}
      focusOnOpen={false}
      size="xl"
      aria-label="Welcome to ShellHub"
      className="sm:max-h-[85vh]"
    >
      <header className="flex items-center justify-between px-6 pt-5 pb-0 shrink-0">
        {/* Progress dots */}
        <div className="flex items-center gap-2" aria-hidden="true">
          {Array.from({ length: TOTAL_STEPS }, (_, i) => i + 1).map((s) => (
            <div
              key={s}
              className={cn(
                "rounded-full transition-all duration-300",
                s < step
                  ? "w-2 h-2 bg-primary"
                  : s === step
                    ? "w-2.5 h-2.5 bg-primary shadow-[0_0_6px_rgba(102,122,204,0.5)]"
                    : "w-2 h-2 bg-border",
              )}
            />
          ))}
        </div>

        <IconButton onClick={handleClose} aria-label="Close wizard">
          <XMarkIcon className="w-4 h-4" />
        </IconButton>
      </header>

      <div
        role="progressbar"
        aria-valuenow={step}
        aria-valuemin={1}
        aria-valuemax={TOTAL_STEPS}
        aria-label={`Step ${step} of ${TOTAL_STEPS}, onboarding progress`}
        className="mx-6 mt-4 mb-5 h-px bg-border shrink-0 overflow-hidden rounded-full"
      >
        <div
          className="h-full bg-primary transition-all duration-500"
          style={{ width: `${(step / TOTAL_STEPS) * 100}%` }}
        />
      </div>

      <main className="flex-auto overflow-y-auto px-6 pb-6 min-h-0">
        {/* Watches for the link path across both faces of step 1, so accepting
            via the printed link advances the wizard even from the code face. */}
        {step === 1 && <WizardAcceptedWatcher onConnected={connect} />}
        {step === 1 && !showCode && <WizardStepInstall />}
        {step === 1 && showCode && (
          <WizardStepCode otp={otp} error={codeError} />
        )}
        {step === 2 && <WizardStepComplete device={device} />}
      </main>

      <footer className="px-6 py-4 border-t border-border shrink-0 flex items-center justify-between">
        {step === 1 && !showCode ? (
          // Escape hatch to the full page for Docker-specific methods, fleets,
          // or advanced agent options: anything past the simple codeless
          // install. Leaving defers the wizard (handleClose) rather than
          // dismissing it for good.
          <Link
            to="/devices/add"
            onClick={handleClose}
            className="flex items-center gap-1.5 text-xs text-text-muted hover:text-text-secondary transition-colors"
          >
            <AdjustmentsHorizontalIcon className="w-4 h-4" />
            Advanced Install
          </Link>
        ) : step === 1 && showCode ? (
          <Button
            variant="ghost"
            onClick={backToInstall}
            icon={<ArrowLeftIcon className="w-3.5 h-3.5" strokeWidth={2.5} />}
          >
            Back
          </Button>
        ) : (
          <span />
        )}

        <div className="flex items-center gap-3">
          {step < TOTAL_STEPS && (
            // Skip suppresses the wizard for good (the header X only defers it).
            <Button variant="ghost" onClick={onDismiss}>
              Skip
            </Button>
          )}

          {step === 1 && !showCode && (
            // The link path auto-advances on its own; Next is the manual route
            // to the code-entry face, for when the user isn't sitting at the
            // target machine to open the link.
            <Button
              onClick={() => setShowCode(true)}
              iconRight={
                <ArrowRightIcon className="w-3.5 h-3.5" strokeWidth={2.5} />
              }
            >
              Next
            </Button>
          )}

          {step === 1 && showCode && (
            // Accept the typed code. In dev demo mode it jumps straight to the
            // final step with a fake device, since a fake code can't resolve.
            <Button
              onClick={isWizardDemo() ? () => connect(DEMO_DEVICE) : acceptCode}
              loading={accepting}
              disabled={isWizardDemo() ? false : !otp.isComplete}
            >
              Pair device
            </Button>
          )}

          {step === TOTAL_STEPS && (
            // Finishing means the user completed onboarding, so dismiss for good.
            <Button
              onClick={onDismiss}
              iconRight={
                <ArrowRightIcon className="w-3.5 h-3.5" strokeWidth={2.5} />
              }
            >
              Finish
            </Button>
          )}
        </div>
      </footer>
    </BaseDialog>
  );
}
