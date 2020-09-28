import { formatDate } from '@/components/filter/date';

describe('Date', () => {
  it('Verify formatDate', () => {
    const actual = formatDate('2020-05-18T13:27:02.498Z');
    expect(actual).toEqual('Monday, May 18th 2020, 1:27:02 pm');
  });
});
