import { IFirewallRule } from "@/interfaces/IFirewallRule";

export type IAdminFirewallRule = Omit<IFirewallRule, "status">;
