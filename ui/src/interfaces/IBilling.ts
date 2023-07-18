export interface IInvoices {
  id: string;
  status: "open" | "draft" | "paid" | "uncollectible" | "voided";
  currency: "usd" | "brl";
  amount: number;
}
export interface IBilling {
  id: string;
  active: boolean;
  status: string;
  end_at: string;
  invoices: Array<IInvoices>;
}
