import { type Control } from "react-hook-form";
import {
  FormInputField,
  FormPasswordField,
  FormCheckboxField,
} from "@/components/common/fields/rhf";
import NamespaceLimitFields from "./NamespaceLimitFields";
import {
  NAME_MAX_LENGTH,
  PASSWORD_HINT,
  PASSWORD_MAX_LENGTH,
  USERNAME_HINT,
  USERNAME_MAX_LENGTH,
} from "@/utils/validation";
import type { UserFormMode, UserFormValues } from "./userSchema";

interface UserFormFieldsProps {
  control: Control<UserFormValues>;
  mode: UserFormMode;
  idPrefix: string;
  /** Edit only: false when the user is already confirmed (cannot un-confirm). */
  canChangeConfirmed?: boolean;
  /** Edit only: true when editing your own admin user (cannot self-demote). */
  disableAdmin?: boolean;
}

export default function UserFormFields({
  control,
  mode,
  idPrefix,
  canChangeConfirmed = true,
  disableAdmin = false,
}: UserFormFieldsProps) {
  const isCreate = mode === "create";

  return (
    <>
      <FormInputField
        name="name"
        control={control}
        id={`${idPrefix}-name`}
        label="Name"
        errorRole="status"
        placeholder={isCreate ? "John Doe" : undefined}
        maxLength={NAME_MAX_LENGTH}
        required
      />

      <FormInputField
        name="username"
        control={control}
        id={`${idPrefix}-username`}
        label="Username"
        errorRole="status"
        placeholder={isCreate ? "johndoe" : undefined}
        hint={USERNAME_HINT}
        autoComplete="username"
        maxLength={USERNAME_MAX_LENGTH}
        required
      />

      <FormInputField
        name="email"
        control={control}
        id={`${idPrefix}-email`}
        label="Email"
        type="email"
        errorRole="status"
        placeholder={isCreate ? "john@example.com" : undefined}
        autoComplete="email"
        required
      />

      <FormPasswordField
        name="password"
        control={control}
        id={`${idPrefix}-password`}
        label="Password"
        errorRole="status"
        placeholder={
          isCreate ? "Enter password" : "Leave blank to keep current"
        }
        hint={isCreate ? PASSWORD_HINT : undefined}
        maxLength={PASSWORD_MAX_LENGTH}
        suppressPasswordManager
        required={isCreate}
      />

      <NamespaceLimitFields control={control} idPrefix={idPrefix} />

      {mode === "edit" && (
        <FormCheckboxField
          name="confirmed"
          control={control}
          id={`${idPrefix}-confirmed`}
          label="Confirmed"
          disabled={!canChangeConfirmed}
          title={
            !canChangeConfirmed
              ? "Cannot remove confirmation from a confirmed user"
              : undefined
          }
        />
      )}

      <FormCheckboxField
        name="admin"
        control={control}
        id={`${idPrefix}-admin`}
        label="Admin user"
        disabled={disableAdmin}
        title={
          disableAdmin ? "Cannot remove your own admin privilege" : undefined
        }
      />
    </>
  );
}
