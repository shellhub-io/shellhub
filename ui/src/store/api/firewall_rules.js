import http from '@/store/helpers/http';

export const postFirewall = async (data) => http().post('/firewall/rules', {
  priority: parseInt(data.priority, 10),
  action: data.policy,
  active: data.status === 'active',
  filter: data.filter,
  source_ip: data.source_ip,
  username: data.username,
});

export const fetchFirewalls = async (perPage, page) => http().get(`/firewall/rules?per_page=${perPage}&page=${page}`);

export const getFirewall = async (id) => http().get(`/firewall/rules/${id}`);

export const putFirewall = async (data) => http().put(`/firewall/rules/${data.id}`, {
  priority: parseInt(data.priority, 10),
  action: data.policy,
  filter: data.filter,
  active: data.status === 'active',
  source_ip: data.source_ip,
  username: data.username,
});

export const removeFirewall = async (id) => http().delete(`/firewall/rules/${id}`);
