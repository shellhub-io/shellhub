import { rulesApi } from "@/api/http";
import { IFirewallRule } from "@/interfaces/IFirewallRule";

interface FilterHostnameType {
  hostname: string;
}

interface FilterTagsType {
  tags: string[];
}
interface FirewallRuleType {
  id: number;
  policy: "allow" | "deny";
  priority: string;
  status: string;
  source_ip: string;
  username: string;
  filter: FilterHostnameType | FilterTagsType;
}

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

export const putFirewall = async (data: FirewallRuleType) => rulesApi.updateFirewallRule(data.id, {
  priority: parseInt(data.priority, 10),
  action: data.policy,
  filter: data.filter,
  active: data.status === "active",
  source_ip: data.source_ip,
  username: data.username,
});

export const removeFirewall = async (id: number) => rulesApi.deleteFirewallRule(id);
