import { useState } from "react";
import { useResetOnOpen } from "@/hooks/useResetOnOpen";
import ConfirmDialog from "@/components/common/ConfirmDialog";
import InputField from "@/components/common/fields/InputField";

/**
 * Enrols this browser as an SSH identity. The key is generated here and its private half never
 * leaves the browser, which is what the dialog has to make clear before it is accepted.
 */
export default function BrowserEnrollDialog({
  open,
  defaultName,
  onClose,
  onConfirm,
}: {
  open: boolean;
  defaultName: string;
  onClose: () => void;
  onConfirm: (name: string) => Promise<void>;
}) {
  const [name, setName] = useState(defaultName);
  const [error, setError] = useState<string | null>(null);

  useResetOnOpen(open, () => {
    setName(defaultName);
    setError(null);
  });

  const handleConfirm = async () => {
    setError(null);
    try {
      await onConfirm(name.trim() || defaultName);
    } catch {
      setError("Could not register this browser. Please try again.");
    }
  };

  return (
    <ConfirmDialog
      open={open}
      onClose={onClose}
      onConfirm={handleConfirm}
      title="Register this browser"
      description="To connect from here, ShellHub registers this browser as an SSH identity — a key that never leaves it. You can revoke it anytime in SSH Identities."
      confirmLabel="Register and connect"
      variant="primary"
      errorMessage={error}
    >
      <InputField
        id="browser-enroll-name"
        label="Name"
        value={name}
        onChange={setName}
        placeholder={defaultName}
      />
    </ConfirmDialog>
  );
}
