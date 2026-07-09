import { useWatch, useController } from "react-hook-form";
import { useVaultStore, DuplicateKeyError } from "@/stores/vaultStore";
import {
  validatePrivateKey,
  getFingerprint,
  getAlgorithm,
} from "@/utils/sshKeys";
import FormDrawer from "@/components/common/FormDrawer";
import KeyFileInput from "@/components/common/fields/KeyFileInput";
import {
  FormInputField,
  FormPasswordField,
} from "@/components/common/fields/rhf";
import type { VaultKeyEntry } from "@/types/vault";
import { useDrawerForm } from "@/hooks/useDrawerForm";
import {
  vaultKeySchema,
  buildVaultKeyDefaults,
  buildVaultKeyPayload,
  type VaultKeyFormValues,
} from "./vaultKeySchema";

interface Props {
  open: boolean;
  editKey: VaultKeyEntry | null;
  onClose: () => void;
}

export default function KeyDrawer({ open, editKey, onClose }: Props) {
  const addKey = useVaultStore((s) => s.addKey);
  const updateKey = useVaultStore((s) => s.updateKey);
  const isEdit = !!editKey;

  const form = useDrawerForm(
    open,
    vaultKeySchema,
    buildVaultKeyDefaults(editKey),
  );
  const { control, setError, setValue, getValues } = form;

  const encrypted = useWatch({ control, name: "encrypted" });

  const {
    field: dataField,
    fieldState: { error: dataError },
  } = useController({ name: "data", control });

  const handleKeyChange = (pem: string) => {
    dataField.onChange(pem);
    const trimmed = pem.trim();
    const result = trimmed ? validatePrivateKey(trimmed) : null;
    setValue("encrypted", !!result?.valid && !!result.encrypted, {
      shouldValidate: true,
    });
    setValue("passphrase", "", { shouldValidate: true });
  };

  const handleFileName = (fileName: string) => {
    if (!getValues("name"))
      setValue("name", fileName, { shouldValidate: true });
  };

  const validateForPaste = (text: string) =>
    validatePrivateKey(text.trim()).valid;

  const onSubmit = async (values: VaultKeyFormValues) => {
    const data = values.data.trim();
    const passphrase = values.encrypted ? values.passphrase : undefined;

    let fingerprint: string;
    let algorithm: string;
    try {
      fingerprint = getFingerprint(data, passphrase);
      algorithm = getAlgorithm(data, passphrase);
    } catch (err) {
      if (values.encrypted) {
        const errName = err instanceof Error ? err.name : undefined;
        setError("passphrase", {
          message:
            errName === "KeyParseError"
              ? "Incorrect passphrase"
              : "Could not decrypt key with this passphrase",
        });
      } else {
        setError("data", { message: "Failed to read private key" });
      }
      return;
    }

    try {
      const payload = buildVaultKeyPayload(values, fingerprint, algorithm);
      if (isEdit && editKey) {
        await updateKey(editKey.id, payload);
      } else {
        await addKey(payload);
      }
      onClose();
    } catch (err: unknown) {
      if (err instanceof DuplicateKeyError) {
        if (err.field === "both" || err.field === "name") {
          setError("name", { message: "Name is already used" });
        }
        if (err.field === "both" || err.field === "private_key") {
          setError("data", { message: "Private key is already stored" });
        }
      } else {
        const msg = err instanceof Error ? err.message : "";
        setError("root", {
          message: msg || `Failed to ${isEdit ? "update" : "add"} key`,
        });
      }
    }
  };

  return (
    <FormDrawer
      form={form}
      onSubmit={onSubmit}
      open={open}
      onClose={onClose}
      title={isEdit ? "Edit Private Key" : "Add Private Key"}
      submitLabel={isEdit ? "Save Changes" : "Add Key"}
    >
      <FormInputField
        name="name"
        control={control}
        id="key-name"
        label="Name"
        maxLength={255}
        placeholder="e.g. Production Server"
      />

      <KeyFileInput
        label="Private Key"
        id="key-data"
        value={dataField.value}
        onChange={handleKeyChange}
        validate={validateForPaste}
        onFileName={handleFileName}
        disabled={isEdit}
        error={dataError?.message}
        placeholder={"-----BEGIN OPENSSH PRIVATE KEY-----\n..."}
        rows={8}
        hint="RSA, DSA, ECDSA, ED25519 — PEM and OpenSSH formats."
        loadedLabel={`Key loaded${encrypted ? " (encrypted)" : ""}`}
        emptyLabel="Drop private key file, paste, or browse"
      />

      {encrypted && (
        <FormPasswordField
          name="passphrase"
          control={control}
          id="key-passphrase"
          label="Passphrase"
          placeholder="Enter passphrase for encrypted key"
          autoComplete="off"
          hint="This key is encrypted. The passphrase is not stored."
        />
      )}
    </FormDrawer>
  );
}
