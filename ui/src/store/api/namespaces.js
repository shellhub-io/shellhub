import http from '@/store/helpers/http';

export const postNamespace = async (data) => http().post('/namespace', {
  name: data.name,
});

export const fetchNamespaces = async () => http().get('/namespace');

export const getNamespace = async (id) => http().get(`/namespace/${id}`);

export const removeNamespace = async (id) => http().delete(`/namespace/${id}`);

export const putNamespace = async (data) => http().put(`/namespace/${data.id}`, {
  name: data.name,
});

export const addUserToNamespace = async (data) => http().patch(`/namespace/${data.tenant_id}/add`, {
  username: data.username,
});

export const removeUserFromNamespace = async (data) => http().patch(`/namespace/${data.tenant_id}/del`, {
  username: data.username,
});

export const tenantSwitch = async (data) => http().get(`/auth/token/${data.tenant_id}`);
