import type { FieldErrors as RhfFieldErrors, Resolver } from "react-hook-form";
import { validatePassword } from "@/utils/validation";

export interface ChangePasswordFormValues {
  current: string;
  newPw: string;
  confirmPw: string;
}

type PlainErrors = Partial<Record<keyof ChangePasswordFormValues, string>>;

export function changePasswordResolver(values: ChangePasswordFormValues): PlainErrors {
  const errors: PlainErrors = {};

  const newPwError = values.newPw ? validatePassword(values.newPw) : null;
  if (newPwError) errors.newPw = newPwError;

  if (values.confirmPw && values.confirmPw !== values.newPw) {
    errors.confirmPw = "Passwords do not match";
  }

  return errors;
}

export const rhfChangePasswordResolver: Resolver<ChangePasswordFormValues> = (values) => {
  const plain = changePasswordResolver(values);
  const errors: RhfFieldErrors<ChangePasswordFormValues> = {};

  for (const [key, message] of Object.entries(plain) as [keyof ChangePasswordFormValues, string][]) {
    errors[key] = { type: "validate", message };
  }

  if (Object.keys(errors).length > 0) {
    return { values: {}, errors };
  }

  return { values, errors: {} };
};
