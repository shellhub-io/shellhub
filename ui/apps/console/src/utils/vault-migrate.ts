import { LocalVaultBackend } from "@/utils/vault-backend-local";
import { ServerVaultBackend } from "@/utils/vault-backend-server";
import {
  setVaultStorageMode,
  type VaultScope,
} from "@/utils/vault-backend-factory";

export { localVaultExists } from "@/utils/vault-backend-local";

/** Whether a vault already exists on the server for the current user. */
export async function serverVaultExists(scope?: VaultScope): Promise<boolean> {
  const server = new ServerVaultBackend(scope);
  return (await server.loadMeta()) !== null;
}

/**
 * Copies the local vault to the server, clears the local copy, and switches
 * the storage mode to "server". When the server already has a vault it is
 * replaced — callers must get explicit confirmation for that first (see
 * [serverVaultExists]).
 */
export async function migrateLocalToServer(scope?: VaultScope): Promise<void> {
  const local = new LocalVaultBackend(scope);
  const server = new ServerVaultBackend(scope);

  const meta = await local.loadMeta();
  if (!meta) throw new Error("No local vault to migrate.");
  const data = await local.loadData();
  const settings = await local.loadSettings();

  await server.clear();

  await server.saveMeta(meta);
  if (data) await server.saveData(data);
  await server.saveSettings(settings);

  await local.clear();
  setVaultStorageMode("server", scope);
}

/**
 * Adopts the existing server vault, discarding the local one, and switches
 * the storage mode to "server". Used when migrating and the user chooses the
 * vault that already exists on their account.
 */
export async function adoptServerVault(scope?: VaultScope): Promise<void> {
  const local = new LocalVaultBackend(scope);
  await local.clear();
  setVaultStorageMode("server", scope);
}

/**
 * Copies the server vault to local storage, removes it from the server, and
 * switches the storage mode to "local".
 */
export async function migrateServerToLocal(scope?: VaultScope): Promise<void> {
  const local = new LocalVaultBackend(scope);
  const server = new ServerVaultBackend(scope);

  const meta = await server.loadMeta();
  if (!meta) throw new Error("No synced vault to migrate.");
  const data = await server.loadData();
  const settings = await server.loadSettings();

  await local.saveMeta(meta);
  if (data) await local.saveData(data);
  await local.saveSettings(settings);

  await server.clear();
  setVaultStorageMode("local", scope);
}
