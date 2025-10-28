import { adminApi } from "@/api/http";

const fetchSessions = async (perPage: number, page: number) => adminApi.getSessionsAdmin(page, perPage);

const getSession = async (uid: string) => adminApi.getSessionAdmin(uid);

export { fetchSessions, getSession };
