import http from '@/helpers/http';

export const
  fetchDevices = async (perPage, page) => {
    return http().get(`/devices?per_page=${perPage}&page=${page}`);
  },

  removeDevice = async (uid) => {
    return http().delete(`/devices/${uid}`);
  },

  renameDevice = async (data) => {
    return http().patch(`/devices/${data.uid}`, { name: data.name });
  },

  getDevice = async (uid) => {
    return http().get(`/devices/${uid}`);
  };