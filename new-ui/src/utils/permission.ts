import { authorizer } from "../authorizer";

export default (role : string, action : string) => {
  const hasPermission = !!authorizer.permissions[role]
    && authorizer.permissions[role].includes(action);

  return hasPermission;
};
