import { useId } from "react";
import { MagnifyingGlassIcon } from "@heroicons/react/24/outline";
import { cn } from "@shellhub/design-system/cn";
import InputField from "@/components/common/fields/InputField";

interface SearchFieldProps {
  value: string;
  onChange: (next: string) => void;
  placeholder: string;
  /** Required accessible name. Rendered as a visually hidden <label>. */
  "aria-label": string;
  /** Layout overrides only (margins, alignment). Don't override colors or sizes. */
  className?: string;
  id?: string;
  /** Remove the max-width cap. Defaults to capping at `max-w-sm`. */
  full?: boolean;
}

export default function SearchField({
  value,
  onChange,
  placeholder,
  "aria-label": ariaLabel,
  className,
  id,
  full = false,
}: SearchFieldProps) {
  const generatedId = useId();
  const inputId = id ?? generatedId;

  const wrapperClasses = cn(!full && "max-w-sm w-full", className);

  return (
    <div className={wrapperClasses || undefined}>
      <InputField
        id={inputId}
        label={ariaLabel}
        hideLabel
        variant="mono"
        type="search"
        value={value}
        onChange={onChange}
        placeholder={placeholder}
        prependIcon={
          <MagnifyingGlassIcon
            aria-hidden="true"
            className="w-4 h-4 text-text-muted"
            strokeWidth={2}
          />
        }
      />
    </div>
  );
}
