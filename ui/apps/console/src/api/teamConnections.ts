import {
  createTeamConnection as createTeamConnectionSdk,
  updateTeamConnection as updateTeamConnectionSdk,
  deleteTeamConnection as deleteTeamConnectionSdk,
  getTeamConnectionStatus as getTeamConnectionStatusSdk,
  getTeamConnectionPrefs as getTeamConnectionPrefsSdk,
  updateTeamConnectionPrefs as updateTeamConnectionPrefsSdk,
} from "@/client";
import type { TeamConnection, TeamConnectionPrefs } from "@/client";

// Team connections are an Enterprise/Cloud feature, license-gated (402 on an
// edition without it). Maps the generated SDK types to the app types at the
// boundary.

export interface TeamConnectionBody {
  label: string;
  kind: "external" | "device";
  host?: string;
  port?: number;
  device_uid?: string;
}

/** Per-user auth preference payload for a team connection (never the secret). */
export interface TeamConnectionPrefsBody {
  username?: string;
  auth_method?: string;
  key_fingerprint?: string;
}

export async function createTeamConnection(
  body: TeamConnectionBody,
): Promise<TeamConnection> {
  const { data } = await createTeamConnectionSdk({ body, throwOnError: true });

  return data;
}

export async function updateTeamConnection(
  id: string,
  body: TeamConnectionBody,
): Promise<TeamConnection> {
  const { data } = await updateTeamConnectionSdk({
    path: { id },
    body,
    throwOnError: true,
  });

  return data;
}

export async function deleteTeamConnection(id: string): Promise<void> {
  await deleteTeamConnectionSdk({ path: { id }, throwOnError: true });
}

export async function getTeamConnectionStatus(id: string): Promise<boolean> {
  const { data } = await getTeamConnectionStatusSdk({
    path: { id },
    throwOnError: true,
  });

  return data?.online ?? false;
}

export async function getTeamConnectionPrefs(
  id: string,
): Promise<TeamConnectionPrefs> {
  const { data } = await getTeamConnectionPrefsSdk({
    path: { id },
    throwOnError: true,
  });

  return data;
}

export async function putTeamConnectionPrefs(
  id: string,
  body: TeamConnectionPrefsBody,
): Promise<TeamConnectionPrefs> {
  const { data } = await updateTeamConnectionPrefsSdk({
    path: { id },
    body,
    throwOnError: true,
  });

  return data;
}
