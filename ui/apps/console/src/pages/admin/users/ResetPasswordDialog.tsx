import { useState, useId } from "react";
import {
  KeyIcon,
  CheckCircleIcon,
} from "@heroicons/react/24/outline";
import { useResetOnOpen } from "@/hooks/useResetOnOpen";
import { useResetUserPassword } from "@/hooks/useAdminUserMutations";
import { isSdkError } from "@/api/errors";
import CopyButton from "@/components/common/CopyButton";
import BaseDialog from "@/components/common/BaseDialog";
import InputField from "@/components/common/fields/InputField";
import { Button } from "@shellhub/design-system/primitives";
import DialogHeader from "@/components/common/DialogHeader";

interface ResetPasswordDialogProps {
  open: boolean;
  onClose: () => void;
  userId: string;
}

/**
 * Resets a user's password as an admin. The user is not notified, so whoever does this has to
 * tell them — the dialog says as much.
 */
export default function ResetPasswordDialog({
  open,
  onClose,
  userId,
}: ResetPasswordDialogProps) {
  const resetPassword = useResetUserPassword();
  const [step, setStep] = useState<"confirm" | "result">("confirm");
  const [generatedPassword, setGeneratedPassword] = useState("");
  const [error, setError] = useState("");

  const autoId = useId();
  const titleId = `reset-pw-title-${autoId}`;
  const descId = `reset-pw-desc-${autoId}`;

  useResetOnOpen(open, () => {
    setStep("confirm");
    setGeneratedPassword("");
    setError("");
  });

  const handleEnable = async () => {
    setError("");
    try {
      const data = await resetPassword.mutateAsync({ path: { id: userId } });
      setGeneratedPassword(data?.password ?? "");
      setStep("result");
    } catch (err) {
      if (isSdkError(err) && err.status === 400) {
        setError("This user already has a local password.");
      } else {
        setError("Failed to set password. Please try again.");
      }
    }
  };

  return (
    <BaseDialog
      open={open}
      onClose={onClose}
      size="sm"
      aria-labelledby={titleId}
      aria-describedby={descId}
    >
      {step === "confirm" ? (
        <>
          <DialogHeader
            icon={<KeyIcon />}
            title="Enable local authentication"
            description="This SAML-only user gets a temporary password to sign in with. They should change it after their first login."
            titleId={titleId}
            descriptionId={descId}
            onClose={onClose}
          />
          {error && (
            <p role="alert" className="px-6 pb-4 text-2xs text-accent-red">
              {error}
            </p>
          )}

          <div className="flex justify-end gap-2 px-6 py-4 border-t border-border">
            <Button variant="ghost" onClick={onClose}>
              Cancel
            </Button>
            <Button
              variant="primary"
              onClick={() => void handleEnable()}
              disabled={resetPassword.isPending}
              loading={resetPassword.isPending}
            >
              Enable
            </Button>
          </div>
        </>
      ) : (
        <>
          <DialogHeader
            icon={<CheckCircleIcon />}
            iconColor="green"
            title="Password generated"
            description="Copy it now. It won't be shown again."
            titleId={titleId}
            descriptionId={descId}
          />
          <div className="px-6 pb-6">
            <div className="flex items-center gap-2">
              <div className="flex-1">
                <InputField
                  id={`${autoId}-generated-password`}
                  label="Generated password"
                  hideLabel
                  readOnly
                  value={generatedPassword}
                  onChange={() => {}}
                  variant="mono"
                />
              </div>
              <CopyButton text={generatedPassword} size="md" showLabel />
            </div>
          </div>

          <div className="flex justify-end px-6 py-4 border-t border-border">
            <Button variant="primary" onClick={onClose}>
              Close
            </Button>
          </div>
        </>
      )}
    </BaseDialog>
  );
}
