import { z } from "zod";
import {
  validateName,
  validateUsername,
  validateEmail,
  validateRecoveryEmail,
} from "./validate";

/**
 * The profile as it stands. The schema is built around it so a field can be validated only when
 * it has actually changed.
 */
export interface CurrentProfileValues {
  name: string;
  username: string;
  email: string;
}

const editProfileFields = z.object({
  name: z.string(),
  username: z.string(),
  email: z.string(),
  recoveryEmail: z.string(),
});

/**
 * The edit-profile form's values, derived from the schema.
 */
export type EditProfileFormValues = z.infer<typeof editProfileFields>;

/**
 * Builds the edit-profile schema around the current values, so an unchanged field is not
 * re-validated — a username that is already taken by this very account must not be rejected.
 */
export function editProfileSchema(current: CurrentProfileValues) {
  return editProfileFields.superRefine((values, ctx) => {
    if (values.name !== current.name) {
      const nameError = validateName(values.name);
      if (nameError)
        ctx.addIssue({ code: "custom", path: ["name"], message: nameError });
    }

    if (values.username !== current.username) {
      const usernameError = validateUsername(values.username);
      if (usernameError)
        ctx.addIssue({
          code: "custom",
          path: ["username"],
          message: usernameError,
        });
    }

    if (values.email !== current.email) {
      const emailError = validateEmail(values.email);
      if (emailError)
        ctx.addIssue({ code: "custom", path: ["email"], message: emailError });
    }

    if (values.recoveryEmail) {
      const recoveryEmailError = validateRecoveryEmail(
        values.recoveryEmail,
        values.email,
      );
      if (recoveryEmailError) {
        ctx.addIssue({
          code: "custom",
          path: ["recoveryEmail"],
          message: recoveryEmailError,
        });
      }
    }
  });
}
