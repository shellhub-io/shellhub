type Members = {
  id: string;
  role: "administrator" | "operator" | "observer" | "owner";
  username: string;
}

type Settings = {
  session_record: boolean;
}

type Billing = {
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

export interface INamespace {
  billing: Billing;
  created_at: string;
  devices_count: number;
  max_devices: number;
  members: Array<Members>;
  name: string;
  owner: string;
  settings: Settings;
  tenant_id: string;
}
