import { useState } from "react";
import {
  FingerPrintIcon,
  ExclamationCircleIcon,
} from "@heroicons/react/24/outline";
import { Button } from "@shellhub/design-system/primitives";
import { useResetOnOpen } from "@/hooks/useResetOnOpen";
import BaseDialog from "@/components/common/BaseDialog";
import InputField from "@/components/common/fields/InputField";

/**
 * First-run consent shown the first time a browser connects in identity access
 * mode. It explains that ShellHub is about to register this browser as an SSH
 * identity — a non-extractable key that never leaves it — and lets the user name
 * it before enrolling. Shown once per browser; later connects reuse the stored
 * key and skip it.
 */
export default function BrowserEnrollDialog({
  open,
  defaultName,
  onClose,
  onConfirm,
}: {
  open: boolean;
  /** Prefilled identity name, e.g. "Chrome 149 on Linux". Editable. */
  defaultName: string;
  onClose: () => void;
  onConfirm: (name: string) => Promise<void>;
}) {
  const [name, setName] = useState(defaultName);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useResetOnOpen(open, () => {
    setName(defaultName);
    setSubmitting(false);
    setError(null);
  });

  const handleConfirm = async () => {
    if (submitting) return;
    setSubmitting(true);
    setError(null);
    try {
      await onConfirm(name.trim() || defaultName);
    } catch {
      setError("Could not register this browser. Please try again.");
      setSubmitting(false);
    }
  };

  if (!open) return null;

  return (
    <BaseDialog
      open={open}
      onClose={onClose}
      size="sm"
      aria-labelledby="browser-enroll-title"
    >
      <div className="p-6 space-y-5">
        <div className="flex items-start gap-3">
          <span className="grid place-items-center w-9 h-9 rounded-lg bg-primary/10 text-primary shrink-0">
            <FingerPrintIcon className="w-5 h-5" strokeWidth={2} />
          </span>
          <div>
            <h2
              id="browser-enroll-title"
              className="text-base font-semibold text-text-primary"
            >
              Register this browser
            </h2>
            <p className="text-sm text-text-muted mt-0.5">
              To connect from here, ShellHub registers this browser as an SSH
              identity — a key that never leaves it. You can revoke it anytime
              in SSH Identities.
            </p>
          </div>
        </div>

        <InputField
          id="browser-enroll-name"
          label="Name"
          value={name}
          onChange={setName}
          placeholder={defaultName}
        />

        {error && (
          <p className="text-xs font-mono text-accent-red flex items-center gap-1.5">
            <ExclamationCircleIcon
              className="w-3.5 h-3.5 shrink-0"
              strokeWidth={2}
            />
            {error}
          </p>
        )}

        <div className="flex justify-end gap-2">
          <Button
            variant="ghost"
            type="button"
            onClick={onClose}
            disabled={submitting}
          >
            Cancel
          </Button>
          <Button
            variant="primary"
            type="button"
            onClick={() => void handleConfirm()}
            loading={submitting}
            disabled={submitting}
          >
            {submitting ? "Registering..." : "Register and connect"}
          </Button>
        </div>
      </div>
    </BaseDialog>
  );
}
