import { authorizer } from "../authorizer";

export default (role: string, action: string) => !!authorizer.role[role] && authorizer.permissions[role].includes(action);
