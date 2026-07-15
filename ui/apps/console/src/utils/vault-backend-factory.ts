import { isEnterpriseOrCloud } from "@/env";
import type { IVaultBackend } from "@/utils/vault-backend";
import {
  LocalVaultBackend,
  localVaultExists,
} from "@/utils/vault-backend-local";
import { ServerVaultBackend } from "@/utils/vault-backend-server";

export type VaultStorageMode = "local" | "server";
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

export function dismissVaultSyncPromo(scope?: VaultScope): void {
  localStorage.setItem(promoKey(scope), "true");
}

export function getVaultBackend(scope?: VaultScope): IVaultBackend {
  if (getVaultStorageMode(scope) === "server")
    return new ServerVaultBackend(scope);
  return new LocalVaultBackend(scope);
}
