import { useState, FormEvent } from "react";
import {
  ExclamationCircleIcon,
  UserIcon,
  CpuChipIcon,
} from "@heroicons/react/24/outline";
import { Button } from "@shellhub/design-system/primitives";
import { isSdkError } from "@/api/errors";
import { useResetOnOpen } from "@/hooks/useResetOnOpen";
import {
  useCreateSSHIdentity,
  useRenameSSHIdentity,
} from "@/hooks/useSSHIdentityMutations";
import { useCreateServiceAccount } from "@/hooks/useServiceAccountMutations";
import { useHasPermission } from "@/hooks/useHasPermission";
import type { SshIdentity } from "@/client";
import { isPublicKeyValid } from "@/utils/sshKeys";
import Drawer from "@/components/common/Drawer";
import InputField from "@/components/common/fields/InputField";
import KeyFileInput from "@/components/common/fields/KeyFileInput";
import RadioCard from "@/components/common/fields/RadioCard";
import RadioGroupField from "@/components/common/fields/RadioGroupField";

// Who a newly added key belongs to: the caller, or a new service account. Enrolling for a
// service account gives an automated system its own identity instead of binding the key to a
// person. Only offered to callers who can create service accounts; rename never shows it.
type Target = "self" | "service-account";

function IdentityDrawer({
  open,
  editIdentity,
  onClose,
}: {
  open: boolean;
  editIdentity: SshIdentity | null;
  onClose: () => void;
}) {
  const createIdentity = useCreateSSHIdentity();
  const renameIdentity = useRenameSSHIdentity();
  const createServiceAccount = useCreateServiceAccount();
  const canCreateServiceAccount = useHasPermission("serviceAccount:create");
  const isEdit = !!editIdentity;

  const [target, setTarget] = useState<Target>("self");
  const [name, setName] = useState("");
  const [keyData, setKeyData] = useState("");
  const [keyError, setKeyError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useResetOnOpen(open, () => {
    setTarget("self");
    setName(editIdentity?.name ?? "");
    setKeyData("");
    setKeyError(null);
    setSubmitting(false);
    setError(null);
  });

  const isServiceAccount = !isEdit && target === "service-account";

  const handleKeyDataChange = (v: string) => {
    setKeyData(v);
    if (v && !isPublicKeyValid(v))
      setKeyError("This is not a valid public key.");
    else setKeyError(null);
  };

  const handleFileName = (filename: string) => {
    if (!name) setName(filename || "Imported key");
  };

  const confirmDisabled = isEdit
    ? !name.trim()
    : !name.trim() || !keyData.trim() || !!keyError;

  const handleSubmit = async (e?: FormEvent) => {
    e?.preventDefault();
    if (confirmDisabled) return;
    setError(null);
    setSubmitting(true);
    try {
      if (isEdit && editIdentity) {
        await renameIdentity.mutateAsync({
          path: { id: editIdentity.id },
          body: { name: name.trim() },
        });
      } else if (isServiceAccount) {
        await createServiceAccount.mutateAsync({
          body: { name: name.trim(), data: keyData.trim() },
        });
      } else {
        await createIdentity.mutateAsync({
          body: { name: name.trim(), data: keyData.trim() },
        });
      }
      onClose();
    } catch (err: unknown) {
      if (!isEdit && isSdkError(err) && err.status === 409) {
        setKeyError("This key is already enrolled in this namespace.");
      } else {
        setError(
          err instanceof Error
            ? err.message
            : `Failed to ${isEdit ? "rename" : "enroll"} key`,
        );
      }
    } finally {
      setSubmitting(false);
    }
  };

  const submitLabel = submitting
    ? "Saving..."
    : isEdit
      ? "Save Changes"
      : isServiceAccount
        ? "Create Service Account"
        : "Add Key";

  return (
    <Drawer
      open={open}
      onClose={onClose}
      title={isEdit ? "Rename Key" : "Add a Key"}
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
            {submitLabel}
          </Button>
        </>
      }
    >
      <form onSubmit={(e) => void handleSubmit(e)} className="space-y-5">
        {!isEdit && canCreateServiceAccount && (
          <RadioGroupField
            label="Enroll this key for"
            value={target}
            onChange={setTarget}
          >
            <RadioCard
              value="self"
              icon={<UserIcon className="w-4 h-4" />}
              label="Myself"
              description="The key becomes your own identity."
            />
            <RadioCard
              value="service-account"
              icon={<CpuChipIcon className="w-4 h-4" />}
              label="A new service account"
              description="A non-human identity for an automated system, separate from you."
            />
          </RadioGroupField>
        )}

        <InputField
          id="ssh-identity-name"
          label={isServiceAccount ? "Service account name" : "Name"}
          value={name}
          onChange={setName}
          placeholder={
            isServiceAccount
              ? "Name for the service account, e.g. ci-bot"
              : "Name used to identify the key, e.g. laptop"
          }
        />

        {!isEdit && (
          <KeyFileInput
            id="ssh-identity-data"
            label="Public key data"
            value={keyData}
            onChange={handleKeyDataChange}
            validate={isPublicKeyValid}
            onFileName={handleFileName}
            error={keyError || undefined}
            accept=".pub,.pem,.key,.txt"
            placeholder="ssh-ed25519 AAAAC3NzaC1lZDI1NTE5..."
            rows={3}
            hint={
              isServiceAccount
                ? "Paste the OpenSSH public key the automated system will connect with."
                : "Paste an OpenSSH public key to pre-enroll it (e.g. a CI or server key)."
            }
          />
        )}

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

export default IdentityDrawer;
