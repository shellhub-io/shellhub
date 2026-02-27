import { useState, useEffect, FormEvent } from "react";
import { ExclamationCircleIcon } from "@heroicons/react/24/outline";
import { useVaultStore, DuplicateKeyError } from "@/stores/vaultStore";
import { validatePrivateKey, getFingerprint } from "@/utils/ssh-keys";
import Drawer from "@/components/common/Drawer";
import KeyFileInput from "@/components/common/KeyFileInput";
import { LABEL, INPUT } from "@/utils/styles";
import type { VaultKeyEntry } from "@/types/vault";

interface Props {
  open: boolean;
  editKey: VaultKeyEntry | null;
  onClose: () => void;
}

export default function KeyDrawer({ open, editKey, onClose }: Props) {
  const { addKey, updateKey } = useVaultStore();
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

  useEffect(() => {
    if (!open) return;
    if (editKey) {
      setName(editKey.name);
      setKeyData(editKey.data);
      setEncrypted(editKey.hasPassphrase);
    } else {
      setName("");
      setKeyData("");
      setEncrypted(false);
    }
    setNameError(null);
    setPassphrase("");
    setKeyError(null);
    setPassphraseError(null);
    setError(null);
  }, [open, editKey]);

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

  const validateForPaste = (text: string) => validatePrivateKey(text.trim()).valid;

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

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    if (!canSubmit) return;

    let fingerprint: string;
    try {
      fingerprint = getFingerprint(
        keyData.trim(),
        encrypted ? passphrase : undefined,
      );
    } catch (err) {
      if (encrypted) {
        const errName = (err as { name?: string }).name;
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
        });
      } else {
        await addKey({
          name: name.trim(),
          data: keyData.trim(),
          hasPassphrase: encrypted,
          fingerprint,
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
            onClick={handleSubmit}
            disabled={!canSubmit}
            className="px-5 py-2.5 bg-primary hover:bg-primary-600 text-white rounded-lg text-sm font-semibold disabled:opacity-dim disabled:cursor-not-allowed transition-all"
          >
            {submitting ? (
              <span className="flex items-center gap-2">
                <span className="w-3.5 h-3.5 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                Saving...
              </span>
            ) : isEdit ? (
              "Save Changes"
            ) : (
              "Add Key"
            )}
          </button>
        </>
      }
    >
      <form onSubmit={handleSubmit} className="space-y-5">
        <div>
          <label htmlFor="key-name" className={LABEL}>
            Name
          </label>
          <input
            id="key-name"
            type="text"
            value={name}
            onChange={(e) => handleNameChange(e.target.value)}
            maxLength={255}
            placeholder="e.g. Production Server"
            autoFocus={open}
            aria-invalid={!!nameError}
            aria-describedby={nameError ? "key-name-error" : undefined}
            className={INPUT}
          />
          {nameError && (
            <p
              id="key-name-error"
              className="text-2xs text-accent-red mt-1.5 flex items-center gap-1"
            >
              <ExclamationCircleIcon className="w-3.5 h-3.5 shrink-0" />
              {nameError}
            </p>
          )}
        </div>

        <KeyFileInput
          label="Private Key"
          id="key-data"
          value={keyData}
          onChange={handleKeyChange}
          validate={validateForPaste}
          onFileName={handleFileName}
          disabled={isEdit}
          error={keyError}
          accept=".pem,.key,.txt"
          placeholder={"-----BEGIN OPENSSH PRIVATE KEY-----\n..."}
          rows={8}
          hint="RSA, DSA, ECDSA, ED25519 — PEM and OpenSSH formats."
          loadedLabel={`Key loaded${encrypted ? " (encrypted)" : ""}`}
          emptyLabel="Drop private key file, paste, or browse"
        />

        {encrypted && (
          <div>
            <label htmlFor="key-passphrase" className={LABEL}>
              Passphrase
            </label>
            <input
              id="key-passphrase"
              type="password"
              autoComplete="off"
              value={passphrase}
              onChange={(e) => handlePassphraseChange(e.target.value)}
              placeholder="Enter passphrase for encrypted key"
              aria-invalid={!!passphraseError}
              aria-describedby={
                passphraseError ? "key-passphrase-error" : "key-passphrase-hint"
              }
              className={INPUT}
            />
            {passphraseError ? (
              <p
                id="key-passphrase-error"
                className="text-2xs text-accent-red mt-1.5 flex items-center gap-1"
              >
                <ExclamationCircleIcon className="w-3.5 h-3.5 shrink-0" />
                {passphraseError}
              </p>
            ) : (
              <p id="key-passphrase-hint" className="text-2xs text-text-muted mt-1.5">
                This key is encrypted. The passphrase is not stored.
              </p>
            )}
          </div>
        )}

        {error && (
          <p className="text-xs font-mono text-accent-red flex items-center gap-1.5">
            <ExclamationCircleIcon className="w-3.5 h-3.5 shrink-0" strokeWidth={2} />
            {error}
          </p>
        )}
      </form>
    </Drawer>
  );
}
