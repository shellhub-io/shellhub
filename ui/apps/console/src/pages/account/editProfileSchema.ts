import { z } from "zod";
import { validateEmail, validateName } from "@/utils/validation";
import { validateRecoveryEmail } from "./validate";
import { MFA_CODE_REQUIRED_MESSAGE } from "./mfaCode";

/**
 * The profile as it stands. The schema is built around it so a field can be validated only when
 * it has actually changed. mfaEnabled says whether a change that needs the current password also
 * takes an MFA code.
 */
export interface CurrentProfileValues {
  name: string;
  email: string;
  recoveryEmail: string;
  mfaEnabled?: boolean;
}

const editProfileFields = z.object({
  name: z.string(),
  email: z.string(),
  recoveryEmail: z.string(),
  currentPassword: z.string(),
  mfaCode: z.string(),
});

/**
 * The edit-profile form's values, derived from the schema.
 */
export type EditProfileFormValues = z.infer<typeof editProfileFields>;

/**
 * Whether values.email or values.recoveryEmail differs from its counterpart in current, compared
 * exactly, so a difference in case or padding alone counts as a change.
 */
export function changesAnAddress(
  values: Pick<EditProfileFormValues, "email" | "recoveryEmail">,
  current: CurrentProfileValues,
): boolean {
  return (
    values.email !== current.email ||
    values.recoveryEmail !== current.recoveryEmail
  );
}

type AddressErrors = Partial<Record<"email" | "recoveryEmail", string>>;

function addressErrors(
  values: Pick<EditProfileFormValues, "email" | "recoveryEmail">,
  current: CurrentProfileValues,
): AddressErrors {
  const errors: AddressErrors = {};

  if (values.email !== current.email) {
    const emailError = validateEmail(values.email);
    if (emailError) errors.email = emailError;
  }

  if (current.recoveryEmail && !values.recoveryEmail.trim()) {
    errors.recoveryEmail = "Recovery email can't be removed";
  } else if (values.recoveryEmail) {
    const recoveryEmailError = validateRecoveryEmail(
      values.recoveryEmail,
      values.email,
    );
    if (recoveryEmailError) errors.recoveryEmail = recoveryEmailError;
  }

  return errors;
}

/**
 * Whether saving values needs the current password, and with MFA a code: true when either email
 * differs from current and neither has an error (clearing a stored recovery email is one). While
 * an email has an error the form cannot be saved, so it asks for neither.
 */
export function needsCurrentPassword(
  values: Pick<EditProfileFormValues, "email" | "recoveryEmail">,
  current: CurrentProfileValues,
): boolean {
  const errors = addressErrors(values, current);
  return (
    changesAnAddress(values, current) && !errors.email && !errors.recoveryEmail
  );
}

/**
 * Builds the edit-profile schema around the current values, so an unchanged name or email is not
 * re-validated: a stored value that fails today's rules must not block saving the other fields.
 * The recovery email is checked whenever it is set, since it must differ from the email, and a
 * stored one cannot be cleared. A field counts as changed when its input differs from the stored
 * value at all, padding included. Changing either email to a valid address requires the current
 * password, and on an account with MFA enabled a code from the authenticator app or a recovery
 * code.
 */
export function editProfileSchema(current: CurrentProfileValues) {
  return editProfileFields.superRefine((values, ctx) => {
    if (values.name !== current.name) {
      const nameError = validateName(values.name);
      if (nameError)
        ctx.addIssue({ code: "custom", path: ["name"], message: nameError });
    }

    const errors = addressErrors(values, current);
    if (errors.email)
      ctx.addIssue({ code: "custom", path: ["email"], message: errors.email });
    if (errors.recoveryEmail) {
      ctx.addIssue({
        code: "custom",
        path: ["recoveryEmail"],
        message: errors.recoveryEmail,
      });
    }

    const passwordNeeded = needsCurrentPassword(values, current);
    if (passwordNeeded && !values.currentPassword) {
      ctx.addIssue({
        code: "custom",
        path: ["currentPassword"],
        message: "Enter your current password to change an email",
      });
    }

    if (
      current.mfaEnabled &&
      passwordNeeded &&
      !values.mfaCode.trim()
    ) {
      ctx.addIssue({
        code: "custom",
        path: ["mfaCode"],
        message: MFA_CODE_REQUIRED_MESSAGE,
      });
    }
  });
}
