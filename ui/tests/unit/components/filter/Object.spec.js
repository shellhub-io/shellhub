import formatDeviceSort from '@/components/filter/object';

describe('Device', () => {
  it('Verify formatDeviceSort', () => {
    const inputData = [
      [undefined, undefined],
      ['online', false],
      ['online', true],
      ['hostname', false],
      ['hostname', true],
      ['tag', false],
      ['tag', true],
    ];

    const expectedData = [
      {
        field: null,
        status: false,
        statusString: 'asc',
      },
      {
        field: 'online',
        status: false,
        statusString: 'asc',
      },
      {
        field: 'online',
        status: true,
        statusString: 'desc',
      },
      {
        field: 'name',
        status: false,
        statusString: 'asc',
      },
      {
        field: 'name',
        status: true,
        statusString: 'desc',
      },
      {
        field: 'tag',
        status: false,
        statusString: 'asc',
      },
      {
        field: 'tag',
        status: true,
        statusString: 'desc',
      },
    ];

    inputData.forEach((input, index) => {
      const actual = formatDeviceSort(input[0], input[1]);

      expect(actual).toEqual(expectedData[index]);
    });
  });
});
