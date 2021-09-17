import http from '@/store/helpers/http';

export const fetchDevices = async (
  perPage,
  page,
  search,
  status,
  sortStatusField,
  sortStatusString,
) => {
  let query = `/devices?per_page=${perPage}&page=${page}&status=${status}`;

  if (search !== null) {
    query += `&filter=${search}`;
  }

  if (sortStatusField !== null) {
    query += `&sort_by=${sortStatusField}&order_by=${sortStatusString}`;
  }
  return http().get(query);
};

export const getDevice = async (uid) => http().get(`/devices/${uid}`);

export const renameDevice = async (data) => http().patch(`/devices/${data.uid}`, { name: data.name });

export const acceptDevice = async (uid) => http().patch(`/devices/${uid}/accept`);

export const rejectDevice = async (uid) => http().patch(`/devices/${uid}/reject`);

export const updateDeviceTag = async (data) => http().put(`/devices/${data.uid}/tags`, { tags: data.tags });

export const removeDevice = async (uid) => http().delete(`/devices/${uid}`);
