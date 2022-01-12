import http from '@/store/helpers/http';

export const postNamespace = async (data) => http().post('/namespaces', {
  name: data.name,
});

export const fetchNamespaces = async () => http().get('/namespaces');

export const getNamespace = async (id) => http().get(`/namespaces/${id}`);

export const removeNamespace = async (id) => http().delete(`/namespaces/${id}`);

export const putNamespace = async (data) => http().put(`/namespaces/${data.id}`, {
  name: data.name,
});

export const addUserToNamespace = async (data) => http().post(`/namespaces/${data.tenant_id}/members`, {
  username: data.username,
  role: data.role,
});

export const editUserToNamespace = async (data) => http().patch(`/namespaces/${data.tenant_id}/members/${data.user_id}`, {
  role: data.role,
});

export const removeUserFromNamespace = async (data) => http().delete(`/namespaces/${data.tenant_id}/members/${data.user_id}`);
export const tenantSwitch = async (data) => http().get(`/auth/token/${data.tenant_id}`);
