import { ComponentProps, useState } from "react";
import { EyeIcon, EyeSlashIcon } from "@heroicons/react/24/outline";
import InputField from "@/components/common/fields/InputField";

type PasswordFieldProps = Omit<
  ComponentProps<typeof InputField>,
  "type" | "appendIcon"
> & {
  suppressPasswordManager?: boolean;
};

export default function PasswordField({
  suppressPasswordManager,
  autoComplete,
  ...rest
}: PasswordFieldProps) {
  const [visible, setVisible] = useState(false);
  const resolvedAutoComplete =
    autoComplete ?? (suppressPasswordManager ? "off" : "new-password");
  const suppressAttrs = suppressPasswordManager
    ? {
        "data-1p-ignore": "true",
        "data-lpignore": "true",
        "data-form-type": "other",
      }
    : undefined;

  return (
    <InputField
      {...rest}
      {...suppressAttrs}
      type={visible ? "text" : "password"}
      autoComplete={resolvedAutoComplete}
      appendIcon={
        <button
          type="button"
          onClick={() => setVisible((v) => !v)}
          aria-label={visible ? "Hide password" : "Show password"}
          className="text-text-muted hover:text-text-secondary transition-colors"
          tabIndex={-1}
        >
          {visible ? (
            <EyeSlashIcon className="w-4 h-4" />
          ) : (
            <EyeIcon className="w-4 h-4" />
          )}
        </button>
      }
    />
  );
}
