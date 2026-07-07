import type { FieldErrors, Resolver } from "react-hook-form";
import {
  validateName,
  validateUsername,
  validateEmail,
  validateRecoveryEmail,
} from "./validate";

export interface EditProfileFormValues {
  name: string;
  username: string;
  email: string;
  recoveryEmail: string;
}

type PlainErrors = Partial<Record<keyof EditProfileFormValues, string>>;

export function editProfileResolver(values: EditProfileFormValues): PlainErrors {
  const errors: PlainErrors = {};

  const nameError = validateName(values.name);
  if (nameError) errors.name = nameError;

  const usernameError = validateUsername(values.username);
  if (usernameError) errors.username = usernameError;

  const emailError = validateEmail(values.email);
  if (emailError) errors.email = emailError;

  const recoveryEmailError = values.recoveryEmail
    ? validateRecoveryEmail(values.recoveryEmail, values.email)
    : null;
  if (recoveryEmailError) errors.recoveryEmail = recoveryEmailError;

  return errors;
}

export const rhfEditProfileResolver: Resolver<EditProfileFormValues> = (values) => {
  const plain = editProfileResolver(values);
  const errors: FieldErrors<EditProfileFormValues> = {};

  for (const [key, message] of Object.entries(plain) as [keyof EditProfileFormValues, string][]) {
    errors[key] = { type: "validate", message };
  }

  if (Object.keys(errors).length > 0) {
    return { values: {}, errors };
  }

  return { values, errors: {} };
};
