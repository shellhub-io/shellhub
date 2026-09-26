import { useState, FormEvent } from "react";
import {
  ExclamationCircleIcon,
  PencilSquareIcon,
  FingerPrintIcon,
} from "@heroicons/react/24/outline";
import { Button } from "@shellhub/design-system/primitives";
import { useResetOnOpen } from "@/hooks/useResetOnOpen";
import {
  useCreateSSHIdentity,
  useRenameSSHIdentity,
} from "@/hooks/useSSHIdentityMutations";
import { useCreateApiKeySSHIdentity } from "@/hooks/useSSHIdentityMutations";
import type { SshIdentity } from "@/client";
import { isPublicKeyValid } from "@/utils/sshKeys";
import Modal from "@/components/common/Modal";
import InputField from "@/components/common/fields/InputField";
import KeyFileInput from "@/components/common/fields/KeyFileInput";
import IdentityLifecycleFields from "@/components/common/IdentityLifecycleFields";
import {
  keyExpiryPayload,
  identityLifecyclePayload,
  sshIdentitySource,
  isAlreadyEnrolled,
} from "@/utils/sshIdentity";
import KeyExpiryField from "@/components/common/KeyExpiryField";
import { useBrowserKeyFingerprint } from "@/hooks/useBrowserKey";

/**
 * Enrols or renames an SSH identity. On edit only the name changes — the key is what the
 * identity is.
 *
 * apiKeyName switches the owner from the caller to that API key, which is how an automation
 * gets a credential. The owner is never a choice here: the identities page enrols for the
 * person, and an API key's own screen enrols for the key, so each caller already knows.
 */
function IdentityModal({
  open,
  editIdentity,
  apiKeyName,
  onClose,
}: {
  open: boolean;
  editIdentity: SshIdentity | null;
  apiKeyName?: string;
  onClose: () => void;
}) {
  const createIdentity = useCreateSSHIdentity();
  const renameIdentity = useRenameSSHIdentity();
  const createApiKeyIdentity = useCreateApiKeySSHIdentity();
  const browserKeyFingerprint = useBrowserKeyFingerprint();
  const isEdit = !!editIdentity;

  const [name, setName] = useState("");
  const [keyData, setKeyData] = useState("");
  const [keyError, setKeyError] = useState<string | null>(null);
  const [expiresIn, setExpiresIn] = useState("-1");
  const [singleUse, setSingleUse] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useResetOnOpen(open, () => {
    setName(editIdentity?.name ?? "");
    setKeyData("");
    setKeyError(null);
    setExpiresIn("-1");
    setSingleUse(false);
    setSubmitting(false);
    setError(null);
  });
  const isAPIKey = !isEdit && !!apiKeyName;

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
      } else if (isAPIKey) {
        await createApiKeyIdentity.mutateAsync({
          path: { name: apiKeyName ?? "" },
          body: {
            name: name.trim(),
            data: keyData.trim(),
            ...identityLifecyclePayload(expiresIn, singleUse),
          },
        });
      } else {
        await createIdentity.mutateAsync({
          body: {
            name: name.trim(),
            data: keyData.trim(),
            ...keyExpiryPayload(expiresIn),
          },
        });
      }
      onClose();
    } catch (err: unknown) {
      if (!isEdit && isAlreadyEnrolled(err)) {
        setKeyError("This key is already an identity in this namespace.");
      } else {
        setError(
          err instanceof Error
            ? err.message
            : `Failed to ${isEdit ? "rename" : "add"} key`,
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
      : "Add Key";

  return (
    <Modal
      open={open}
      onClose={onClose}
      icon={isEdit ? <PencilSquareIcon /> : <FingerPrintIcon />}
      title={isEdit ? "Rename key" : "Add a key"}
      description={
        isEdit
          ? "Only the name changes. The key stays the same."
          : "Add an SSH key to connect with."
      }
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
        <InputField
          id="ssh-identity-name"
          label="Name"
          value={name}
          onChange={setName}
          placeholder={
            isAPIKey
              ? "Name used to identify the key, e.g. deploy"
              : "Name used to identify the key, e.g. laptop"
          }
        />

        {isEdit && editIdentity && (
          <p className="text-xs text-text-muted">
            {
              sshIdentitySource(
                editIdentity.source,
                editIdentity.fingerprint === browserKeyFingerprint,
              ).description
            }
          </p>
        )}

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
              isAPIKey
                ? "Paste the OpenSSH public key the automation will connect with."
                : "Paste an OpenSSH public key to add it ahead of time (e.g. a CI or server key)."
            }
          />
        )}

        {!isEdit && !isAPIKey && (
          <KeyExpiryField
            expiresIn={expiresIn}
            onExpiresInChange={setExpiresIn}
          />
        )}

        {isAPIKey && (
          <IdentityLifecycleFields
            expiresIn={expiresIn}
            onExpiresInChange={setExpiresIn}
            singleUse={singleUse}
            onSingleUseChange={setSingleUse}
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
    </Modal>
  );
}

export default IdentityModal;
