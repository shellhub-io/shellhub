import { z } from "zod";
import { validateNamespaceName } from "@/utils/validation";

export const namespaceRenameSchema = z
  .object({
    name: z.string(),
  })
  .superRefine((values, ctx) => {
    const nameError = validateNamespaceName(values.name);
    if (nameError) ctx.addIssue({ code: "custom", path: ["name"], message: nameError });
  });

export type NamespaceRenameFormValues = z.infer<typeof namespaceRenameSchema>;

export function buildNamespaceRenameDefaults(currentName: string): NamespaceRenameFormValues {
  return { name: currentName };
}
