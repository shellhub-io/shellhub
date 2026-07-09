import { useController, type Control, type FieldValues, type Path } from "react-hook-form";
import type { ComponentProps } from "react";
import TagsSelector from "@/components/common/fields/TagsSelector";

type TagsSelectorProps = Omit<
  ComponentProps<typeof TagsSelector>,
  "selected" | "onChange"
>;

type Props<T extends FieldValues> = TagsSelectorProps & {
  name: Path<T>;
  control: Control<T>;
};

export default function FormTagsSelector<T extends FieldValues>({
  name,
  control,
  error: errorOverride,
  ...rest
}: Props<T>) {
  const {
    field,
    fieldState: { error: fieldError },
  } = useController({ name, control });

  const resolvedError = errorOverride ?? fieldError?.message;

  return (
    <TagsSelector
      {...rest}
      selected={field.value ?? []}
      onChange={(tags) => {
        field.onChange(tags);
      }}
      error={resolvedError}
    />
  );
}
