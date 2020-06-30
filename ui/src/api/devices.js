import http from '@/helpers/http';

export const fetchDevices = async (perPage, page, search, status) => {
  let query = '';
  if (search === null) {
    query = `/devices?per_page=${perPage}&page=${page}&status=${status}`;
  } else {
    query = `/devices?per_page=${perPage}&page=${page}&filter=${search}&status=${status}`;
  }
  return http().get(query);
};

export const removeDevice = async (uid) => http().delete(`/devices/${uid}`);

export const renameDevice = async (data) => http().patch(`/devices/${data.uid}`, { name: data.name });

export const getDevice = async (uid) => http().get(`/devices/${uid}`);

export const acceptDevice = async (uid) => http().patch(`/devices/${uid}/accept`);
