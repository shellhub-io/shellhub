import timezoneMock from 'timezone-mock';
import { formatDate, formatDateWithoutDayAndHours } from '@/components/filter/date';

describe('Date', () => {
  const date = '2020-05-18T13:27:02.498Z';

  beforeEach(() => {
    timezoneMock.register('UTC');
  });

  it('Verify formatDate', () => {
    const actual = formatDate(date);

    expect(actual).toEqual('Monday, May 18th 2020, 1:27:02 pm');
  });
});

describe('Format date without day and hours', () => {
  it('Verify formatDate', () => {
    const actual = formatDateWithoutDayAndHours('2020-05-18T13:27:02.498Z');

    expect(actual).toEqual('May 18th 2020');
  });
});
