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

export const addUserToNamespace = async (data) => http().patch(`/namespaces/${data.tenant_id}/add`, {
  username: data.username,
});

export const removeUserFromNamespace = async (data) => http().patch(`/namespaces/${data.tenant_id}/del`, {
  username: data.username,
});

export const tenantSwitch = async (data) => http().get(`/auth/token/${data.tenant_id}`);

export const webhookUpdate = async (data) => http().patch(`/namespaces/${data.tenant_id}/webhook`, {
  url: data.url,
});

export const webhookStatusUpdate = async (data) => http().patch(`/namespaces/${data.tenant_id}/webhook/activate`, {
  status: data.status,
});
