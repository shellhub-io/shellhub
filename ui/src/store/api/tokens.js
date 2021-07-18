import http from '@/store/helpers/http';

export const postToken = async () => http().post('/tokens');

export const fetchTokens = async () => http().get('/tokens');

export const getToken = async (id) => http().get(`/tokens/${id}`);

export const putToken = async (data) => http().put(`/tokens/${data.id}/update`, {
  read_only: data.read_only,
});

export const removeToken = async (id) => http().delete(`/tokens/${id}/delete`);
