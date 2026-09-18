import { useState, FormEvent } from "react";
import {
  ExclamationCircleIcon,
  UserIcon,
  CpuChipIcon,
} from "@heroicons/react/24/outline";
import { Button } from "@shellhub/design-system/primitives";
import { useResetOnOpen } from "@/hooks/useResetOnOpen";
import {
  useCreateSSHIdentity,
  useRenameSSHIdentity,
} from "@/hooks/useSSHIdentityMutations";
import { useCreateApiKeySSHIdentity } from "@/hooks/useSSHIdentityMutations";
import { useApiKeys } from "@/hooks/useApiKeys";
import { useHasPermission } from "@/hooks/useHasPermission";
import type { SshIdentity } from "@/client";
import { isPublicKeyValid } from "@/utils/sshKeys";
import Drawer from "@/components/common/Drawer";
import InputField from "@/components/common/fields/InputField";
import KeyFileInput from "@/components/common/fields/KeyFileInput";
import RadioCard from "@/components/common/fields/RadioCard";
import RadioGroupField from "@/components/common/fields/RadioGroupField";
import IdentityLifecycleFields from "@/components/common/IdentityLifecycleFields";
import {
  keyExpiryPayload,
  identityLifecyclePayload,
  sshIdentitySource,
  isAlreadyEnrolled,
} from "@/utils/sshIdentity";
import KeyExpiryField from "@/components/common/KeyExpiryField";
import { useBrowserKeyFingerprint } from "@/hooks/useBrowserKey";
import { INPUT, LABEL } from "@/utils/styles";

// Who a newly added key belongs to: the caller, or an API key. Enrolling for an API key gives
// an automation its own credential instead of binding the key to a person. Offered only to
// callers who may manage identities; rename never shows it.
type Target = "self" | "api-key";

/**
 * Enrols or renames an SSH identity. On edit only the name changes — the key is what the
 * identity is.
 */
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
  const createApiKeyIdentity = useCreateApiKeySSHIdentity();
  const canManageIdentities = useHasPermission("sshIdentity:manage");
  const { apiKeys } = useApiKeys({ perPage: 100 });
  const browserKeyFingerprint = useBrowserKeyFingerprint();
  const isEdit = !!editIdentity;

  const [target, setTarget] = useState<Target>("self");
  const [name, setName] = useState("");
  const [keyData, setKeyData] = useState("");
  const [keyError, setKeyError] = useState<string | null>(null);
  const [apiKeyName, setApiKeyName] = useState("");
  const [expiresIn, setExpiresIn] = useState("-1");
  const [singleUse, setSingleUse] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useResetOnOpen(open, () => {
    setTarget("self");
    setName(editIdentity?.name ?? "");
    setKeyData("");
    setKeyError(null);
    setApiKeyName("");
    setExpiresIn("-1");
    setSingleUse(false);
    setSubmitting(false);
    setError(null);
  });
  const isAPIKey = !isEdit && target === "api-key";

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
    : !name.trim() || !keyData.trim() || !!keyError || (isAPIKey && !apiKeyName);

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
          path: { name: apiKeyName },
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
        {!isEdit && canManageIdentities && (
          <RadioGroupField
            label="Add this key for"
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
              value="api-key"
              icon={<CpuChipIcon className="w-4 h-4" />}
              label="An API key"
              description="An automation connects with it. Where it may reach is set by access policies."
            />
          </RadioGroupField>
        )}

        {isAPIKey && (
          <div>
            <label htmlFor="ssh-identity-api-key" className={LABEL}>
              API key
            </label>
            <select
              id="ssh-identity-api-key"
              value={apiKeyName}
              onChange={(e) => setApiKeyName(e.target.value)}
              className={INPUT}
            >
              <option value="">Choose an API key...</option>
              {apiKeys.map((key) => (
                <option key={key.id} value={key.name}>
                  {key.name}
                </option>
              ))}
            </select>
          </div>
        )}

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
    </Drawer>
  );
}

export default IdentityDrawer;
