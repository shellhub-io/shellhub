import { INamespace, INamespaceMember } from "@/interfaces/INamespace";

type AdminUserStatus = "accepted" | "pending";
interface IAdminBilling {
  active: boolean;
  current_period_end: string;
  customer_id: string;
  payment_failed: null | string | boolean;
  payment_method_id: string;
  price_id: string;
  state: string;
  sub_item_id: string;
  subscription_id: string;
}

interface IAdminNamespaceMember extends Omit<INamespaceMember, "status"> {
  status: AdminUserStatus
}

export interface IAdminNamespace extends Omit<INamespace, "billing" | "members"> {
  billing?: IAdminBilling;
  members: IAdminNamespaceMember[];
}
