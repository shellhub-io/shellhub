import formatCurrency from '@/components/filter/currency';

describe('Currency', () => {
  it('Verify format', () => {
    const actual = formatCurrency('0000');

    expect(actual).toEqual('$0.00');
  });
});
