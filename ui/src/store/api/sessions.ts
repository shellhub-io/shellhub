import { sessionsApi } from "@/api/http";
import { ISessions } from "@/interfaces/ISessions";

export const fetchSessions = async (page: number, perPage: number) => sessionsApi.getSessions(page, perPage);

export const getSession = async (uid: string) => sessionsApi.getSession(uid);

export const deleteSessionLogs = async (uid: string) => sessionsApi.clsoeSession(uid);

export const closeSession = async (
  session: ISessions,
) => sessionsApi.clsoeSession(session.uid, { device: session.device_uid });

/* Uses fetch instead of the OpenAPI spec method because the player uses a ReadableStream.
* Axios does not support ReadableStreams in the browser yet. */
export const getLog = async (uid: string) => {
  const configuration = sessionsApi.getConfiguration();
  return fetch(`/api/sessions/${uid}/records/0`, { headers: { Authorization: `Bearer ${configuration?.accessToken}` } });
};
