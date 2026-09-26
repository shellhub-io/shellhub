import { z } from "zod";
import { validateName, validateEmail, validateRecoveryEmail } from "./validate";

/**
 * The profile as it stands. The schema is built around it so a field can be validated only when
 * it has actually changed.
 */
export interface CurrentProfileValues {
  name: string;
  email: string;
}

const editProfileFields = z.object({
  name: z.string(),
  email: z.string(),
  recoveryEmail: z.string(),
});

/**
 * The edit-profile form's values, derived from the schema.
 */
export type EditProfileFormValues = z.infer<typeof editProfileFields>;

/**
 * Builds the edit-profile schema around the current values, so an unchanged field is not
 * re-validated: a stored value that fails today's rules must not block saving the other fields.
 */
export function editProfileSchema(current: CurrentProfileValues) {
  return editProfileFields.superRefine((values, ctx) => {
    if (values.name !== current.name) {
      const nameError = validateName(values.name);
      if (nameError)
        ctx.addIssue({ code: "custom", path: ["name"], message: nameError });
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
