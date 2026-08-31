import { z } from "zod";
import { validatePrivateKey } from "@/utils/sshKeys";
import type { VaultKeyEntry } from "@/types/vault";

/**
 * Validates a vault key form, including parsing the key itself — an unparseable key is rejected
 * here rather than stored and found broken at connect time.
 */
export const vaultKeySchema = z
  .object({
    name: z.string(),
    data: z.string(),
    encrypted: z.boolean(),
    passphrase: z.string(),
  })
  .superRefine((values, ctx) => {
    if (!values.name.trim()) {
      ctx.addIssue({ code: "custom", path: ["name"], message: "Name is required" });
    }

    const data = values.data.trim();
    if (!data) {
      ctx.addIssue({ code: "custom", path: ["data"], message: "Private key is required" });
    } else {
      const result = validatePrivateKey(data);
      if (!result.valid) {
        ctx.addIssue({
          code: "custom",
          path: ["data"],
          message: result.error ?? "Invalid private key format.",
        });
      }
    }

    if (values.encrypted && !values.passphrase.trim()) {
      ctx.addIssue({ code: "custom", path: ["passphrase"], message: "Passphrase is required" });
    }
  });

/**
 * The vault key form's values, derived from the schema.
 */
export type VaultKeyFormValues = z.infer<typeof vaultKeySchema>;

/**
 * What is actually stored for a key. Narrower than the form: the passphrase is used to parse the
 * key and then dropped, never written to the vault.
 */
export type VaultKeyPayload = Pick<
  VaultKeyEntry,
  "name" | "data" | "hasPassphrase" | "fingerprint" | "algorithm"
>;

/**
 * Fills the vault key form from an existing entry, or with blanks for a new one.
 */
export function buildVaultKeyDefaults(
  entry: VaultKeyEntry | null,
): VaultKeyFormValues {
  return {
    name: entry?.name ?? "",
    data: entry?.data ?? "",
    encrypted: entry?.hasPassphrase ?? false,
    passphrase: "",
  };
}

/**
 * Builds the stored entry from validated values plus the fingerprint and algorithm read out of
 * the key, so both are derived from the key material rather than typed.
 */
export function buildVaultKeyPayload(
  values: VaultKeyFormValues,
  fingerprint: string,
  algorithm: string,
): VaultKeyPayload {
  return {
    name: values.name.trim(),
    data: values.data.trim(),
    hasPassphrase: values.encrypted,
    fingerprint,
    algorithm,
  };
}
