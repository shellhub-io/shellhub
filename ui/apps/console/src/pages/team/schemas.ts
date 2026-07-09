import { z } from "zod";
import { EMAIL_REGEX } from "@/utils/validation";
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
