import { useController, type Control, type FieldValues, type Path } from "react-hook-form";
import { isPublicKeyValid } from "@/utils/sshKeys";
import KeyFileInput from "@/components/common/fields/KeyFileInput";

type Props<T extends FieldValues> = {
  name: Path<T>;
  control: Control<T>;
  disabled?: boolean;
  onFileName?: (name: string) => void;
};

function KeyDataInput<T extends FieldValues>({
  name,
  control,
  disabled,
  onFileName,
}: Props<T>) {
  const {
    field,
    fieldState: { error },
  } = useController({ name, control });

  return (
    <KeyFileInput
      id="public-key-data"
      label="Public key data"
      value={field.value}
      onChange={field.onChange}
      validate={isPublicKeyValid}
      onFileName={onFileName}
      disabled={disabled}
      error={error?.message}
      accept=".pub,.pem,.key,.txt"
      placeholder="ssh-rsa AAAAB3NzaC1yc2E..."
      rows={3}
      hint="RSA, DSA, ECDSA, ED25519 — PEM and OpenSSH formats."
      disabledHint="Public key data cannot be modified after creation."
    />
  );
}

export default KeyDataInput;
