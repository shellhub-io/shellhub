import { z } from "zod";
import { validateNamespaceName } from "@/utils/validation";

/**
 * Validates the namespace rename form against the same rules the create form uses, since the
 * name ends up in an SSHID either way.
 */
export const namespaceRenameSchema = z
  .object({
    name: z.string(),
  })
  .superRefine((values, ctx) => {
    const nameError = validateNamespaceName(values.name);
    if (nameError) ctx.addIssue({ code: "custom", path: ["name"], message: nameError });
  });

/**
 * The namespace rename form's values, derived from the schema.
 */
export type NamespaceRenameFormValues = z.infer<typeof namespaceRenameSchema>;

/**
 * Fills the rename form with the current name, so the field opens showing what it is changing.
 */
export function buildNamespaceRenameDefaults(currentName: string): NamespaceRenameFormValues {
  return { name: currentName };
}
