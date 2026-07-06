import type { FieldErrors, Resolver } from "react-hook-form";
import { validatePassword } from "@/utils/validation";

export interface UpdatePasswordFormValues {
  password: string;
  confirmPassword: string;
}

export const updatePasswordResolver: Resolver<UpdatePasswordFormValues> = (values) => {
  const errors: FieldErrors<UpdatePasswordFormValues> = {};

  const passwordError = validatePassword(values.password);
  if (passwordError) {
    errors.password = { type: "validate", message: passwordError };
  }

  if (values.confirmPassword !== values.password) {
    errors.confirmPassword = { type: "validate", message: "Passwords do not match" };
  }

  if (Object.keys(errors).length > 0) {
    return { values: {}, errors };
  }

  return { values: { ...values }, errors: {} };
};
