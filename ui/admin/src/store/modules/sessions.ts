import { defineStore } from "pinia";
import { ref } from "vue";
import { IAdminSession } from "@admin/interfaces/ISession";
import * as sessionsApi from "../api/sessions";

const useSessionsStore = defineStore("adminSessions", () => {
  const sessions = ref<Array<IAdminSession>>([]);
  const sessionCount = ref<number>(0);
  const session = ref<IAdminSession>({} as IAdminSession);

  const fetchSessionList = async (data: { perPage: number; page: number }) => {
    const res = await sessionsApi.fetchSessions(data.perPage, data.page);

    sessions.value = res.data as Array<IAdminSession>;
    sessionCount.value = parseInt(res.headers["x-total-count"] as string, 10);
  };

  const fetchSessionById = async (uid: string) => {
    const { data } = await sessionsApi.getSession(uid);
    session.value = data as IAdminSession;
  };

  return {
    sessions,
    sessionCount,
    session,
    fetchSessionList,
    fetchSessionById,
  };
});

export default useSessionsStore;
