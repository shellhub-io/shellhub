// eslint-disable-next-line @typescript-eslint/ban-ts-comment
// @ts-nocheck
export default function infoExtract(data, periodEnd) {
  const { invoices, warning } = data;
  const upcomingInvoice = data.upcoming_invoice;
  const productDescription = data.product_description;
  const pms = data.payment_methods;

  const showLink = (r, s) => {
    if (s === "open") {
      return r;
    }

    return "---";
  };

  const parseInvoices = (invs) => {
    if (invs.length === 0) {
      return [];
    }

    return invs.reduce((ac, inv) => [...ac, {
      paid: inv.paid,
      status: inv.status,
      url: showLink(inv.hosted_invoice_url, inv.status),
      pdf: inv.invoice_pdf,
      dueDate: inv.due_date === 0 ? inv.created : inv.due_date,
      amountDue: inv.amount_due,
      attempted: inv.attempted,
      currency: inv.currency,
      accountCountry: inv.account_country,
    }], []);
  };

  const info = {
    periodEnd,
    description: productDescription,
    currency: upcomingInvoice.currency,
    accountCountry: upcomingInvoice.account_country,
    nextPaymentDue: upcomingInvoice.amount_due,
    nextPaymentPaid: upcomingInvoice.amount_paid,
  };

  const cards = [
    ...pms.map((v) => ({
      id: v.id,
      brand: v.card.brand,
      expYear: v.card.exp_year,
      default: false,
      expMonth: v.card.exp_month,
      last4: v.card.last4,
    })),
  ];

  return {
    info,
    cards,
    warning,
    invoices: parseInvoices(invoices),
  };
}
