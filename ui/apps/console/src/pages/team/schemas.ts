import { z } from "zod";
import { EMAIL_REGEX } from "@/utils/validation";
import type { ApiKey, ApiKeyCreate } from "@/client";
import type { NamespaceMember } from "@/hooks/useNamespaces";
import { ROLES, isAssignableRole, type AssignableRole } from "./helpers";

const roleField = z.enum(ROLES);

const emailField = z.string().refine((v) => EMAIL_REGEX.test(v.trim()), {
  message: "Enter a valid email address.",
});

export const addMemberSchema = z.object({
  email: emailField,
  role: roleField,
});

export type AddMemberFormValues = z.infer<typeof addMemberSchema>;

export const ADD_MEMBER_DEFAULTS: AddMemberFormValues = {
  email: "",
  role: "operator",
};

export function buildAddMemberBody(values: AddMemberFormValues) {
  return { email: values.email.trim(), role: values.role };
}

export const editRoleSchema = z.object({ role: roleField });

export type EditRoleFormValues = z.infer<typeof editRoleSchema>;

function assignableRoleOr(
  role: unknown,
  fallback: AssignableRole,
): AssignableRole {
  return isAssignableRole(role) ? role : fallback;
}

export function buildMemberRoleDefaults(
  member: NamespaceMember | null,
): EditRoleFormValues {
  return { role: assignableRoleOr(member?.role, "operator") };
}

export const generateKeySchema = z.object({
  name: z.string().superRefine((value, ctx) => {
    if (value.length < 3) {
      ctx.addIssue({
        code: "custom",
        message: "Name must be at least 3 characters.",
      });
    } else if (value.length > 20) {
      ctx.addIssue({
        code: "custom",
        message: "Name must be at most 20 characters.",
      });
    } else if (!/^[a-zA-Z0-9_-]+$/.test(value)) {
      ctx.addIssue({
        code: "custom",
        message: "Name can only contain letters, numbers, - and _.",
      });
    }
  }),
  role: roleField,
  // Kept as a string so it binds directly to the radio-pill group; converted
  // back to the numeric API value in buildGenerateKeyBody.
  expiresIn: z.string(),
});

export type GenerateKeyFormValues = z.infer<typeof generateKeySchema>;

export const GENERATE_KEY_DEFAULTS: GenerateKeyFormValues = {
  name: "",
  role: "administrator",
  expiresIn: "30",
};

export function buildGenerateKeyBody(
  values: GenerateKeyFormValues,
): ApiKeyCreate {
  return {
    name: values.name.trim(),
    role: values.role,
    expires_at: Number(values.expiresIn) as ApiKeyCreate["expires_at"],
  };
}

export const editKeySchema = z.object({
  name: z
    .string()
    .refine((v) => v.trim().length > 0, { message: "Name is required." }),
  role: roleField,
});

export type EditKeyFormValues = z.infer<typeof editKeySchema>;

export function buildEditKeyDefaults(apiKey: ApiKey | null): EditKeyFormValues {
  return {
    name: apiKey?.name ?? "",
    role: assignableRoleOr(apiKey?.role, "administrator"),
  };
}

export function buildEditKeyBody(values: EditKeyFormValues) {
  return { name: values.name.trim(), role: values.role };
}
