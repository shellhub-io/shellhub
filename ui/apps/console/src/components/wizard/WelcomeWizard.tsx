import { useCallback, useState } from "react";
import {
  XMarkIcon,
  ArrowRightIcon,
  BookOpenIcon,
} from "@heroicons/react/24/outline";
import { cn } from "@shellhub/design-system/cn";
import { Button, IconButton } from "@shellhub/design-system/primitives";
import { useResetOnOpen } from "@/hooks/useResetOnOpen";
import BaseDialog from "@/components/common/BaseDialog";
import WizardStep1Welcome from "./WizardStep1Welcome";
import WizardStep2Install from "./WizardStep2Install";
import WizardStep4Complete from "./WizardStep4Complete";

interface WelcomeWizardProps {
  open: boolean;
  onClose: () => void;
}

const TOTAL_STEPS = 3;

export default function WelcomeWizard({ open, onClose }: WelcomeWizardProps) {
  const [step, setStep] = useState(1);
  const [device, setDevice] = useState<{ uid: string; name: string } | null>(
    null,
  );

  useResetOnOpen(open, () => {
    setStep(1);
    setDevice(null);
  });

  // Stable predicate: closing is allowed on steps 1–2 but blocked on the final
  // step. Memoized to prevent BaseDialog's cancel-event listener from
  // re-attaching on every render (its dep-array includes canClose).
  const canClose = useCallback(() => step < TOTAL_STEPS, [step]);

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
          <WizardStep2Install
            onConnected={(d) => {
              setDevice(d);
              setStep(3);
            }}
          />
        )}
        {step === 3 && <WizardStep4Complete device={device} />}
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
            // Disabled — the device auto-accepts and WizardStep2Install advances
            // to the final step the moment it connects.
            <Button
              disabled
              iconRight={
                <ArrowRightIcon className="w-3.5 h-3.5" strokeWidth={2.5} />
              }
            >
              Next
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
