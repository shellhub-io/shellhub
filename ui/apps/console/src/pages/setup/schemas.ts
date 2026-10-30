import { z } from "zod";
import {
  validateEmail,
  validateIdentifier,
  validateName,
  validateNamespaceName,
  validatePassword,
  validateUsername,
} from "@/utils/validation";

type FieldCheck<T> = [field: keyof T & string, error: string | null];

function refineFields<T>(checks: (values: T) => FieldCheck<T>[]) {
  return (values: T, ctx: z.RefinementCtx) => {
    for (const [field, message] of checks(values)) {
      if (message) ctx.addIssue({ code: "custom", path: [field], message });
    }
  };
}

function passwordsMatch(confirmPassword: string, password: string): string | null {
  return confirmPassword === password ? null : "Passwords do not match";
}

const accountFields = z.object({
  name: z.string().trim(),
  username: z.string().trim(),
  password: z.string(),
  confirmPassword: z.string(),
});

function accountChecks(values: z.infer<typeof accountFields>): FieldCheck<typeof values>[] {
  return [
    ["name", validateName(values.name)],
    ["username", validateUsername(values.username)],
    ["password", validatePassword(values.password)],
    ["confirmPassword", passwordsMatch(values.confirmPassword, values.password)],
  ];
}

/**
 * The invitation form. There is no email: it comes from the invitation itself, and no terms or
 * marketing consent, which belong to cloud's open sign-up rather than to an invitation somebody
 * was sent.
 */
export const inviteSchema = accountFields.superRefine(refineFields(accountChecks));

/**
 * The invitation form's values, with name and username trimmed.
 */
export type InviteFormValues = z.infer<typeof inviteSchema>;

/**
 * The sign-up form, which also requires the privacy policy to be accepted.
 */
export const signUpSchema = accountFields
  .extend({ email: z.string().trim(), acceptPrivacyPolicy: z.boolean() })
  .superRefine(
    refineFields((values) => [
      ...accountChecks(values),
      ["email", validateEmail(values.email)],
      [
        "acceptPrivacyPolicy",
        values.acceptPrivacyPolicy ? null : "You must accept the privacy policy",
      ],
    ]),
  );

/**
 * The sign-up form's values, with name, username and email trimmed.
 */
export type SignUpFormValues = z.infer<typeof signUpSchema>;

/**
 * The first-run setup form: the admin account and the first namespace. The namespace is
 * pre-filled from the username but is still the user's to change, so it is checked too.
 */
export const setupSchema = accountFields
  .extend({ email: z.string().trim(), namespace: z.string() })
  .superRefine(
    refineFields((values) => [
      ...accountChecks(values),
      ["email", validateEmail(values.email)],
      ["namespace", validateNamespaceName(values.namespace)],
    ]),
  );

/**
 * The setup form's values, with name, username and email trimmed.
 */
export type SetupFormValues = z.infer<typeof setupSchema>;

/**
 * The sign-in form. The identifier is accepted when it is a valid username or a valid email, since
 * the one field takes either.
 */
export const loginSchema = z
  .object({
    username: z.string().trim(),
    password: z.string().min(1, "Password is required"),
  })
  .superRefine(refineFields((values) => [["username", validateIdentifier(values.username)]]));

/**
 * The sign-in form's values, with the identifier trimmed.
 */
export type LoginFormValues = z.infer<typeof loginSchema>;

/**
 * The forgot-password form: one field taking either a username or an email.
 */
export const forgotPasswordSchema = z
  .object({ account: z.string().trim() })
  .superRefine(refineFields((values) => [["account", validateIdentifier(values.account)]]));

/**
 * The forgot-password form's values, with the identifier trimmed.
 */
export type ForgotPasswordFormValues = z.infer<typeof forgotPasswordSchema>;

/**
 * The set-a-new-password form. Reports the length rule on `password` and "Passwords do not match"
 * on `confirmPassword`; neither field is trimmed.
 */
export const updatePasswordSchema = z
  .object({ password: z.string(), confirmPassword: z.string() })
  .superRefine(
    refineFields((values) => [
      ["password", validatePassword(values.password)],
      ["confirmPassword", passwordsMatch(values.confirmPassword, values.password)],
    ]),
  );

/**
 * The set-a-new-password form's values. The password is not trimmed: leading and trailing spaces
 * are part of a password.
 */
export type UpdatePasswordFormValues = z.infer<typeof updatePasswordSchema>;

/**
 * The MFA recovery form. The code is trimmed before it is checked: these are copied from a saved
 * file, and often arrive with whitespace attached.
 */
export const mfaRecoverSchema = z.object({
  recoveryCode: z.string().trim().min(1, "Recovery code is required"),
});

/**
 * The MFA recovery form's values, with the code trimmed.
 */
export type MfaRecoverFormValues = z.infer<typeof mfaRecoverSchema>;
