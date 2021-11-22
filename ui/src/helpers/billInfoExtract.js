export default function infoExtract(data, periodEnd) {
  const latestInvoice = data.latest_invoice;
  const upcomingInvoice = data.upcoming_invoice;
  const productDescription = data.product_description;
  const pms = data.payment_methods;
  const pm = data.default_payment_method;

  const info = {
    periodEnd,
    description: productDescription,
    latestPaymentDue: latestInvoice.amount_due,
    latestPaymentPaid: latestInvoice.amount_paid,
    nextPaymentDue: upcomingInvoice.amount_due,
    nextPaymentPaid: upcomingInvoice.amount_paid,
  };

  const defaultCard = {
    brand: pm.card.brand,
    expYear: pm.card.exp_year,
    default: true,
    expMonth: pm.card.exp_month,
    last4: pm.card.last4,
    id: pm.id,
  };

  const cards = [
    defaultCard,
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
    defaultCard,
    cards,
  };
}
