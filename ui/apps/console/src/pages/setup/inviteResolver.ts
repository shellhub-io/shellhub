import type { FieldErrors, Resolver } from "react-hook-form";
import { validate } from "./validate";

// The invitee completing an invitation only sets their profile + password; the
// email is auto-derived from the invite code, and there's no ToS/marketing (that
// belongs to Cloud's open sign-up). So this is a trimmed sign-up resolver.
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
