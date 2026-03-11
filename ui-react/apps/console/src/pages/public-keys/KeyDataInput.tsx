import { isPublicKeyValid } from "../../utils/sshKeys";
import KeyFileInput from "../../components/common/KeyFileInput";

function KeyDataInput({
  value,
  onChange,
  error,
  disabled,
  onFileName,
}: {
  value: string;
  onChange: (v: string) => void;
  error?: string;
  disabled?: boolean;
  onFileName?: (name: string) => void;
}) {
  return (
    <KeyFileInput
      label="Public key data"
      value={value}
      onChange={onChange}
      validate={isPublicKeyValid}
      onFileName={onFileName}
      disabled={disabled}
      error={error}
      accept=".pub,.pem,.key,.txt"
      placeholder="ssh-rsa AAAAB3NzaC1yc2E..."
      rows={3}
      hint="RSA, DSA, ECDSA, ED25519 — PEM and OpenSSH formats."
      disabledHint="Public key data cannot be modified after creation."
    />
  );
}

export default KeyDataInput;
