import {
  useController,
  type Control,
  type FieldValues,
  type Path,
} from "react-hook-form";
import { Toggle } from "@shellhub/design-system/primitives";
import { cn } from "@shellhub/design-system/cn";
import { LABEL } from "@/utils/styles";
import FieldError from "@/components/common/fields/FieldError";

type Props<T extends FieldValues> = {
  name: Path<T>;
  control: Control<T>;
  label: string;
  id?: string;
  activeLabel?: string;
  inactiveLabel?: string;
  error?: string;
};

export default function FormToggleField<T extends FieldValues>({
  name,
  control,
  label,
  id,
  activeLabel = "Active",
  inactiveLabel = "Inactive",
  error: errorOverride,
}: Props<T>) {
  const {
    field,
    fieldState: { error: fieldError },
  } = useController({ name, control });

  const checked = Boolean(field.value);
  const resolvedError = errorOverride ?? fieldError?.message;
  const baseId = id ?? String(name);
  const labelId = `${baseId}-label`;
  const errorId = `${baseId}-error`;

  return (
    <div>
      <span id={labelId} className={LABEL}>
        {label}
      </span>

      <div
        className={cn(
          "flex items-center gap-2.5 text-sm",
          checked ? "text-primary" : "text-text-muted",
        )}
      >
        <Toggle
          id={id}
          enabled={checked}
          onChange={(v) => field.onChange(v)}
          onBlur={field.onBlur}
          aria-labelledby={labelId}
          aria-invalid={resolvedError ? true : undefined}
          aria-describedby={resolvedError ? errorId : undefined}
        />
        {checked ? activeLabel : inactiveLabel}
      </div>

      <FieldError id={errorId}>{resolvedError}</FieldError>
    </div>
  );
}
