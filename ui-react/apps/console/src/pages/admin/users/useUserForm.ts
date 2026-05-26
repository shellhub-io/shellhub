import { useState } from "react";
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

export interface UseUserFormOptions<M extends UserFormMode> {
  mode: M;
}

export interface UserFormApi<M extends UserFormMode> {
  mode: M;
  values: UserFormValues;
  errors: UserFormErrors;
  setField: <K extends keyof UserFormValues>(
    key: K,
    value: UserFormValues[K],
  ) => void;
  validateField: (key: UserFormErrorKey) => void;
  validateAll: () => boolean;
  isSubmittable: boolean;
  buildPayload: (initial?: UserAdminResponse | null) => UserFormPayload<M>;
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

export function useUserForm<M extends UserFormMode>({
  mode,
}: UseUserFormOptions<M>): UserFormApi<M> {
  const [values, setValues] = useState<UserFormValues>(blankValues);
  const [errors, setErrors] = useState<UserFormErrors>({});

  const setField: UserFormApi<M>["setField"] = (key, value) => {
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

  const buildPayload: UserFormApi<M>["buildPayload"] = (initial) => {
    const base = {
      name: values.name.trim(),
      username: values.username.trim(),
      email: values.email.trim(),
      admin: values.admin,
      max_namespaces: computeMaxNamespaces(initial),
    };

    if (mode === "create") {
      const payload: UserAdminCreateRequest = {
        ...base,
        password: values.password,
      };
      return payload;
    }

    const payload: UserAdminUpdateRequest = {
      ...base,
      confirmed: values.confirmed,
      // Omit when blank so the backend keeps the current password.
      ...(values.password !== "" && { password: values.password }),
    };
    return payload as UserFormPayload<M>;
  };

  const reset: UserFormApi<M>["reset"] = (next) => {
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
