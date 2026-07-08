import { z } from "zod";
import { validateNamespaceName } from "@/utils/validation";
import type { Namespace } from "@/client";

const editNamespaceFields = z.object({
  name: z.string(),
  maxDevices: z.string(),
  sessionRecord: z.boolean(),
});

export type EditNamespaceFormValues = z.infer<typeof editNamespaceFields>;

const MAX_DEVICES_ERROR =
  "Max devices must be a number greater than or equal to -1";

/**
 * The name is only re-validated when it actually changes, so a namespace whose
 * stored name predates the current rules can still be saved without edits.
 */
export function editNamespaceSchema(originalName: string) {
  return editNamespaceFields.superRefine((values, ctx) => {
    if (values.name !== originalName) {
      const nameError = validateNamespaceName(values.name);
      if (nameError)
        ctx.addIssue({ code: "custom", path: ["name"], message: nameError });
    }

    if (!(parseInt(values.maxDevices, 10) >= -1)) {
      ctx.addIssue({
        code: "custom",
        path: ["maxDevices"],
        message: MAX_DEVICES_ERROR,
      });
    }
  });
}

export function buildEditNamespaceDefaults(
  namespace: Namespace | null,
): EditNamespaceFormValues {
  return {
    name: namespace?.name ?? "",
    maxDevices: String(namespace?.max_devices ?? -1),
    sessionRecord: namespace?.settings?.session_record ?? false,
  };
}

/**
 * The SDK types the edit body as a full Namespace, so the original entity is
 * spread in and only the editable fields are overwritten.
 */
export function buildEditNamespaceBody(
  namespace: Namespace,
  values: EditNamespaceFormValues,
): Namespace {
  return {
    ...namespace,
    name: values.name.trim(),
    max_devices: parseInt(values.maxDevices, 10),
    settings: {
      connection_announcement:
        namespace.settings?.connection_announcement ?? "",
      session_record: values.sessionRecord,
      ssh_access_mode: namespace.settings?.ssh_access_mode ?? "legacy",
    },
  };
}
