import { useState, useId } from "react";
import { ExclamationTriangleIcon } from "@heroicons/react/24/outline";
import { useResetOnOpen } from "@/hooks/useResetOnOpen";
import { useResetUserPassword } from "@/hooks/useAdminUserMutations";
import { isSdkError } from "@/api/errors";
import CopyButton from "@/components/common/CopyButton";
import BaseDialog from "@/components/common/BaseDialog";
import InputField from "@/components/common/fields/InputField";
import { Button } from "@shellhub/design-system/primitives";

interface ResetPasswordDialogProps {
  open: boolean;
  onClose: () => void;
  userId: string;
}

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
          {/* Header */}
          <div className="p-6 pb-0">
            <h2
              id={titleId}
              className="text-base font-semibold text-text-primary"
            >
              Enable Local Authentication
            </h2>
          </div>

          {/* Body */}
          <div className="px-6 pt-2 pb-6">
            <p id={descId} className="text-sm text-text-muted mb-6">
              This will generate a temporary password for this SAML-only user,
              enabling them to log in with local credentials. They should change
              this password after their first login.
            </p>
            {error && (
              <p role="alert" className="text-2xs text-accent-red mb-4">
                {error}
              </p>
            )}
          </div>

          {/* Footer */}
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
          {/* Header */}
          <div className="p-6 pb-0">
            <h2
              id={titleId}
              className="text-base font-semibold text-text-primary"
            >
              Password Generated
            </h2>
          </div>

          {/* Body */}
          <div className="px-6 pt-2 pb-6">
            <div className="flex items-start gap-2 p-3 bg-accent-yellow/8 border border-accent-yellow/20 rounded-lg mb-4">
              <ExclamationTriangleIcon
                className="w-4 h-4 text-accent-yellow shrink-0 mt-0.5"
                strokeWidth={2}
              />
              <p id={descId} className="text-2xs text-accent-yellow">
                Make sure to copy this password now. It will not be shown again.
              </p>
            </div>
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

          {/* Footer */}
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
