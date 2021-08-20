import capitalizeFirstLetter from '@/components/filter/string';

describe('Word', () => {
  it('Verify format', () => {
    const actual = capitalizeFirstLetter('shellHub');

    expect(actual).toEqual('ShellHub');
  });
});
