import { useState } from "react";
import type { UserAdminRequest, UserAdminResponse } from "@/client";
import {
  MAX_NAMESPACES_ERROR,
  isMaxNamespacesValid,
  validateEmail,
  validateName,
  validatePassword,
  validateUsername,
} from "@/utils/validation";

export type UserFormMode = "create" | "edit";

export interface UserFormValues {
  name: string;
  username: string;
  email: string;
  password: string;
  confirmed: boolean;
  admin: boolean;
  limitEnabled: boolean;
  limitDisabled: boolean;
  maxNamespaces: string;
}

export type UserFormErrorKey =
  | "name"
  | "username"
  | "email"
  | "password"
  | "maxNamespaces";

export type UserFormErrors = Partial<Record<UserFormErrorKey, string>>;

export interface UseUserFormOptions {
  mode: UserFormMode;
}

export interface UserFormApi {
  mode: UserFormMode;
  values: UserFormValues;
  errors: UserFormErrors;
  setField: <K extends keyof UserFormValues>(
    key: K,
    value: UserFormValues[K],
  ) => void;
  validateField: (key: UserFormErrorKey) => void;
  validateAll: () => boolean;
  isSubmittable: boolean;
  buildPayload: (initial?: UserAdminResponse | null) => UserAdminRequest;
  reset: (next?: UserAdminResponse | null) => void;
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

export function useUserForm({ mode }: UseUserFormOptions): UserFormApi {
  const [values, setValues] = useState<UserFormValues>(blankValues);
  const [errors, setErrors] = useState<UserFormErrors>({});

  const setField: UserFormApi["setField"] = (key, value) => {
    setValues((prev) => ({ ...prev, [key]: value }));
    setErrors((prev) => {
      if (!(key in prev) && !isNamespaceLimitKey(key)) return prev;
      const next: UserFormErrors = { ...prev };
      if (isNamespaceLimitKey(key)) delete next.maxNamespaces;
      else delete next[key as UserFormErrorKey];
      return next;
    });
  };

  const computeFieldError = (
    key: UserFormErrorKey,
    v: UserFormValues,
  ): string | null => {
    if (key === "name") return validateName(v.name);
    if (key === "username") return validateUsername(v.username);
    if (key === "email") return validateEmail(v.email);
    if (key === "password") {
      if (mode === "edit" && v.password === "") return null;
      return validatePassword(v.password);
    }
    if (key === "maxNamespaces") {
      return isMaxNamespacesValid(
        v.limitEnabled,
        v.limitDisabled,
        v.maxNamespaces,
      )
        ? null
        : MAX_NAMESPACES_ERROR;
    }
    return null;
  };

  const validateField = (key: UserFormErrorKey) => {
    const err = computeFieldError(key, values);
    setErrors((prev) => {
      const next = { ...prev };
      if (err) next[key] = err;
      else delete next[key];
      return next;
    });
  };

  const validateAll = (): boolean => {
    const next: UserFormErrors = {};
    const keys: UserFormErrorKey[] = [
      "name",
      "username",
      "email",
      "password",
      "maxNamespaces",
    ];
    for (const key of keys) {
      const err = computeFieldError(key, values);
      if (err) next[key] = err;
    }
    setErrors(next);
    return Object.keys(next).length === 0;
  };

  const isSubmittable =
    values.name.trim() !== "" &&
    values.username.trim() !== "" &&
    values.email.trim() !== "" &&
    (mode === "edit" || values.password.trim() !== "") &&
    isMaxNamespacesValid(
      values.limitEnabled,
      values.limitDisabled,
      values.maxNamespaces,
    );

  const computeMaxNamespaces = (
    initial?: UserAdminResponse | null,
  ): number | undefined => {
    if (!values.limitEnabled) {
      if (mode === "edit") {
        const orig = initial?.max_namespaces;
        return orig !== undefined && orig < 0 ? orig : undefined;
      }
      return undefined;
    }
    if (values.limitDisabled) return 0;
    return parseInt(values.maxNamespaces, 10);
  };

  const buildPayload: UserFormApi["buildPayload"] = (initial) => {
    // Always send `password`. On edit, an empty string signals "no change" to
    // the backend (preserves the pre-refactor wire behavior and keeps the
    // generated `UserAdminRequest.password: string` contract honest).
    return {
      name: values.name.trim(),
      username: values.username.trim(),
      email: values.email.trim(),
      password: values.password,
      admin: values.admin,
      max_namespaces: computeMaxNamespaces(initial),
      ...(mode === "edit" && { confirmed: values.confirmed }),
    };
  };

  const reset: UserFormApi["reset"] = (next) => {
    setValues(next ? valuesFromUser(next) : blankValues());
    setErrors({});
  };

  return {
    mode,
    values,
    errors,
    setField,
    validateField,
    validateAll,
    isSubmittable,
    buildPayload,
    reset,
  };
}

function isNamespaceLimitKey(key: keyof UserFormValues): boolean {
  return (
    key === "limitEnabled" || key === "limitDisabled" || key === "maxNamespaces"
  );
}
