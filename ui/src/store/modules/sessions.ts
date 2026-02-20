import { defineStore } from "pinia";
import { ref } from "vue";
import * as sessionsApi from "../api/sessions";
import { parseTotalCount } from "@/utils/headers";
import { ISession } from "@/interfaces/ISession";

const useSessionsStore = defineStore("sessions", () => {
  const sessions = ref<Array<ISession>>([]);
  const session = ref<ISession>({} as ISession);
  const sessionCount = ref(0);

  const fetchSessionList = async (data?: { page: number; perPage: number }) => {
    try {
      const res = await sessionsApi.fetchSessions(data?.page || 1, data?.perPage || 10);
      sessions.value = res.data as ISession[];
      sessionCount.value = parseTotalCount(res.headers);
    } catch (error) {
      sessions.value = [];
      sessionCount.value = 0;
      throw error;
    }
  };

  const getSession = async (uid: string) => {
    try {
      const res = await sessionsApi.getSession(uid);
      session.value = res.data as ISession;
    } catch (error) {
      session.value = {} as ISession;
      throw error;
    }
  };

  const getSessionLogs = async (uid: string) => {
    const res = await sessionsApi.getLog(uid);
    return res.data as unknown as string;
  };

  const closeSession = async (sessionData: Pick<ISession, "uid" | "device_uid">) => {
    await sessionsApi.closeSession(sessionData);
  };

  const deleteSessionLogs = async (uid: string) => {
    await sessionsApi.deleteSessionLogs(uid);
    session.value = {
      ...session.value,
      recorded: false,
    };
  };

  return {
    sessions,
    session,
    sessionCount,

    fetchSessionList,
    getSession,
    getSessionLogs,
    closeSession,
    deleteSessionLogs,
  };
});

export default useSessionsStore;
