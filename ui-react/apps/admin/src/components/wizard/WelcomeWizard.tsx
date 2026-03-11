import { useCallback, useState } from "react";
import { XMarkIcon, ArrowRightIcon, BookOpenIcon } from "@heroicons/react/24/outline";
import { useResetOnOpen } from "@/hooks/useResetOnOpen";
import { useDevicesStore } from "@/stores/devicesStore";
import { Device } from "@/types/device";
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
  const [pendingDevice, setPendingDevice] = useState<Device | null>(null);
  const [accepting, setAccepting] = useState(false);

  useResetOnOpen(open, () => {
    setStep(1);
    setPendingDevice(null);
    setAccepting(false);
  });

  const { accept } = useDevicesStore();

  // Stable predicate: closing is allowed on steps 1–3 but blocked on step 4.
  // Memoized to prevent BaseDialog's cancel-event listener from re-attaching
  // on every render (the listener dep-array includes canClose).
  const canClose = useCallback(() => step < TOTAL_STEPS, [step]);

  const handleAccept = async () => {
    if (!pendingDevice) return;
    setAccepting(true);
    try {
      await accept(pendingDevice.uid);
      setStep(4);
    } catch {
      // Keep on step 3 — user can retry
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
          <button
            onClick={onClose}
            className="p-1.5 rounded-lg text-text-muted hover:text-text-primary hover:bg-hover-medium transition-all"
            aria-label="Close wizard"
          >
            <XMarkIcon className="w-4 h-4" />
          </button>
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
      <main className="flex-1 overflow-y-auto px-6 min-h-0">
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
            <button
              onClick={onClose}
              className="px-4 py-2 rounded-lg text-xs font-medium text-text-muted hover:text-text-secondary hover:bg-hover-medium transition-all"
            >
              Close
            </button>
          )}

          {step === 1 && (
            <PrimaryButton onClick={() => setStep(2)}>
              Next <ArrowRightIcon className="w-3.5 h-3.5" strokeWidth={2.5} />
            </PrimaryButton>
          )}

          {step === 2 && (
            // Disabled — polling in WizardStep2Install auto-advances to step 3
            <PrimaryButton disabled>
              Next <ArrowRightIcon className="w-3.5 h-3.5" strokeWidth={2.5} />
            </PrimaryButton>
          )}

          {step === 3 && (
            <PrimaryButton
              onClick={handleAccept}
              disabled={!pendingDevice || accepting}
              loading={accepting}
            >
              {accepting ? "Accepting…" : "Accept"}
            </PrimaryButton>
          )}

          {step === TOTAL_STEPS && (
            // Finish always closes directly. canClose blocks ESC/backdrop on
            // earlier steps; the Finish button is intentionally unrestricted.
            <PrimaryButton onClick={onClose}>
              Finish <ArrowRightIcon className="w-3.5 h-3.5" strokeWidth={2.5} />
            </PrimaryButton>
          )}
        </div>
      </footer>
    </BaseDialog>
  );
}

function PrimaryButton({
  children,
  onClick,
  disabled = false,
  loading = false,
}: {
  children: React.ReactNode;
  onClick?: () => void;
  disabled?: boolean;
  loading?: boolean;
}) {
  return (
    <button
      onClick={onClick}
      disabled={disabled || loading}
      className="inline-flex items-center gap-1.5 px-4 py-2 rounded-lg text-xs font-semibold transition-all duration-200
        bg-primary text-white hover:bg-primary-600 active:scale-[0.98]
        disabled:opacity-40 disabled:cursor-not-allowed disabled:hover:bg-primary disabled:active:scale-100"
    >
      {children}
    </button>
  );
}
