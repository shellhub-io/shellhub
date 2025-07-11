export interface IPaymentMethod {
  id: string;
  number: string;
  brand: string;
  exp_month: number;
  exp_year: number;
  cvc: string;
  default: boolean;
}

export interface ICustomer {
  id: string,
  name: string,
  email: string,
  payment_methods: Array<IPaymentMethod>,
}
