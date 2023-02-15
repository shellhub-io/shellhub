export interface IBilling {
  subscription_id: string;
  current_period_end: string;
  price_id: string;
  customer_id: string;
  payment_method_id: string;
  payment_failed: string | null;
  state: string;
  active: boolean;
  sub_item_id: string;
}

interface IBillingDataInfo {
  periodEnd: string;
  description: string;
  currency: string;
  accountCountry: string;
  nextPaymentDue: number;
  nextPaymentPaid: number;
}

export interface IBIllingDataCard {
  brand: string;
  expYear: number;
  default: boolean;
  expMonth: number;
  last4: string;
  id: string;
}

export interface IBIllingDataInvoice {
  paid: boolean;
  status: string;
  url: string;
  pdf: string;
  dueDate: number;
  amountDue: number;
  attempted: boolean;
  currency: string;
  accountCountry: string;
}

export interface IBillingData {
  cards: IBIllingDataCard[];
  defaultCard: IBIllingDataCard;
  info: IBillingDataInfo;
  invoices: IBIllingDataInvoice[];
  warning: boolean;
}
