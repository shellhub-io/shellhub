import { ReactNode, startTransition, useCallback, useState } from "react";
import { isSdkError } from "@/api/errors";
import {
  ArrowRightIcon,
  ChevronLeftIcon,
  XMarkIcon,
} from "@heroicons/react/24/outline";
import BaseDialog from "../common/BaseDialog";
import { useCreateSubscription, useSubscription } from "@/hooks/useBilling";
import BillingLetter from "./BillingLetter";
import BillingPayment from "./BillingPayment";
import BillingCheckout from "./BillingCheckout";
import BillingSuccessful from "./BillingSuccessful";

const STEPS = ["Overview", "Payment method", "Review", "Success"] as const;
const TOTAL_STEPS = STEPS.length;

interface BillingDialogProps {
  open: boolean;
  onClose: () => void;
  onSuccess?: () => void;
}

export default function BillingDialog({
  open,
  onClose,
  onSuccess,
}: BillingDialogProps) {
  const [step, setStep] = useState(1);
  const [hasDefault, setHasDefault] = useState(false);
  const [error, setError] = useState("");
  const createSubscription = useCreateSubscription();
  const { refetch: refetchSubscription } = useSubscription(false);

  const goNext = useCallback(
    () => startTransition(() => setStep((s) => s + 1)),
    [],
  );
  const goBack = useCallback(() => {
    setError("");
    startTransition(() => setStep((s) => Math.max(1, s - 1)));
  }, []);

  const subscribe = async () => {
    setError("");
    try {
      await createSubscription.mutateAsync({});
      const { data: sub } = await refetchSubscription();
      const subStatus = sub?.status;
      if (!subStatus || !["active", "trialing"].includes(subStatus)) {
        setError(
          subStatus === "incomplete" || subStatus === "incomplete_expired"
            ? "Your payment requires additional confirmation. Please check your email or contact support."
            : "Your subscription wasn't fully activated. Please try again.",
        );
        return;
      }
      startTransition(() => setStep(TOTAL_STEPS));
    } catch (err) {
      const status = isSdkError(err) ? err.status : undefined;
      if (status === 402) {
        setError(
          "You have unpaid invoices from a previous subscription. Open the billing portal to settle them before subscribing again.",
        );
      } else {
        setError(
          "We couldn't complete your subscription. Please try again in a few moments.",
        );
      }
    }
  };

  const handleClose = () => {
    if (step === TOTAL_STEPS && onSuccess) onSuccess();
    onClose();
  };

  // Block ESC/backdrop while a subscription request is in flight, and on the
  // success step (parent should drive closing via the "Done" button).
  const canClose = useCallback(
    () => !createSubscription.isPending && step < TOTAL_STEPS,
    [createSubscription.isPending, step],
  );

  return (
    <BaseDialog
      open={open}
      onClose={handleClose}
      canClose={canClose}
      size="xl"
      aria-label="Subscribe to ShellHub Cloud"
      className="sm:max-h-[85vh]"
    >
      <span role="status" aria-live="polite" className="sr-only">
        {`Step ${step} of ${TOTAL_STEPS}: ${STEPS[step - 1]}`}
      </span>

      {/* Header: progress dots + close button */}
      <header className="flex items-center justify-between px-6 pt-5 pb-0 shrink-0">
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
            type="button"
            onClick={handleClose}
            disabled={createSubscription.isPending}
            className="p-1.5 rounded-lg text-text-muted hover:text-text-primary hover:bg-hover-medium transition-all disabled:opacity-40"
            aria-label="Close wizard"
          >
            <XMarkIcon className="w-4 h-4" />
          </button>
        ) : (
          <div className="w-7 h-7" aria-hidden="true" />
        )}
      </header>

      {/* Thin progress bar */}
      <div
        role="progressbar"
        aria-valuenow={step}
        aria-valuemin={1}
        aria-valuemax={TOTAL_STEPS}
        aria-label={`Step ${step} of ${TOTAL_STEPS}`}
        className="mx-6 mt-4 mb-5 h-px bg-border shrink-0 overflow-hidden rounded-full"
      >
        <div
          className="h-full bg-primary transition-all duration-500"
          style={{ width: `${(step / TOTAL_STEPS) * 100}%` }}
        />
      </div>

      {/* Content */}
      <main className="flex-auto overflow-y-auto px-6 min-h-0">
        {step === 1 && <BillingLetter />}
        {step === 2 && (
          <BillingPayment
            onHasDefault={() => setHasDefault(true)}
            onNoPaymentMethods={() => setHasDefault(false)}
          />
        )}
        {step === 3 && <BillingCheckout />}
        {step === 4 && <BillingSuccessful />}
      </main>

      {error && (
        <div
          role="alert"
          className="mx-6 mt-3 px-3 py-2 rounded-lg bg-accent-red/10 border border-accent-red/20 text-xs text-accent-red"
        >
          {error}
        </div>
      )}

      {/* Footer */}
      <footer className="px-6 py-4 mt-4 border-t border-border shrink-0 flex items-center justify-between">
        {step > 1 && step < TOTAL_STEPS ? (
          <button
            type="button"
            onClick={goBack}
            disabled={createSubscription.isPending}
            className="inline-flex items-center gap-1 text-xs font-medium text-text-muted hover:text-text-secondary transition-colors disabled:opacity-40"
          >
            <ChevronLeftIcon className="w-3.5 h-3.5" strokeWidth={2.5} />
            Back
          </button>
        ) : (
          <span />
        )}

        <div className="flex items-center gap-3">
          {step < TOTAL_STEPS && (
            <button
              type="button"
              onClick={handleClose}
              disabled={createSubscription.isPending}
              className="px-4 py-2 rounded-lg text-xs font-medium text-text-muted hover:text-text-secondary hover:bg-hover-medium transition-all disabled:opacity-40"
            >
              Close
            </button>
          )}

          {step === 1 && (
            <PrimaryButton onClick={goNext}>
              Next <ArrowRightIcon className="w-3.5 h-3.5" strokeWidth={2.5} />
            </PrimaryButton>
          )}

          {step === 2 && (
            <PrimaryButton onClick={goNext} disabled={!hasDefault}>
              Next <ArrowRightIcon className="w-3.5 h-3.5" strokeWidth={2.5} />
            </PrimaryButton>
          )}

          {step === 3 && (
            <PrimaryButton
              onClick={() => void subscribe()}
              disabled={createSubscription.isPending}
              loading={createSubscription.isPending}
            >
              {createSubscription.isPending
                ? "Subscribing…"
                : "Confirm subscription"}
            </PrimaryButton>
          )}

          {step === TOTAL_STEPS && (
            <PrimaryButton onClick={handleClose}>
              Done <ArrowRightIcon className="w-3.5 h-3.5" strokeWidth={2.5} />
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
  children: ReactNode;
  onClick?: () => void;
  disabled?: boolean;
  loading?: boolean;
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      disabled={disabled || loading}
      className="inline-flex items-center gap-1.5 px-4 py-2 rounded-lg text-xs font-semibold transition-all duration-200
        bg-primary text-white hover:bg-primary-600 active:scale-[0.98]
        disabled:opacity-40 disabled:cursor-not-allowed disabled:hover:bg-primary disabled:active:scale-100"
    >
      {loading && (
        <span
          aria-hidden="true"
          className="w-3.5 h-3.5 border-2 border-white/30 border-t-white rounded-full animate-spin"
        />
      )}
      {children}
    </button>
  );
}
