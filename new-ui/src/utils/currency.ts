export const formatCurrency = (value : number, currency ?: string) => {
  const valueFormated = value / 100;
  const fmt = Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: currency || 'USD',
  });
  return fmt.format(valueFormated);
};