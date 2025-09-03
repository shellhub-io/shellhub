import { sessionsApi } from "@/api/http";
import { ISession } from "@/interfaces/ISession";

export const fetchSessions = async (page: number, perPage: number) => sessionsApi.getSessions(page, perPage);

export const getSession = async (uid: string) => sessionsApi.getSession(uid);

export const closeSession = async (
  session: Pick<ISession, "uid" | "device_uid">,
) => sessionsApi.clsoeSession(session.uid, { device: session.device_uid });

export const getLog = async (uid: string) => sessionsApi.getSessionRecord(uid, 0);

export const deleteSessionLogs = async (uid: string) => sessionsApi.deleteSessionRecord(uid, 0);
