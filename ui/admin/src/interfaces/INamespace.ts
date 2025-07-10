import { INamespace } from "@/interfaces/INamespace";

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

export interface IAdminNamespace extends Omit<INamespace, "billing"> {
  billing?: IAdminBilling;
}
