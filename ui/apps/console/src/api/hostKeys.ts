import {
  scanKnownHost as scanKnownHostSdk,
  acceptKnownHost as acceptKnownHostSdk,
  getKnownHost as getKnownHostSdk,
  deleteKnownHost as deleteKnownHostSdk,
} from "@/client";
import type { KnownHost } from "@/client";

// Known-host (TOFU) endpoints for external connections. The API scope is
// "personal" (per-user) or "namespace" (team-shared) and follows the connection.
export type HostKeyScope = "personal" | "namespace";
export type HostKeyStatus = "unverified" | "trusted" | "changed";

export interface HostKeyScanResult {
  key_type: string;
  fingerprint: string;
  public_key: string;
  status: HostKeyStatus;
  stored: KnownHost | null;
}

export async function scanHostKey(
  host: string,
  port: number,
  scope: HostKeyScope,
): Promise<HostKeyScanResult> {
  const { data } = await scanKnownHostSdk({
    body: { host, port, scope },
    throwOnError: true,
  });

  return data as HostKeyScanResult;
}

export async function acceptHostKey(body: {
  host: string;
  port: number;
  scope: HostKeyScope;
  key_type: string;
  public_key: string;
  fingerprint: string;
}): Promise<KnownHost> {
  const { data } = await acceptKnownHostSdk({ body, throwOnError: true });

  return data;
}

export async function getHostKey(
  host: string,
  port: number,
  scope: HostKeyScope,
): Promise<KnownHost | null> {
  const { data } = await getKnownHostSdk({
    query: { host, port, scope },
    throwOnError: true,
  });

  return data ?? null;
}

export async function forgetHostKey(
  host: string,
  port: number,
  scope: HostKeyScope,
): Promise<void> {
  await deleteKnownHostSdk({
    query: { host, port, scope },
    throwOnError: true,
  });
}
