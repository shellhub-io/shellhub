import { useState, useId } from "react";
import { ExclamationTriangleIcon } from "@heroicons/react/24/outline";
import { useResetOnOpen } from "@/hooks/useResetOnOpen";
import { useCreateActivationToken } from "@/hooks/useMemberMutations";
import CopyButton from "@/components/common/CopyButton";
import BaseDialog from "@/components/common/BaseDialog";
import InputField from "@/components/common/fields/InputField";
import { Button } from "@shellhub/design-system/primitives";
import { buildActivationLink } from "./helpers";

interface ActivationLinkDialogProps {
  open: boolean;
  onClose: () => void;
  userId: string;
  email: string;
}

export default function ActivationLinkDialog({
  open,
  onClose,
  userId,
  email,
}: ActivationLinkDialogProps) {
  const createToken = useCreateActivationToken();
  const [step, setStep] = useState<"confirm" | "result">("confirm");
  const [link, setLink] = useState("");
  const [error, setError] = useState("");

  const autoId = useId();
  const titleId = `activation-link-title-${autoId}`;
  const descId = `activation-link-desc-${autoId}`;

  useResetOnOpen(open, () => {
    setStep("confirm");
    setLink("");
    setError("");
  });

  const handleGenerate = async () => {
    setError("");
    try {
      const data = await createToken.mutateAsync({ path: { id: userId } });
      setLink(buildActivationLink(userId, data?.token ?? ""));
      setStep("result");
    } catch {
      setError("Failed to generate the activation link. Please try again.");
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
          <div className="p-6 pb-0">
            <h2
              id={titleId}
              className="text-base font-semibold text-text-primary"
            >
              Generate Activation Link
            </h2>
          </div>

          <div className="px-6 pt-2 pb-6">
            <p id={descId} className="text-sm text-text-muted mb-6">
              Generate a one-time link for{" "}
              <span className="font-medium text-text-primary">{email}</span> to
              set a password and activate their account. Share it directly with
              them; it expires after a while and can only be used once.
            </p>
            {error && (
              <p role="alert" className="text-2xs text-accent-red mb-4">
                {error}
              </p>
            )}
          </div>

          <div className="flex justify-end gap-2 px-6 py-4 border-t border-border">
            <Button variant="ghost" onClick={onClose}>
              Cancel
            </Button>
            <Button
              variant="primary"
              onClick={() => void handleGenerate()}
              disabled={createToken.isPending}
              loading={createToken.isPending}
            >
              Generate
            </Button>
          </div>
        </>
      ) : (
        <>
          <div className="p-6 pb-0">
            <h2
              id={titleId}
              className="text-base font-semibold text-text-primary"
            >
              Activation Link
            </h2>
          </div>

          <div className="px-6 pt-2 pb-6">
            <div className="flex items-start gap-2 p-3 bg-accent-yellow/8 border border-accent-yellow/20 rounded-lg mb-4">
              <ExclamationTriangleIcon
                className="w-4 h-4 text-accent-yellow shrink-0 mt-0.5"
                strokeWidth={2}
              />
              <p id={descId} className="text-2xs text-accent-yellow">
                Copy this link now and share it with the user. It will not be
                shown again.
              </p>
            </div>
            <div className="flex items-center gap-2">
              <div className="flex-1">
                <InputField
                  id={`${autoId}-activation-link`}
                  label="Activation link"
                  hideLabel
                  readOnly
                  value={link}
                  onChange={() => {}}
                  variant="mono"
                />
              </div>
              <CopyButton text={link} size="md" showLabel />
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
