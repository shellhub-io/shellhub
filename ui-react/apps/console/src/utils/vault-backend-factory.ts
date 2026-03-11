import type { IVaultBackend } from "@/utils/vault-backend";
import { LocalVaultBackend } from "@/utils/vault-backend-local";

export function getVaultBackend(scope?: { user: string; tenant: string }): IVaultBackend {
  return new LocalVaultBackend(scope);
}
