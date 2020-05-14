import http from '@/helpers/http';

export const
  fetchDevices = async (perPage, page, search) => {
    let query = '';
    if (search === null) { 
      query = `/devices?per_page=${perPage}&page=${page}`;
    } else {
      query = `/devices?per_page=${perPage}&page=${page}&filter=${search}`;
    }
    return http().get(query);
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