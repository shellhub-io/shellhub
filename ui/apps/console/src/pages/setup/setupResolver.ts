import type { FieldErrors, Resolver } from "react-hook-form";
import { validate, type FormErrors } from "./validate";

export interface SetupFormValues {
  name: string;
  username: string;
  email: string;
  password: string;
  confirmPassword: string;
}

type ValidateField = keyof FormErrors;

const VALIDATE_FIELDS: ValidateField[] = [
  "name",
  "username",
  "email",
  "password",
  "confirmPassword",
];

export const setupResolver: Resolver<SetupFormValues> = (values) => {
  const formErrors = validate(values);
  const errors: FieldErrors<SetupFormValues> = {};

  for (const field of VALIDATE_FIELDS) {
    const message = formErrors[field];

    if (message) {
      errors[field] = { type: "validate", message };
    }
  }

  if (Object.keys(errors).length > 0) {
    return { values: {}, errors };
  }

  return { values, errors: {} };
};
