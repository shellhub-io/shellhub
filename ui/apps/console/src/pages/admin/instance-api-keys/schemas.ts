import { z } from "zod";
import { validateKeyName } from "@/utils/validation";

/**
 * The expiry choices for an instance API key. There is no "never": an instance-wide credential
 * that never expires is the one most likely to be forgotten and later found somewhere it should
 * not be.
 */
export const INSTANCE_KEY_EXPIRY_OPTIONS = [
  { label: "30 days", value: 30 },
  { label: "60 days", value: 60 },
  { label: "90 days", value: 90 },
  { label: "1 year", value: 365 },
] as const;

/**
 * Validates the generate-instance-key form.
 */
export const generateInstanceKeySchema = z.object({
  name: z.string().superRefine((value, ctx) => {
    const nameError = validateKeyName(value);
    if (nameError) ctx.addIssue({ code: "custom", message: nameError });
  }),
  expiresAt: z.enum(["30", "60", "90", "365"]),
});

/**
 * The generate-instance-key form's values, derived from the schema.
 */
export type GenerateInstanceKeyFormValues = z.infer<
  typeof generateInstanceKeySchema
>;

/**
 * What the form starts as. The expiry opens on the shortest option, matching the namespace key
 * modal, so the group is never in the empty state that reads as "selecting this is optional".
 */
export const GENERATE_INSTANCE_KEY_DEFAULTS: GenerateInstanceKeyFormValues = {
  name: "",
  expiresAt: "30",
};

/**
 * Builds the create request body, converting the radio group's string back into the number the
 * API expects.
 */
export function buildGenerateInstanceKeyBody(
  values: GenerateInstanceKeyFormValues,
) {
  return {
    name: values.name.trim(),
    expires_at: Number(values.expiresAt) as 30 | 60 | 90 | 365,
  };
}
