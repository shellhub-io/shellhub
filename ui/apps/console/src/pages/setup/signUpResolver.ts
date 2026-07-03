import type { FieldErrors, Resolver } from "react-hook-form";
import { validate, type FormErrors } from "./validate";

export interface SignUpFormValues {
  name: string;
  username: string;
  email: string;
  password: string;
  confirmPassword: string;
  acceptPrivacyPolicy: boolean;
}

type ValidateField = keyof FormErrors;

const VALIDATE_FIELDS: ValidateField[] = [
  "name",
  "username",
  "email",
  "password",
  "confirmPassword",
];

export const signUpResolver: Resolver<SignUpFormValues> = (values) => {
  const formErrors = validate(values);
  const errors: FieldErrors<SignUpFormValues> = {};

  for (const field of VALIDATE_FIELDS) {
    const message = formErrors[field];

    if (message) {
      errors[field] = { type: "validate", message };
    }
  }

  if (!values.acceptPrivacyPolicy) {
    errors.acceptPrivacyPolicy = {
      type: "required",
      message: "You must accept the privacy policy",
    };
  }

  if (Object.keys(errors).length > 0) {
    return { values: {}, errors };
  }

  return { values, errors: {} };
};
