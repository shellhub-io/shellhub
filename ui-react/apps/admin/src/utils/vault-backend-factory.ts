import type { IVaultBackend } from "@/utils/vault-backend";
import { LocalVaultBackend } from "@/utils/vault-backend-local";

let instance: IVaultBackend | null = null;

export function getVaultBackend(): IVaultBackend {
  if (!instance) {
    // Future: select ServerVaultBackend for cloud/enterprise via getConfig()
    instance = new LocalVaultBackend();
  }
  return instance;
}

export function resetVaultBackend(): void {
  instance = null;
}
