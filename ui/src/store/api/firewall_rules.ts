import { rulesApi } from "@/api/http";
import { IFirewallRule } from "@/interfaces/IFirewallRule";

export const postFirewall = async (data: IFirewallRule) => rulesApi.createFirewallRule({
  priority: data.priority,
  action: data.action,
  active: data.active,
  filter: data.filter,
  source_ip: data.source_ip,
  username: data.username,
});

export const fetchFirewalls = async (perPage: number, page: number) => rulesApi.getFirewallRules(page, perPage);

export const getFirewall = async (id: number) => rulesApi.getFirewallRule(id);

export const putFirewall = async (data: IFirewallRule & { id: number }) => rulesApi.updateFirewallRule(data.id, {
  priority: data.priority,
  action: data.action,
  active: data.active,
  filter: data.filter,
  source_ip: data.source_ip,
  username: data.username,
});

export const removeFirewall = async (id: number) => rulesApi.deleteFirewallRule(id);
