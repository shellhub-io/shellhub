import { z } from "zod";
import { EMAIL_REGEX, validateKeyName } from "@/utils/validation";
import type { ApiKey, ApiKeyCreate } from "@/client";
import { ROLES, assignableRoleOr } from "./helpers";

const roleField = z.enum(ROLES);

const emailField = z.string().refine((v) => EMAIL_REGEX.test(v.trim()), {
  message: "Enter a valid email address.",
});

/**
 * Validates the add-member form.
 */
export const addMemberSchema = z.object({
  email: emailField,
  role: roleField,
});

/**
 * The add-member form's values, derived from the schema.
 */
export type AddMemberFormValues = z.infer<typeof addMemberSchema>;

/**
 * What the add-member form starts as. Operator rather than administrator: the default should be
 * the lesser privilege.
 */
export const ADD_MEMBER_DEFAULTS: AddMemberFormValues = {
  email: "",
  role: "operator",
};

/**
 * Builds the add-member request body, trimming the address so a pasted one with trailing space
 * still resolves.
 */
export function buildAddMemberBody(values: AddMemberFormValues) {
  return { email: values.email.trim(), role: values.role };
}

const keyNameField = z.string().superRefine((value, ctx) => {
  const nameError = validateKeyName(value);
  if (nameError) ctx.addIssue({ code: "custom", message: nameError });
});

/**
 * Validates the generate-key form. The name rule is shared with the edit form through
 * keyNameField, so a name that is acceptable when a key is created stays acceptable when it is
 * renamed.
 */
export const generateKeySchema = z.object({
  name: keyNameField,
  role: roleField,
  expiresIn: z.string(),
});

/**
 * The generate-key form's values, derived from the schema.
 */
export type GenerateKeyFormValues = z.infer<typeof generateKeySchema>;

/**
 * What the generate-key form starts as: an administrator key expiring in thirty days.
 */
export const GENERATE_KEY_DEFAULTS: GenerateKeyFormValues = {
  name: "",
  role: "administrator",
  expiresIn: "30",
};

/**
 * Builds the create request for an API key.
 */
export function buildGenerateKeyBody(
  values: GenerateKeyFormValues,
): ApiKeyCreate {
  return {
    name: values.name.trim(),
    role: values.role,
    expires_at: Number(values.expiresIn) as ApiKeyCreate["expires_at"],
  };
}

/**
 * Validates the edit-key form. The expiry is absent: it is fixed when the key is made.
 */
export const editKeySchema = z.object({
  name: keyNameField,
  role: roleField,
});

/**
 * The edit-key form's values, derived from the schema.
 */
export type EditKeyFormValues = z.infer<typeof editKeySchema>;

/**
 * Fills the edit-key form from an existing key.
 */
export function buildEditKeyDefaults(apiKey: ApiKey | null): EditKeyFormValues {
  return {
    name: apiKey?.name ?? "",
    role: assignableRoleOr(apiKey?.role, "administrator"),
  };
}

/**
 * Builds the update request for an API key.
 */
export function buildEditKeyBody(values: EditKeyFormValues) {
  return { name: values.name.trim(), role: values.role };
}
