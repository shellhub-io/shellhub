import { z } from "zod";
import type { PublicKeyRequest, PublicKeyResponse } from "@/client";
import { isPublicKeyValid } from "@/utils/sshKeys";
import { validateName } from "@/utils/validation";

/**
 * Whether the key form is creating or editing, which decides if the key material is editable.
 */
export type KeyMode = "create" | "edit";

const keyFields = z.object({
  name: z.string(),
  data: z.string(),
  usernameOption: z.enum(["all", "username"]),
  username: z.string(),
  filterOption: z.enum(["all", "hostname", "tags"]),
  hostname: z.string(),
  tags: z.array(z.string()),
});

/**
 * The key form's values, derived from the schema.
 */
export type KeyFormValues = z.infer<typeof keyFields>;

/**
 * What a new key form starts as: no filter and every username, which is the widest setting and
 * so is left for the user to narrow deliberately.
 */
export const DEFAULT_VALUES: KeyFormValues = {
  name: "",
  data: "",
  usernameOption: "all",
  username: "",
  filterOption: "all",
  hostname: "",
  tags: [],
};

/**
 * Builds the key schema for a mode. On edit the key material is not revalidated, since it cannot
 * be changed.
 */
export function keySchema(mode: KeyMode) {
  return keyFields.superRefine((values, ctx) => {
    const nameError = validateName(values.name);
    if (nameError) {
      ctx.addIssue({ code: "custom", path: ["name"], message: nameError });
    }

    if (mode === "create") {
      if (!values.data.trim()) {
        ctx.addIssue({ code: "custom", path: ["data"], message: "Public key data is required" });
      } else if (!isPublicKeyValid(values.data)) {
        ctx.addIssue({ code: "custom", path: ["data"], message: "This is not a valid public key" });
      }
    }

    if (values.usernameOption === "username" && !values.username.trim()) {
      ctx.addIssue({ code: "custom", path: ["username"], message: "Username pattern is required" });
    }

    if (values.filterOption === "hostname" && !values.hostname.trim()) {
      ctx.addIssue({ code: "custom", path: ["hostname"], message: "Hostname pattern is required" });
    }

    if (values.filterOption === "tags") {
      if (values.tags.length === 0) {
        ctx.addIssue({ code: "custom", path: ["tags"], message: "Select at least one tag" });
      } else if (values.tags.length > 3) {
        ctx.addIssue({ code: "custom", path: ["tags"], message: "You can select up to 3 tags" });
      }
    }
  });
}

/**
 * Turns validated form values into the request body, sending only the filter branch that was
 * chosen.
 */
export function buildKeyBody(values: KeyFormValues): PublicKeyRequest {
  const filter: PublicKeyRequest["filter"] =
    values.filterOption === "hostname"
      ? { hostname: values.hostname }
      : values.filterOption === "tags"
        ? { tags: values.tags }
        : { hostname: ".*" };

  return {
    name: values.name.trim(),
    data: btoa(values.data.trim()),
    username:
      values.usernameOption === "username" ? values.username.trim() : ".*",
    filter,
  };
}

/**
 * Fills the key form from an existing key. The stored key material is base64, so it is decoded
 * for display; a value that does not decode is shown as-is rather than blanked.
 */
export function buildKeyDefaults(key: PublicKeyResponse): KeyFormValues {
  const decodedData = (() => {
    try {
      return atob(key.data);
    } catch {
      return key.data;
    }
  })();

  const hasTags = key.filter.tags && key.filter.tags.length > 0;
  const hasHostname =
    key.filter.hostname !== undefined && key.filter.hostname !== ".*";

  const filterOption: KeyFormValues["filterOption"] = hasTags
    ? "tags"
    : hasHostname
      ? "hostname"
      : "all";

  return {
    name: key.name,
    data: decodedData,
    usernameOption: key.username === ".*" ? "all" : "username",
    username: key.username === ".*" ? "" : key.username,
    filterOption,
    hostname: filterOption === "hostname" ? (key.filter.hostname ?? "") : "",
    tags: hasTags ? key.filter.tags.map((t) => t.name) : [],
  };
}
