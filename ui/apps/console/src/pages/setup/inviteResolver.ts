import type { FieldErrors, Resolver } from "react-hook-form";
import { validate } from "./validate";

/**
 * The fields an invitee fills in. There is no email: it comes from the invitation itself, and no
 * terms or marketing consent, which belong to cloud's open sign-up rather than to an invitation
 * somebody was sent.
 */
export interface InviteFormValues {
  name: string;
  username: string;
  password: string;
  confirmPassword: string;
}

const VALIDATE_FIELDS = [
  "name",
  "username",
  "password",
  "confirmPassword",
] as const;

/**
 * Validates the invitation form for react-hook-form. It reuses the sign-up rules with a
 * placeholder email, so the shared checks cannot drift from sign-up's.
 */
export const inviteResolver: Resolver<InviteFormValues> = (values) => {
  const formErrors = validate({ ...values, email: "invite@placeholder.local" });
  const errors: FieldErrors<InviteFormValues> = {};

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
