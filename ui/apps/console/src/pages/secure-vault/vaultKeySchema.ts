import { z } from "zod";
import { validatePrivateKey } from "@/utils/sshKeys";
import type { VaultKeyEntry } from "@/types/vault";

export const vaultKeySchema = z
  .object({
    name: z.string(),
    data: z.string(),
    /** Derived from parsing the key; drives the passphrase field. */
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

export type VaultKeyFormValues = z.infer<typeof vaultKeySchema>;

export type VaultKeyPayload = Pick<
  VaultKeyEntry,
  "name" | "data" | "hasPassphrase" | "fingerprint" | "algorithm"
>;

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
