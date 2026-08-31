import { isEnterpriseOrCloud } from "@/env";
import type { IVaultBackend } from "@/utils/vault-backend";
import {
  LocalVaultBackend,
  localVaultExists,
} from "@/utils/vault-backend-local";
import { ServerVaultBackend } from "@/utils/vault-backend-server";

/**
 * Where a vault is kept: in this browser only, or synced through the server.
 */
export type VaultStorageMode = "local" | "server";
/**
 * Whose vault. Every stored key is scoped by user and tenant, so two accounts on one browser,
 * or one account in two namespaces, never read each other's vault.
 */
export type VaultScope = { user: string; tenant: string };

const VAULT_STORAGE_MODE_KEY = "shellhub-vault-storage";

function modeKey(scope?: VaultScope): string {
  return scope
    ? `${VAULT_STORAGE_MODE_KEY}:${scope.user}:${scope.tenant}`
    : VAULT_STORAGE_MODE_KEY;
}

/**
 * Whether server-side vault storage is available in this deployment
 * (Cloud/Enterprise). Community Edition is always local.
 */
export function isVaultServerEnabled(): boolean {
  return isEnterpriseOrCloud();
}

/**
 * The storage mode in effect for the given scope. The user's explicit choice
 * (persisted per user/namespace) wins; without one, an existing local vault
 * keeps using local storage — switching to the server is an explicit
 * migration — and new vaults default to the server when available.
 */
export function getVaultStorageMode(scope?: VaultScope): VaultStorageMode {
  if (!isVaultServerEnabled()) return "local";

  const stored = localStorage.getItem(modeKey(scope));
  if (stored === "local" || stored === "server") return stored;

  return localVaultExists(scope) ? "local" : "server";
}

/**
 * Records the chosen storage mode for a scope. Only the preference is written here; moving the
 * contents between backends is the caller's job.
 */
export function setVaultStorageMode(
  mode: VaultStorageMode,
  scope?: VaultScope,
): void {
  localStorage.setItem(modeKey(scope), mode);
}

const VAULT_SYNC_PROMO_KEY = "shellhub-vault-sync-promo-dismissed";

function promoKey(scope?: VaultScope): string {
  return scope
    ? `${VAULT_SYNC_PROMO_KEY}:${scope.user}:${scope.tenant}`
    : VAULT_SYNC_PROMO_KEY;
}

/** Whether the user opted out of the "sync your vault" prompt shown on lock. */
export function isVaultSyncPromoDismissed(scope?: VaultScope): boolean {
  return localStorage.getItem(promoKey(scope)) === "true";
}

/**
 * Remembers that the sync promotion was dismissed, so it is not offered again for this scope.
 */
export function dismissVaultSyncPromo(scope?: VaultScope): void {
  localStorage.setItem(promoKey(scope), "true");
}

/**
 * The backend for a scope, according to its stored mode. Local is the default, so a scope that
 * has never chosen keeps its vault in the browser.
 */
export function getVaultBackend(scope?: VaultScope): IVaultBackend {
  if (getVaultStorageMode(scope) === "server")
    return new ServerVaultBackend(scope);
  return new LocalVaultBackend(scope);
}
