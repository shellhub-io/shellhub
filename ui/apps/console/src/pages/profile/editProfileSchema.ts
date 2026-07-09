import { z } from "zod";
import {
  validateName,
  validateUsername,
  validateEmail,
  validateRecoveryEmail,
} from "./validate";

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

export type EditProfileFormValues = z.infer<typeof editProfileFields>;

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
