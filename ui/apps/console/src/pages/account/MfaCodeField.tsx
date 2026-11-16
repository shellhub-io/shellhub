import type { Control, FieldValues, Path } from "react-hook-form";
import { FormInputField } from "@/components/common/fields/rhf";
import { MFA_CODE_HINT } from "@/pages/account/mfaCode";

/**
 * The field for an authenticator or recovery code, bound to the mfaCode value of a form that asks
 * an account with MFA for one.
 */
export default function MfaCodeField<T extends FieldValues & { mfaCode: string }>({
  control,
  id,
  onValueChange,
}: {
  control: Control<T>;
  id: string;
  onValueChange?: () => void;
}) {
  return (
    <FormInputField
      name={"mfaCode" as Path<T>}
      control={control}
      id={id}
      label="Authenticator or Recovery Code"
      autoComplete="one-time-code"
      placeholder="123456"
      hint={MFA_CODE_HINT}
      onValueChange={onValueChange}
    />
  );
}
