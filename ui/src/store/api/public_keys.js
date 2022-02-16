import http from '@/store/helpers/http';

export const postPublicKey = async (data) => http().post('/sshkeys/public-keys', data);

export const fetchPublicKeys = async (perPage, page) => http().get(`/sshkeys/public-keys?per_page=${perPage}&page=${page}`);

export const getPublicKey = async (fingerprint) => http().get(`/sshkeys/public-keys/${fingerprint}`);

export const putPublicKey = async (data) => http().put(`/sshkeys/public-keys/${data.fingerprint}`, data);

export const removePublicKey = async (fingerprint) => http().delete(`/sshkeys/public-keys/${fingerprint}`);
