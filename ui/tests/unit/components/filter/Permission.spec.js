import permission from '../../../../src/components/filter/permission';
import { actions, authorizer } from '../../../../src/authorizer';

const cases = [
  [false, -1, -1],
  [false, 'valid', -1],
  [false, -1, 'valid'],
  [true, authorizer.accessType.observer, actions.device.connect],
  [false, authorizer.accessType.observer, actions.device.rename],
];

describe('Permission', () => {
  test.each(cases)('permission returns %s when access type is %s and action is %s', (expected, accessType, action) => {
    expect(permission(accessType, action, authorizer.permissions)).toEqual(expected);
  });
});
