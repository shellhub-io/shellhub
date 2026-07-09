import { useWatch, type Control } from "react-hook-form";
import {
  FormCheckboxField,
  FormNumericInput,
} from "@/components/common/fields/rhf";
import type { UserFormValues } from "./userSchema";

interface NamespaceLimitFieldsProps {
  control: Control<UserFormValues>;
  idPrefix: string;
}

export default function NamespaceLimitFields({
  control,
  idPrefix,
}: NamespaceLimitFieldsProps) {
  const limitEnabled = useWatch({ control, name: "limitEnabled" });
  const limitDisabled = useWatch({ control, name: "limitDisabled" });

  return (
    <div className="space-y-3">
      <FormCheckboxField
        name="limitEnabled"
        control={control}
        id={`${idPrefix}-limit-enabled`}
        label="Set namespace creation limit"
      />
      {limitEnabled && (
        <div className="ml-6 space-y-3 animate-fade-in">
          <FormCheckboxField
            name="limitDisabled"
            control={control}
            id={`${idPrefix}-limit-disabled`}
            label="Disable namespace creation"
          />
          {!limitDisabled && (
            <FormNumericInput
              name="maxNamespaces"
              control={control}
              id={`${idPrefix}-max-ns`}
              label="Max namespaces"
            />
          )}
        </div>
      )}
    </div>
  );
}
