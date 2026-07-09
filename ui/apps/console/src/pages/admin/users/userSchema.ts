import { z } from "zod";
import type {
  UserAdminCreateRequest,
  UserAdminResponse,
  UserAdminUpdateRequest,
} from "@/client";
import {
  MAX_NAMESPACES_ERROR,
  isMaxNamespacesValid,
  validateEmail,
  validateName,
  validatePassword,
  validateUsername,
} from "@/utils/validation";

export type UserFormMode = "create" | "edit";

export type UserFormPayload<M extends UserFormMode> = M extends "create"
  ? UserAdminCreateRequest
  : UserAdminUpdateRequest;

const userFields = z.object({
  name: z.string(),
  username: z.string(),
  email: z.string(),
  password: z.string(),
  confirmed: z.boolean(),
  admin: z.boolean(),
  limitEnabled: z.boolean(),
  limitDisabled: z.boolean(),
  maxNamespaces: z.string(),
});

export type UserFormValues = z.infer<typeof userFields>;

/**
 * On edit a blank password means "keep the current one", so it skips the
 * length rule; every other field is validated the same way in both modes.
 */
export function userSchema(mode: UserFormMode) {
  return userFields.superRefine((v, ctx) => {
    const nameError = validateName(v.name);
    if (nameError) ctx.addIssue({ code: "custom", path: ["name"], message: nameError });

    const usernameError = validateUsername(v.username);
    if (usernameError) ctx.addIssue({ code: "custom", path: ["username"], message: usernameError });

    const emailError = validateEmail(v.email);
    if (emailError) ctx.addIssue({ code: "custom", path: ["email"], message: emailError });

    if (!(mode === "edit" && v.password === "")) {
      const passwordError = validatePassword(v.password);
      if (passwordError) ctx.addIssue({ code: "custom", path: ["password"], message: passwordError });
    }

    if (!isMaxNamespacesValid(v.limitEnabled, v.limitDisabled, v.maxNamespaces)) {
      ctx.addIssue({ code: "custom", path: ["maxNamespaces"], message: MAX_NAMESPACES_ERROR });
    }
  });
}

function blankValues(): UserFormValues {
  return {
    name: "",
    username: "",
    email: "",
    password: "",
    confirmed: false,
    admin: false,
    limitEnabled: false,
    limitDisabled: false,
    maxNamespaces: "1",
  };
}

function valuesFromUser(user: UserAdminResponse): UserFormValues {
  const max = user.max_namespaces;
  const limitEnabled = max !== undefined && max >= 0;
  const limitDisabled = limitEnabled && max === 0;
  return {
    name: user.name ?? "",
    username: user.username ?? "",
    email: user.email ?? "",
    password: "",
    confirmed: user.status === "confirmed",
    admin: user.admin ?? false,
    limitEnabled,
    limitDisabled,
    maxNamespaces: limitEnabled && max ? String(max) : "1",
  };
}

export function buildUserDefaults(user?: UserAdminResponse | null): UserFormValues {
  return user ? valuesFromUser(user) : blankValues();
}

function computeMaxNamespaces(
  mode: UserFormMode,
  values: UserFormValues,
  initial?: UserAdminResponse | null,
): number | undefined {
  if (!values.limitEnabled) {
    if (mode === "edit") {
      const orig = initial?.max_namespaces;
      return orig !== undefined && orig < 0 ? orig : undefined;
    }
    return undefined;
  }
  if (values.limitDisabled) return 0;
  return parseInt(values.maxNamespaces, 10);
}

export function buildUserPayload<M extends UserFormMode>(
  mode: M,
  values: UserFormValues,
  initial?: UserAdminResponse | null,
): UserFormPayload<M> {
  const base = {
    name: values.name.trim(),
    username: values.username.trim(),
    email: values.email.trim(),
    admin: values.admin,
    max_namespaces: computeMaxNamespaces(mode, values, initial),
  };

  if (mode === "create") {
    const payload: UserAdminCreateRequest = { ...base, password: values.password };
    return payload;
  }

  const payload: UserAdminUpdateRequest = {
    ...base,
    confirmed: values.confirmed,
    // Omit when blank so the backend keeps the current password.
    ...(values.password !== "" && { password: values.password }),
  };
  return payload as UserFormPayload<M>;
}
