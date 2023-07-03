export interface IBilling {
  id: string;
  active: boolean;
  status: string;
  end_at: string;
  invoices: Array<any>;
}
