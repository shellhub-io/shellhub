import { Filter } from "@/interfaces/IFilter";

export interface IFirewallRule {
  action: "allow" | "deny";
  active: boolean;
  filter: Filter;
  id: string;
  tenant_id: string;
  priority: number;
  source_ip: string;
  username: string;
}
