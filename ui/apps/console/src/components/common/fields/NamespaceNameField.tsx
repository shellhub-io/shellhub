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
}

/**
 * The namespace name field, with the rules shown as a checklist that resolves as the user types
 * — the constraints are unusual enough that discovering them by error is a poor experience.
 */
export default function NamespaceNameField({
  id,
  value,
  onChange,
  error,
}: NamespaceNameFieldProps) {
  return (
    <InputField
      id={id}
      label="Namespace Name"
      value={value}
      onChange={(v) => onChange(v.toLowerCase())}
      placeholder="my-namespace"
      maxLength={NAMESPACE_NAME_MAX_LENGTH}

      error={error ?? undefined}
      hint={NAMESPACE_NAME_HINT}
    />
  );
}
