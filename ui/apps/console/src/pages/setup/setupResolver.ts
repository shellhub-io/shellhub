import type { FieldErrors, Resolver } from "react-hook-form";
import { validateNamespaceName } from "@/utils/validation";
import { validate, type FormErrors } from "./validate";

export interface SetupFormValues {
  name: string;
  username: string;
  namespace: string;
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

  const namespaceError = validateNamespaceName(values.namespace);
  if (namespaceError) {
    errors.namespace = { type: "validate", message: namespaceError };
  }

  if (Object.keys(errors).length > 0) {
    return { values: {}, errors };
  }

  return { values, errors: {} };
};
