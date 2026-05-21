import InputField from "@/components/common/fields/InputField";
import {
  NAMESPACE_NAME_HINT,
  NAMESPACE_NAME_MAX_LENGTH,
} from "@/utils/validation";

interface NamespaceNameFieldProps {
  id: string;
  value: string;
  onChange: (value: string) => void;
  error?: string | null;
  autoFocus?: boolean;
}

export default function NamespaceNameField({
  id,
  value,
  onChange,
  error,
  autoFocus,
}: NamespaceNameFieldProps) {
  return (
    <InputField
      id={id}
      label="Namespace Name"
      value={value}
      onChange={(v) => onChange(v.toLowerCase())}
      placeholder="my-namespace"
      maxLength={NAMESPACE_NAME_MAX_LENGTH}
      autoFocus={autoFocus}
      error={error ?? undefined}
      hint={NAMESPACE_NAME_HINT}
    />
  );
}
