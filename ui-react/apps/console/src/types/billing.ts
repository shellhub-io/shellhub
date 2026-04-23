import type {
  GetCustomerResponses,
  GetSubscriptionResponses,
} from "../client/types.gen";

export type BillingStatus =
  | "inactive"
  | "active"
  | "trialing"
  | "incomplete"
  | "incomplete_expired"
  | "past_due"
  | "canceled"
  | "unpaid"
  | "paused"
  | "to_cancel_at_end_of_period";

export type InvoiceStatus =
  | "draft"
  | "open"
  | "paid"
  | "uncollectible"
  | "voided";

export type InvoiceCurrency = "usd" | "brl";

export interface Invoice {
  id: string;
  status: InvoiceStatus;
  currency: InvoiceCurrency;
  amount: number;
}

export interface PaymentMethod {
  id: string;
  number: string;
  brand: string;
  exp_month: number;
  exp_year: number;
  default: boolean;
}

export interface Customer {
  id: string;
  name: string;
  email: string;
  payment_methods: PaymentMethod[];
}

export interface Subscription {
  id: string;
  active: boolean;
  status: BillingStatus;
  end_at: number;
  invoices: Invoice[];
}

export interface NamespaceBilling {
  active?: boolean;
  status?: BillingStatus;
  customer_id?: string;
  subscription_id?: string;
  current_period_end?: number;
  created_at?: string;
  updated_at?: string;
}

type RawCustomer = NonNullable<GetCustomerResponses[200]>;
type RawSubscription = NonNullable<GetSubscriptionResponses[200]>;

/**
 * Narrow the generated (all-optional) customer response into a concrete shape,
 * filling in sensible defaults. The backend always returns at least `id` and
 * `email` for a namespace that has an active customer, but properties are
 * typed optional in the OpenAPI spec.
 */
export function toCustomer(
  raw: RawCustomer | undefined | null,
): Customer | null {
  if (!raw?.id) return null;
  return {
    id: raw.id,
    name: raw.name ?? "",
    email: raw.email ?? "",
    payment_methods: (raw.payment_methods ?? []).map((pm) => ({
      id: pm.id ?? "",
      number: pm.number ?? "",
      brand: pm.brand ?? "",
      exp_month: pm.exp_month ?? 0,
      exp_year: pm.exp_year ?? 0,
      default: pm.default ?? false,
    })),
  };
}

export function toSubscription(
  raw: RawSubscription | undefined | null,
): Subscription | null {
  if (!raw) return null;
  return {
    id: raw.id ?? "",
    active: raw.active ?? false,
    status: raw.status ?? "inactive",
    end_at: raw.end_at ?? 0,
    invoices: (raw.invoices ?? []).map((inv) => ({
      id: inv.id ?? "",
      status: inv.status ?? "draft",
      currency: inv.currency ?? "usd",
      amount: inv.amount ?? 0,
    })),
  };
}

/**
 * `Namespace.billing` is currently typed `{ [key: string]: unknown } | null`
 * in the generated client. This helper narrows it to `NamespaceBilling`
 * without unsafe casts at call sites.
 */
export function readNamespaceBilling(
  billing: { [key: string]: unknown } | null | undefined,
): NamespaceBilling | null {
  if (!billing || typeof billing !== "object") return null;
  const b = billing;
  return {
    active: typeof b.active === "boolean" ? b.active : undefined,
    status:
      typeof b.status === "string" ? (b.status as BillingStatus) : undefined,
    customer_id: typeof b.customer_id === "string" ? b.customer_id : undefined,
    subscription_id:
      typeof b.subscription_id === "string" ? b.subscription_id : undefined,
    current_period_end:
      typeof b.current_period_end === "number"
        ? b.current_period_end
        : undefined,
    created_at: typeof b.created_at === "string" ? b.created_at : undefined,
    updated_at: typeof b.updated_at === "string" ? b.updated_at : undefined,
  };
}
