import type { VaultMeta, VaultData, VaultSettings } from "@/types/vault";
import { ALLOWED_TIMEOUT_MINUTES, DEFAULT_VAULT_SETTINGS } from "@/types/vault";

function safeParse(raw: string | null | undefined): unknown {
  if (!raw) return null;
  try {
    return JSON.parse(raw);
  } catch {
    return null;
  }
}

/**
 * Reads a stored vault header, returning null unless every field is present and the version is
 * one this build understands. Strict on purpose: a header that is trusted and wrong derives a
 * key that silently fails to decrypt anything.
 */
export function parseVaultMeta(
  raw: string | null | undefined,
): VaultMeta | null {
  const parsed = safeParse(raw) as Record<string, unknown> | null;
  if (
    !parsed ||
    parsed.version !== 1 ||
    typeof parsed.salt !== "string" ||
    typeof parsed.iterations !== "number" ||
    !Number.isInteger(parsed.iterations) ||
    parsed.iterations < 100_000 ||
    parsed.iterations > 10_000_000 ||
    typeof parsed.verifier !== "string" ||
    typeof parsed.verifierIv !== "string"
  )
    return null;
  return parsed as unknown as VaultMeta;
}

/**
 * Reads a stored vault body, returning null unless both the IV and the ciphertext are present.
 */
export function parseVaultData(
  raw: string | null | undefined,
): VaultData | null {
  const parsed = safeParse(raw) as Record<string, unknown> | null;
  if (
    !parsed ||
    typeof parsed.iv !== "string" ||
    typeof parsed.ciphertext !== "string"
  )
    return null;
  return parsed as unknown as VaultData;
}

/**
 * Reads the stored vault settings, filling in defaults for anything missing or out of range.
 * Unlike the header this never returns null: a broken setting must not stop the vault opening.
 */
export function parseVaultSettings(
  raw: string | null | undefined,
): VaultSettings {
  const parsed = safeParse(raw) as Record<string, unknown> | null;
  if (!parsed) return { ...DEFAULT_VAULT_SETTINGS };

  const rawTimeout = parsed.autoLockTimeoutMinutes;
  const autoLockTimeoutMinutes =
    typeof rawTimeout === "number" &&
    (ALLOWED_TIMEOUT_MINUTES as readonly number[]).includes(rawTimeout)
      ? rawTimeout
      : DEFAULT_VAULT_SETTINGS.autoLockTimeoutMinutes;

  const lockOnHidden =
    typeof parsed.lockOnHidden === "boolean"
      ? parsed.lockOnHidden
      : DEFAULT_VAULT_SETTINGS.lockOnHidden;

  return { autoLockTimeoutMinutes, lockOnHidden };
}
