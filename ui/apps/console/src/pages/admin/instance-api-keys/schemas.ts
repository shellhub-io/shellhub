import { z } from "zod";

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
  name: z
    .string()
    .regex(
      /^[a-zA-Z0-9_-]{3,20}$/,
      "Name must be 3-20 characters: letters, numbers, - and _ only.",
    ),
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
 * drawer, so the group is never in the empty state that reads as "selecting this is optional".
 */
export const GENERATE_INSTANCE_KEY_DEFAULTS: GenerateInstanceKeyFormValues = {
  name: "",
  expiresAt: "30",
};

/**
 * A chosen expiry, once the form has been validated.
 */
export type InstanceKeyExpiry = GenerateInstanceKeyFormValues["expiresAt"];

/**
 * Builds the create request body, converting the radio group's string back into the number the
 * API expects.
 */
export function buildGenerateInstanceKeyBody(values: {
  name: string;
  expiresAt: InstanceKeyExpiry;
}) {
  return {
    name: values.name.trim(),
    expires_at: Number(values.expiresAt) as 30 | 60 | 90 | 365,
  };
}
