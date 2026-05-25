import { useState, FormEvent } from "react";
import { useResetOnOpen } from "@/hooks/useResetOnOpen";
import { ExclamationCircleIcon } from "@heroicons/react/24/outline";
import { useVaultStore, DuplicateKeyError } from "@/stores/vaultStore";
import {
  validatePrivateKey,
  getFingerprint,
  getAlgorithm,
} from "@/utils/ssh-keys";
import Drawer from "@/components/common/Drawer";
import KeyFileInput from "@/components/common/fields/KeyFileInput";
import InputField from "@/components/common/fields/InputField";
import PasswordField from "@/components/common/fields/PasswordField";
import type { VaultKeyEntry } from "@/types/vault";
import Spinner from "@/components/common/Spinner";

interface Props {
  open: boolean;
  editKey: VaultKeyEntry | null;
  onClose: () => void;
}

export default function KeyDrawer({ open, editKey, onClose }: Props) {
  const addKey = useVaultStore((s) => s.addKey);
  const updateKey = useVaultStore((s) => s.updateKey);
  const isEdit = !!editKey;

  const [name, setName] = useState("");
  const [nameError, setNameError] = useState<string | null>(null);
  const [keyData, setKeyData] = useState("");
  const [encrypted, setEncrypted] = useState(false);
  const [passphrase, setPassphrase] = useState("");
  const [keyError, setKeyError] = useState<string | null>(null);
  const [passphraseError, setPassphraseError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useResetOnOpen(open, () => {
    setName(editKey?.name ?? "");
    setNameError(null);
    setKeyData(editKey?.data ?? "");
    setEncrypted(editKey?.hasPassphrase ?? false);
    setPassphrase("");
    setKeyError(null);
    setPassphraseError(null);
    setSubmitting(false);
    setError(null);
  });

  const handleNameChange = (value: string) => {
    setName(value);
    if (nameError) setNameError(null);
  };

  const handleKeyChange = (pem: string) => {
    setKeyData(pem);
    setPassphrase("");
    setPassphraseError(null);

    if (!pem.trim()) {
      setKeyError(null);
      setEncrypted(false);
      return;
    }

    const result = validatePrivateKey(pem.trim());
    if (!result.valid) {
      setKeyError(result.error);
      setEncrypted(false);
      return;
    }

    setKeyError(null);
    setEncrypted(result.encrypted);
  };

  const handleFileName = (fileName: string) => {
    if (!name) setName(fileName);
  };

  const validateForPaste = (text: string) =>
    validatePrivateKey(text.trim()).valid;

  const handlePassphraseChange = (value: string) => {
    setPassphrase(value);
    if (passphraseError) setPassphraseError(null);
  };

  const canSubmit =
    name.trim() &&
    !nameError &&
    keyData.trim() &&
    !keyError &&
    (!encrypted || passphrase.trim()) &&
    !passphraseError &&
    !submitting;

  const handleSubmit = async (e?: FormEvent) => {
    e?.preventDefault();
    if (!canSubmit) return;

    let fingerprint: string;
    let algorithm: string;
    try {
      const pp = encrypted ? passphrase : undefined;
      fingerprint = getFingerprint(keyData.trim(), pp);
      algorithm = getAlgorithm(keyData.trim(), pp);
    } catch (err) {
      if (encrypted) {
        const errName = err instanceof Error ? err.name : undefined;
        if (errName === "KeyParseError") {
          setPassphraseError("Incorrect passphrase");
        } else {
          setPassphraseError("Could not decrypt key with this passphrase");
        }
      } else {
        setKeyError("Failed to read private key");
      }
      return;
    }

    setError(null);
    setSubmitting(true);
    try {
      if (isEdit && editKey) {
        await updateKey(editKey.id, {
          name: name.trim(),
          data: keyData.trim(),
          hasPassphrase: encrypted,
          fingerprint,
          algorithm,
        });
      } else {
        await addKey({
          name: name.trim(),
          data: keyData.trim(),
          hasPassphrase: encrypted,
          fingerprint,
          algorithm,
        });
      }
      onClose();
    } catch (err: unknown) {
      if (err instanceof DuplicateKeyError) {
        if (err.field === "both" || err.field === "name") {
          setNameError("Name is already used");
        }
        if (err.field === "both" || err.field === "private_key") {
          setKeyError("Private key is already stored");
        }
      } else {
        const msg = err instanceof Error ? err.message : "";
        setError(msg || `Failed to ${isEdit ? "update" : "add"} key`);
      }
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Drawer
      open={open}
      onClose={onClose}
      title={isEdit ? "Edit Private Key" : "Add Private Key"}
      footer={
        <>
          <button
            type="button"
            onClick={onClose}
            className="px-4 py-2.5 text-sm font-medium text-text-secondary hover:text-text-primary rounded-lg hover:bg-hover-subtle transition-colors"
          >
            Cancel
          </button>
          <button
            type="submit"
            onClick={() => void handleSubmit()}
            disabled={!canSubmit}
            className="px-5 py-2.5 bg-primary hover:bg-primary-600 text-white rounded-lg text-sm font-semibold disabled:opacity-dim disabled:cursor-not-allowed transition-all flex items-center gap-2"
          >
            {submitting ? (
              <>
                <Spinner size="sm" tone="onPrimary" />
                Saving...
              </>
            ) : isEdit ? (
              "Save Changes"
            ) : (
              "Add Key"
            )}
          </button>
        </>
      }
    >
      <form onSubmit={(e) => void handleSubmit(e)} className="space-y-5">
        <InputField
          id="key-name"
          label="Name"
          value={name}
          onChange={handleNameChange}
          maxLength={255}
          placeholder="e.g. Production Server"
          autoFocus={open}
          error={nameError ?? undefined}
        />

        <KeyFileInput
          label="Private Key"
          id="key-data"
          value={keyData}
          onChange={handleKeyChange}
          validate={validateForPaste}
          onFileName={handleFileName}
          disabled={isEdit}
          error={keyError}
          placeholder={"-----BEGIN OPENSSH PRIVATE KEY-----\n..."}
          rows={8}
          hint="RSA, DSA, ECDSA, ED25519 — PEM and OpenSSH formats."
          loadedLabel={`Key loaded${encrypted ? " (encrypted)" : ""}`}
          emptyLabel="Drop private key file, paste, or browse"
        />

        {encrypted && (
          <PasswordField
            id="key-passphrase"
            label="Passphrase"
            value={passphrase}
            onChange={handlePassphraseChange}
            placeholder="Enter passphrase for encrypted key"
            autoComplete="off"
            error={passphraseError ?? undefined}
            hint="This key is encrypted. The passphrase is not stored."
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
