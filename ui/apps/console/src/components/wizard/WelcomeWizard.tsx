import { useCallback, useState } from "react";
import {
  XMarkIcon,
  ArrowRightIcon,
  BookOpenIcon,
} from "@heroicons/react/24/outline";
import { Button, IconButton } from "@shellhub/design-system/primitives";
import { useResetOnOpen } from "@/hooks/useResetOnOpen";
import { useAcceptDevice } from "@/hooks/useDeviceMutations";
import type { NormalizedDevice } from "@/hooks/useDevices";
import { getAcceptDeviceErrorMessage } from "@/utils/deviceErrors";
import BaseDialog from "@/components/common/BaseDialog";
import WizardStep1Welcome from "./WizardStep1Welcome";
import WizardStep2Install from "./WizardStep2Install";
import WizardStep3Device from "./WizardStep3Device";
import WizardStep4Complete from "./WizardStep4Complete";

interface WelcomeWizardProps {
  open: boolean;
  onClose: () => void;
}

const TOTAL_STEPS = 4;

export default function WelcomeWizard({ open, onClose }: WelcomeWizardProps) {
  const [step, setStep] = useState(1);
  const [pendingDevice, setPendingDevice] = useState<NormalizedDevice | null>(
    null,
  );
  const [accepting, setAccepting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useResetOnOpen(open, () => {
    setStep(1);
    setPendingDevice(null);
    setAccepting(false);
    setError(null);
  });

  const acceptMutation = useAcceptDevice();

  // Stable predicate: closing is allowed on steps 1–3 but blocked on step 4.
  // Memoized to prevent BaseDialog's cancel-event listener from re-attaching
  // on every render (the listener dep-array includes canClose).
  const canClose = useCallback(() => step < TOTAL_STEPS, [step]);

  const handleAccept = async () => {
    if (!pendingDevice) return;
    setError(null);
    setAccepting(true);
    try {
      await acceptMutation.mutateAsync({ path: { uid: pendingDevice.uid } });
      setStep(4);
    } catch (err) {
      setError(getAcceptDeviceErrorMessage(err));
    } finally {
      setAccepting(false);
    }
  };

  return (
    <BaseDialog
      open={open}
      onClose={onClose}
      canClose={canClose}
      size="xl"
      aria-label="Welcome to ShellHub"
      className="sm:max-h-[85vh]"
    >
      {/* Header */}
      <header className="flex items-center justify-between px-6 pt-5 pb-0 shrink-0">
        {/* Progress dots */}
        <div className="flex items-center gap-2" aria-hidden="true">
          {Array.from({ length: TOTAL_STEPS }, (_, i) => i + 1).map((s) => (
            <div
              key={s}
              className={`rounded-full transition-all duration-300 ${
                s < step
                  ? "w-2 h-2 bg-primary"
                  : s === step
                    ? "w-2.5 h-2.5 bg-primary shadow-[0_0_6px_rgba(102,122,204,0.5)]"
                    : "w-2 h-2 bg-border"
              }`}
            />
          ))}
        </div>

        {step < TOTAL_STEPS ? (
          <IconButton onClick={onClose} aria-label="Close wizard">
            <XMarkIcon className="w-4 h-4" />
          </IconButton>
        ) : (
          // Spacer preserves the justify-between layout on the final step
          // without placing a focusable or ARIA-hidden interactive element in the tree.
          <div className="w-7 h-7" aria-hidden="true" />
        )}
      </header>

      {/* Step progress bar */}
      <div
        role="progressbar"
        aria-valuenow={step}
        aria-valuemin={1}
        aria-valuemax={TOTAL_STEPS}
        aria-label={`Step ${step} of ${TOTAL_STEPS} — onboarding progress`}
        className="mx-6 mt-4 mb-5 h-px bg-border shrink-0 overflow-hidden rounded-full"
      >
        <div
          className="h-full bg-primary transition-all duration-500"
          style={{ width: `${(step / TOTAL_STEPS) * 100}%` }}
        />
      </div>

      {/* Scrollable content */}
      <main className="flex-auto overflow-y-auto px-6 min-h-0">
        {step === 1 && <WizardStep1Welcome />}
        {step === 2 && (
          <WizardStep2Install onDeviceDetected={() => setStep(3)} />
        )}
        {step === 3 && (
          <WizardStep3Device
            device={pendingDevice}
            onDeviceLoaded={setPendingDevice}
          />
        )}
        {step === 4 && <WizardStep4Complete device={pendingDevice} />}
      </main>

      {/* Inline error — shown between scrollable content and footer on step 3 */}
      {step === 3 && error && (
        <p role="alert" className="shrink-0 px-6 py-2 text-xs text-accent-red">
          {error}
        </p>
      )}

      {/* Footer */}
      <footer className="px-6 py-4 border-t border-border shrink-0 flex items-center justify-between">
        {step === 2 ? (
          <a
            href="https://docs.shellhub.io/user-guides/devices/adding"
            target="_blank"
            rel="noopener noreferrer"
            className="flex items-center gap-1.5 text-xs text-text-muted hover:text-text-secondary transition-colors"
          >
            <BookOpenIcon className="w-4 h-4" />
            Docs
          </a>
        ) : (
          <span />
        )}

        <div className="flex items-center gap-3">
          {step < TOTAL_STEPS && (
            <Button variant="ghost" onClick={onClose}>
              Close
            </Button>
          )}

          {step === 1 && (
            <Button
              onClick={() => setStep(2)}
              iconRight={
                <ArrowRightIcon className="w-3.5 h-3.5" strokeWidth={2.5} />
              }
            >
              Next
            </Button>
          )}

          {step === 2 && (
            // Disabled — polling in WizardStep2Install auto-advances to step 3
            <Button
              disabled
              iconRight={
                <ArrowRightIcon className="w-3.5 h-3.5" strokeWidth={2.5} />
              }
            >
              Next
            </Button>
          )}

          {step === 3 && (
            <Button
              onClick={() => void handleAccept()}
              disabled={!pendingDevice}
              loading={accepting}
            >
              {accepting ? "Accepting…" : "Accept"}
            </Button>
          )}

          {step === TOTAL_STEPS && (
            // Finish always closes directly. canClose blocks ESC/backdrop on
            // earlier steps; the Finish button is intentionally unrestricted.
            <Button
              onClick={onClose}
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
