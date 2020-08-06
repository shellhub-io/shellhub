import http from '@/store/helpers/http';

export const postFirewall = async (data) => http().post('/firewall/rules', {
  priority: parseInt(data.priority, 10),
  action: data.action,
  active: data.active,
  source_ip: data.source_ip,
  username: data.username,
  hostname: data.hostname,
});

export const fetchFirewalls = async () => http().get('/firewall/rules');

export const getFirewall = async (id) => http().get(`/firewall/rules/${id}`);

export const putFirewall = async (data) => http().put(`/firewall/rules/${data.id}`, {
  priority: parseInt(data.priority, 10),
  action: data.action,
  active: data.active,
  source_ip: data.source_ip,
  username: data.username,
  hostname: data.hostname,
});

export const removeFirewall = async (id) => http().delete(`/firewall/rules/${id}`);
