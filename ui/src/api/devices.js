import http from '@/helpers/http';

export const fetchDevices = async (perPage, page, search, status, sortStatusField,
  sortStatusString) => {
  let query = '';
  let concatSortString = '';
  if (sortStatusField !== null) {
    concatSortString = `&sort_by=${sortStatusField}&order_by=${sortStatusString}`;
  }
  if (search === null) {
    query = `/devices?per_page=${perPage}&page=${page}&status=${status}${concatSortString}`;
  } else {
    query = `/devices?per_page=${perPage}&page=${page}&filter=${search}&status=${status}${concatSortString}`;
  }
  return http().get(query);
};

export const removeDevice = async (uid) => http().delete(`/devices/${uid}`);

export const renameDevice = async (data) => http().patch(`/devices/${data.uid}`, { name: data.name });

export const getDevice = async (uid) => http().get(`/devices/${uid}`);

export const acceptDevice = async (uid) => http().patch(`/devices/${uid}/accept`);

export const rejectDevice = async (uid) => http().patch(`/devices/${uid}/reject`);
