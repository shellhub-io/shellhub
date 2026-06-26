import {
  createConnection as createConnectionSdk,
  updateConnection as updateConnectionSdk,
  deleteConnection as deleteConnectionSdk,
  getConnectionStatus as getConnectionStatusSdk,
} from "@/client";
import type { Connection } from "@/client";

// Maps the generated (all-optional) SDK response types to the app's Connection
// type at the boundary, so the rest of the app keeps working with required
// fields.

export interface ConnectionBody {
  label: string;
  kind: "external" | "device";
  host?: string;
  port?: number;
  device_uid?: string;
  username?: string;
  /** "" | "password" | "key" */
  auth_method?: string;
  /** Points at the SSH key to use (never the secret). */
  key_fingerprint?: string;
  /** Save an external connection even if the target is currently unreachable. */
  force?: boolean;
}

export async function createConnection(
  body: ConnectionBody,
): Promise<Connection> {
  const { data } = await createConnectionSdk({ body, throwOnError: true });

  return data;
}

export async function updateConnection(
  id: string,
  body: ConnectionBody,
): Promise<Connection> {
  const { data } = await updateConnectionSdk({
    path: { id },
    body,
    throwOnError: true,
  });

  return data;
}

export async function getConnectionStatus(id: string): Promise<boolean> {
  const { data } = await getConnectionStatusSdk({
    path: { id },
    throwOnError: true,
  });

  return data?.online ?? false;
}

export async function deleteConnection(id: string): Promise<void> {
  await deleteConnectionSdk({ path: { id }, throwOnError: true });
}
