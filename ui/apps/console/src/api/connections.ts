import { client } from "../client/client.gen";
import type { Connection } from "@/types/connection";

// These endpoints are not part of the generated OpenAPI SDK yet, so they call
// the hey-api client directly. Auth and tenant are applied by the shared client
// interceptors, same as the generated SDK functions.

export async function listConnections(): Promise<Connection[]> {
  const { data } = await client.get({
    security: [
      { scheme: "bearer", type: "http" },
      { name: "X-API-KEY", type: "apiKey" },
    ],
    url: "/api/connections",
    query: { page: 1, per_page: 100 },
    throwOnError: true,
  });

  return (data as Connection[]) ?? [];
}

export interface ConnectionBody {
  label: string;
  username?: string;
  kind: "direct" | "device";
  host?: string;
  port?: number;
  device_uid?: string;
  /** Save a direct connection even if the target is currently unreachable. */
  force?: boolean;
}

export async function createConnection(
  body: ConnectionBody,
): Promise<Connection> {
  const { data } = await client.post({
    security: [
      { scheme: "bearer", type: "http" },
      { name: "X-API-KEY", type: "apiKey" },
    ],
    url: "/api/connections",
    body,
    headers: { "Content-Type": "application/json" },
    throwOnError: true,
  });

  return data as Connection;
}

export async function updateConnection(
  id: string,
  body: ConnectionBody,
): Promise<Connection> {
  const { data } = await client.put({
    security: [
      { scheme: "bearer", type: "http" },
      { name: "X-API-KEY", type: "apiKey" },
    ],
    url: `/api/connections/${id}`,
    body,
    headers: { "Content-Type": "application/json" },
    throwOnError: true,
  });

  return data as Connection;
}

export async function getConnectionStatus(id: string): Promise<boolean> {
  const { data } = await client.get({
    security: [
      { scheme: "bearer", type: "http" },
      { name: "X-API-KEY", type: "apiKey" },
    ],
    url: `/api/connections/${id}/status`,
    throwOnError: true,
  });

  return (data as { online?: boolean })?.online ?? false;
}

export async function deleteConnection(id: string): Promise<void> {
  await client.delete({
    security: [
      { scheme: "bearer", type: "http" },
      { name: "X-API-KEY", type: "apiKey" },
    ],
    url: `/api/connections/${id}`,
    throwOnError: true,
  });
}
