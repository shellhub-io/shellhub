export interface IInvoice {
  id: string;
  status: "open" | "draft" | "paid" | "uncollectible" | "voided";
  currency: "usd" | "brl";
  amount: number;
}

export interface IBilling {
  id: string;
  active: boolean;
  status: string;
  end_at: number;
  invoices: Array<IInvoice>;
}
