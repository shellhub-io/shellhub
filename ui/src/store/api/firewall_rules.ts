import { rulesApi } from "@/api/http";
import { IFirewallRule } from "@/interfaces/IFirewallRule";

export const createFirewallRule = async (data: IFirewallRule) => rulesApi.createFirewallRule({
  priority: data.priority,
  action: data.action,
  active: data.active,
  filter: data.filter,
  source_ip: data.source_ip,
  username: data.username,
});

export const fetchFirewallRuleList = async (perPage: number, page: number) => rulesApi.getFirewallRules(page, perPage);

export const updateFirewallRule = async (data: IFirewallRule) => rulesApi.updateFirewallRule(data.id, {
  priority: data.priority,
  action: data.action,
  active: data.active,
  filter: data.filter,
  source_ip: data.source_ip,
  username: data.username,
});

export const removeFirewallRule = async (id: string) => rulesApi.deleteFirewallRule(id);
