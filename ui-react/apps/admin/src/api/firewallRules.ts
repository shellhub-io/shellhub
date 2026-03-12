import apiClient, { getTotalCount } from "./client";
import { FirewallRule, FirewallFilter } from "../types/firewallRule";
import { PaginatedResponse } from "../types/api";

export async function getFirewallRules(
  page = 1,
  perPage = 10,
): Promise<PaginatedResponse<FirewallRule>> {
  const res = await apiClient.get<FirewallRule[]>("/api/firewall/rules", {
    params: { page, per_page: perPage },
  });
  return { data: res.data, totalCount: getTotalCount(res) };
}

export async function createFirewallRule(payload: {
  priority: number;
  action: string;
  active: boolean;
  source_ip: string;
  username: string;
  filter: FirewallFilter;
}): Promise<FirewallRule> {
  const { data } = await apiClient.post<FirewallRule>(
    "/api/firewall/rules",
    payload,
  );
  return data;
}

export async function updateFirewallRule(
  id: string,
  payload: {
    priority: number;
    action: string;
    active: boolean;
    source_ip: string;
    username: string;
    filter: FirewallFilter;
  },
): Promise<void> {
  await apiClient.put(`/api/firewall/rules/${id}`, payload);
}

export async function deleteFirewallRule(id: string): Promise<void> {
  await apiClient.delete(`/api/firewall/rules/${id}`);
}
