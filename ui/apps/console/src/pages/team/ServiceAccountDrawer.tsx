import { useState, FormEvent } from "react";
import { ExclamationCircleIcon } from "@heroicons/react/24/outline";
import { Button } from "@shellhub/design-system/primitives";
import { isSdkError } from "@/api/errors";
import { useResetOnOpen } from "@/hooks/useResetOnOpen";
import { useCreateServiceAccount } from "@/hooks/useServiceAccountMutations";
import { isPublicKeyValid } from "@/utils/sshKeys";
import Drawer from "@/components/common/Drawer";
import InputField from "@/components/common/fields/InputField";
import KeyFileInput from "@/components/common/fields/KeyFileInput";

function ServiceAccountDrawer({
  open,
  onClose,
}: {
  open: boolean;
  onClose: () => void;
}) {
  const createServiceAccount = useCreateServiceAccount();

  const [name, setName] = useState("");
  const [keyData, setKeyData] = useState("");
  const [keyError, setKeyError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useResetOnOpen(open, () => {
    setName("");
    setKeyData("");
    setKeyError(null);
    setSubmitting(false);
    setError(null);
  });

  const handleKeyDataChange = (v: string) => {
    setKeyData(v);
    if (v && !isPublicKeyValid(v))
      setKeyError("This is not a valid public key.");
    else setKeyError(null);
  };

  const confirmDisabled = !name.trim() || !keyData.trim() || !!keyError;

  const handleSubmit = async (e?: FormEvent) => {
    e?.preventDefault();
    if (confirmDisabled) return;
    setError(null);
    setSubmitting(true);
    try {
      await createServiceAccount.mutateAsync({
        body: { name: name.trim(), data: keyData.trim() },
      });
      onClose();
    } catch (err: unknown) {
      if (isSdkError(err) && err.status === 409) {
        setKeyError("This key is already enrolled in this namespace.");
      } else {
        setError(
          err instanceof Error
            ? err.message
            : "Failed to create service account",
        );
      }
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Drawer
      open={open}
      onClose={onClose}
      title="Add a Service Account"
      footer={
        <>
          <Button variant="ghost" onClick={onClose}>
            Cancel
          </Button>
          <Button
            variant="primary"
            type="submit"
            onClick={() => void handleSubmit()}
            disabled={submitting || confirmDisabled}
            loading={submitting}
          >
            {submitting ? "Creating..." : "Create"}
          </Button>
        </>
      }
    >
      <form onSubmit={(e) => void handleSubmit(e)} className="space-y-5">
        <InputField
          id="service-account-name"
          label="Name"
          value={name}
          onChange={setName}
          placeholder="Name used to identify the service account, e.g. ci-bot"
        />

        <KeyFileInput
          id="service-account-data"
          label="Public key data"
          value={keyData}
          onChange={handleKeyDataChange}
          validate={isPublicKeyValid}
          error={keyError || undefined}
          accept=".pub,.pem,.key,.txt"
          placeholder="ssh-ed25519 AAAAC3NzaC1lZDI1NTE5..."
          rows={3}
          hint="Paste the OpenSSH public key the automated system will connect with."
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
      </form>
    </Drawer>
  );
}

export default ServiceAccountDrawer;
