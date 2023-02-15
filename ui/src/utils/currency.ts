// This method was created because the API returns the base Int64 value.
// Example: Return 3224 value, and the price is $ 32.24.

const formatCurrency = (value: number, currency?: string) => {
  const valueFormated = value / 100;
  const fmt = Intl.NumberFormat("en-US", {
    style: "currency",
    currency: currency || "USD",
  });
  return fmt.format(valueFormated);
};

export default formatCurrency;
