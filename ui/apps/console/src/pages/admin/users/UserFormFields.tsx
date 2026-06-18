import InputField from "@/components/common/fields/InputField";
import PasswordField from "@/components/common/fields/PasswordField";
import CheckboxField from "@/components/common/fields/CheckboxField";
import NamespaceLimitFields from "./NamespaceLimitFields";
import {
  NAME_MAX_LENGTH,
  PASSWORD_HINT,
  PASSWORD_MAX_LENGTH,
  USERNAME_HINT,
  USERNAME_MAX_LENGTH,
} from "@/utils/validation";
import type { UserFormApi, UserFormMode } from "./useUserForm";

interface UserFormFieldsProps {
  form: UserFormApi<UserFormMode>;
  idPrefix: string;
  /** Edit only: false when the user is already confirmed (cannot un-confirm). */
  canChangeConfirmed?: boolean;
  /** Edit only: true when editing your own admin user (cannot self-demote). */
  disableAdmin?: boolean;
}

export default function UserFormFields({
  form,
  idPrefix,
  canChangeConfirmed = true,
  disableAdmin = false,
}: UserFormFieldsProps) {
  const { mode, values, errors, setField, validateField } = form;
  const isCreate = mode === "create";

  return (
    <>
      <InputField
        id={`${idPrefix}-name`}
        label="Name"
        value={values.name}
        onChange={(v) => setField("name", v)}
        onBlur={() => validateField("name")}
        error={errors.name}
        errorRole="status"
        placeholder={isCreate ? "John Doe" : undefined}

        maxLength={NAME_MAX_LENGTH}
        required
      />

      <InputField
        id={`${idPrefix}-username`}
        label="Username"
        value={values.username}
        onChange={(v) => setField("username", v)}
        onBlur={() => validateField("username")}
        error={errors.username}
        errorRole="status"
        placeholder={isCreate ? "johndoe" : undefined}
        hint={USERNAME_HINT}
        autoComplete="username"
        maxLength={USERNAME_MAX_LENGTH}
        required
      />

      <InputField
        id={`${idPrefix}-email`}
        label="Email"
        type="email"
        value={values.email}
        onChange={(v) => setField("email", v)}
        onBlur={() => validateField("email")}
        error={errors.email}
        errorRole="status"
        placeholder={isCreate ? "john@example.com" : undefined}
        autoComplete="email"
        required
      />

      <PasswordField
        id={`${idPrefix}-password`}
        label="Password"
        value={values.password}
        onChange={(v) => setField("password", v)}
        onBlur={() => validateField("password")}
        error={errors.password}
        errorRole="status"
        placeholder={
          isCreate ? "Enter password" : "Leave blank to keep current"
        }
        hint={isCreate ? PASSWORD_HINT : undefined}
        maxLength={PASSWORD_MAX_LENGTH}
        suppressPasswordManager
        required={isCreate}
      />

      <NamespaceLimitFields
        idPrefix={idPrefix}
        limitEnabled={values.limitEnabled}
        onLimitEnabledChange={(v) => setField("limitEnabled", v)}
        limitDisabled={values.limitDisabled}
        onLimitDisabledChange={(v) => setField("limitDisabled", v)}
        maxNamespaces={values.maxNamespaces}
        onMaxNamespacesChange={(v) => setField("maxNamespaces", v)}
      />

      {mode === "edit" && (
        <CheckboxField
          id={`${idPrefix}-confirmed`}
          label="Confirmed"
          checked={values.confirmed}
          onChange={(v) => setField("confirmed", v)}
          disabled={!canChangeConfirmed}
          title={
            !canChangeConfirmed
              ? "Cannot remove confirmation from a confirmed user"
              : undefined
          }
        />
      )}

      <CheckboxField
        id={`${idPrefix}-admin`}
        label="Admin user"
        checked={values.admin}
        onChange={(v) => setField("admin", v)}
        disabled={disableAdmin}
        title={
          disableAdmin ? "Cannot remove your own admin privilege" : undefined
        }
      />
    </>
  );
}
