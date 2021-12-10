import { authorizer } from '../../authorizer';

export default (role, action) => {
  const hasPermission = !!authorizer.permissions[role]
    && authorizer.permissions[role].includes(action);

  return hasPermission;
};
