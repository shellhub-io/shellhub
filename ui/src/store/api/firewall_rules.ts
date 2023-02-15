import { rulesApi } from "../../api/http";

export const postFirewall = async (data: any) => rulesApi.createFirewallRule({
  priority: parseInt(data.priority, 10),
  action: data.policy,
  active: data.status === "active",
  filter: data.filter,
  source_ip: data.source_ip,
  username: data.username,
});

export const fetchFirewalls = async (perPage: any, page: any) => rulesApi.getFirewallRules(page, perPage);

export const getFirewall = async (id: any) => rulesApi.getFirewallRule(id);

export const putFirewall = async (data: any) => rulesApi.updateFirewallRule(data.id, {
  priority: parseInt(data.priority, 10),
  action: data.policy,
  filter: data.filter,
  active: data.status === "active",
  source_ip: data.source_ip,
  username: data.username,
});

export const removeFirewall = async (id: any) => rulesApi.deleteFirewallRule(id);
