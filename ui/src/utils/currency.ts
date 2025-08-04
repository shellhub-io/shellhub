// This method was created because the API returns the base Int64 value.
// Example: Return 3224 value, and the price is $ 32.24.

const formatCurrency = (amount = 0, currency = "USD") => {
  const amountWithCents = amount / 100;
  const formatter = Intl.NumberFormat("en-US", {
    style: "currency",
    currency,
  });
  return formatter.format(amountWithCents);
};

export default formatCurrency;
