import http from '@/store/helpers/http';

export const fetchSessions = async (perPage, page) => http().get(`/sessions?per_page=${perPage}&page=${page}`);

export const getSession = async (uid) => http().get(`/sessions/${uid}`);

export const deleteSessionLogs = async (uid) => http().delete(`/sessions/${uid}/record`);

export const closeSession = async (session) => http().post(`/sessions/${session.uid}/close`, { device: session.device_uid });

export const getLog = async (uid) => http().get(`/sessions/${uid}/play`);
