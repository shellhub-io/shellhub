import { useEffect, useRef, useState } from "react";
import { XMarkIcon, ArrowRightIcon, BookOpenIcon } from "@heroicons/react/24/outline";
import { useFocusTrap } from "@/hooks/useFocusTrap";
import { useBackdropClose } from "@/hooks/useBackdropClose";
import { useDevicesStore } from "@/stores/devicesStore";
import { Device } from "@/types/device";
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

  const dialogRef = useRef<HTMLDialogElement>(null);
  const backdropHandlers = useBackdropClose(dialogRef, onClose, () => step < TOTAL_STEPS);
  useFocusTrap(dialogRef, open);

  const { accept } = useDevicesStore();

  // Open/close the native modal
  useEffect(() => {
    const dialog = dialogRef.current;
    if (!dialog) return;
    if (open) {
      dialog.showModal();
    } else {
      dialog.close();
    }
  }, [open]);

  // Handle Escape via the native cancel event (fired by showModal dialogs)
  useEffect(() => {
    const dialog = dialogRef.current;
    if (!dialog) return;
    const handleCancel = (e: Event) => {
      e.preventDefault(); // prevent the dialog from closing itself
      if (step < TOTAL_STEPS) onClose();
      // on step 4, Escape is blocked intentionally
    };
    dialog.addEventListener("cancel", handleCancel);
    return () => dialog.removeEventListener("cancel", handleCancel);
  }, [step, onClose]);

  // Reset to step 1 whenever the dialog opens so it always starts fresh
  useEffect(() => {
    if (open) {
      setStep(1);
      setPendingDevice(null);
      setAccepting(false);
    }
  }, [open]);

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

  if (!open) return null;

  return (
      <dialog
        ref={dialogRef}
        data-custom-backdrop
        aria-label="Welcome to ShellHub"
        {...backdropHandlers}
        className="fixed inset-0 z-[60] m-auto w-full h-full sm:h-auto sm:max-w-xl sm:max-h-[85vh] bg-surface sm:border sm:border-border sm:rounded-2xl shadow-2xl shadow-black/40 flex flex-col animate-slide-up"
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

          <button
            onClick={step < TOTAL_STEPS ? onClose : undefined}
            className={`p-1.5 rounded-lg text-text-muted hover:text-text-primary hover:bg-hover-medium transition-all ${step >= TOTAL_STEPS ? "invisible" : ""}`}
            aria-label="Close wizard"
            aria-hidden={step >= TOTAL_STEPS}
            tabIndex={step >= TOTAL_STEPS ? -1 : undefined}
          >
            <XMarkIcon className="w-4 h-4" />
          </button>
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
              <PrimaryButton onClick={onClose}>
                Finish <ArrowRightIcon className="w-3.5 h-3.5" strokeWidth={2.5} />
              </PrimaryButton>
            )}
          </div>
        </footer>
      </dialog>
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
        bg-primary text-white hover:bg-primary/90 active:scale-[0.98]
        disabled:opacity-40 disabled:cursor-not-allowed disabled:hover:bg-primary disabled:active:scale-100"
    >
      {children}
    </button>
  );
}
