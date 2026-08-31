export type Role = "observer" | "operator" | "administrator" | "owner";

enum RoleLevel {
  OBSERVER = 1,
  OPERATOR = 2,
  ADMINISTRATOR = 3,
  OWNER = 4,
}

const roleLevels: Record<Role, RoleLevel> = {
  observer: RoleLevel.OBSERVER,
  operator: RoleLevel.OPERATOR,
  administrator: RoleLevel.ADMINISTRATOR,
  owner: RoleLevel.OWNER,
};

export const isValidRole = (role: string | null): role is Role =>
  role !== null && Object.prototype.hasOwnProperty.call(roleLevels, role);

const permissions = {
  "device:connect": RoleLevel.OBSERVER,
  "device:details": RoleLevel.OBSERVER,
  "device:add": RoleLevel.OPERATOR,
  "device:accept": RoleLevel.OPERATOR,
  "device:reject": RoleLevel.OPERATOR,
  "device:rename": RoleLevel.OPERATOR,
  "device:customField:update": RoleLevel.OPERATOR,
  "device:remove": RoleLevel.ADMINISTRATOR,
  "device:choose": RoleLevel.OWNER,

  "session:details": RoleLevel.OBSERVER,
  "session:play": RoleLevel.ADMINISTRATOR,
  "session:close": RoleLevel.ADMINISTRATOR,
  "session:removeRecord": RoleLevel.ADMINISTRATOR,

  "namespace:create": RoleLevel.OBSERVER,
  "namespace:leave": RoleLevel.OBSERVER,
  "namespace:rename": RoleLevel.ADMINISTRATOR,
  "namespace:addMember": RoleLevel.ADMINISTRATOR,
  "namespace:editMember": RoleLevel.ADMINISTRATOR,
  "namespace:removeMember": RoleLevel.ADMINISTRATOR,
  "namespace:cancelInvitation": RoleLevel.ADMINISTRATOR,
  "namespace:updateSessionRecording": RoleLevel.ADMINISTRATOR,
  "namespace:updateSshAccessMode": RoleLevel.ADMINISTRATOR,
  "namespace:delete": RoleLevel.OWNER,

  "tag:edit": RoleLevel.OPERATOR,
  "tag:remove": RoleLevel.OPERATOR,
  "tag:deviceCreate": RoleLevel.OPERATOR,
  "tag:update": RoleLevel.OPERATOR,

  "webEndpoint:create": RoleLevel.ADMINISTRATOR,
  "webEndpoint:delete": RoleLevel.ADMINISTRATOR,

  "connector:add": RoleLevel.ADMINISTRATOR,
  "connector:edit": RoleLevel.ADMINISTRATOR,
  "connector:remove": RoleLevel.ADMINISTRATOR,

  "firewall:create": RoleLevel.ADMINISTRATOR,
  "firewall:edit": RoleLevel.ADMINISTRATOR,
  "firewall:remove": RoleLevel.ADMINISTRATOR,

  "publicKey:create": RoleLevel.ADMINISTRATOR,
  "publicKey:edit": RoleLevel.ADMINISTRATOR,
  "publicKey:remove": RoleLevel.ADMINISTRATOR,

  "accessPolicy:create": RoleLevel.ADMINISTRATOR,
  "accessPolicy:edit": RoleLevel.ADMINISTRATOR,
  "accessPolicy:remove": RoleLevel.ADMINISTRATOR,

  "sshIdentity:add": RoleLevel.OPERATOR,
  "sshIdentity:manage": RoleLevel.ADMINISTRATOR,

  "serviceAccount:view": RoleLevel.ADMINISTRATOR,
  "serviceAccount:create": RoleLevel.ADMINISTRATOR,
  "serviceAccount:delete": RoleLevel.ADMINISTRATOR,

  "billing:subscribe": RoleLevel.OWNER,
  "billing:unsubscribe": RoleLevel.OWNER,

  "notification:view": RoleLevel.OPERATOR,

  "apiKey:create": RoleLevel.ADMINISTRATOR,
  "apiKey:edit": RoleLevel.ADMINISTRATOR,
  "apiKey:delete": RoleLevel.ADMINISTRATOR,

  "installKey:create": RoleLevel.ADMINISTRATOR,
  "installKey:reveal": RoleLevel.ADMINISTRATOR,
  "installKey:edit": RoleLevel.ADMINISTRATOR,
  "installKey:disable": RoleLevel.ADMINISTRATOR,
  "installKey:revoke": RoleLevel.ADMINISTRATOR,

  "namespace:editBanner": RoleLevel.ADMINISTRATOR,
} as const;

export type Action = keyof typeof permissions;

export function hasPermission(role: string | null, action: Action): boolean {
  if (!isValidRole(role)) return false;
  return roleLevels[role] >= permissions[action];
}
