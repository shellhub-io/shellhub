import { authorizer } from '../../authorizer';

export default (accessType, action) => {
  const hasPermission = !!authorizer.permissions[accessType]
    && authorizer.permissions[accessType].includes(action);

  return hasPermission;
};
