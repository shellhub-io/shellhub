import useAuthStore from "@/store/modules/auth";

enum Roles {
  OBSERVER = 1,
  OPERATOR = 2,
  ADMINISTRATOR = 3,
  OWNER = 4,
}

const permissions = {
  "device:connect": Roles.OBSERVER,
  "device:details": Roles.OBSERVER,
  "device:add": Roles.OPERATOR,
  "device:accept": Roles.OPERATOR,
  "device:reject": Roles.OPERATOR,
  "device:rename": Roles.OPERATOR,
  "device:remove": Roles.ADMINISTRATOR,
  "device:choose": Roles.OWNER,

  "session:details": Roles.OBSERVER,
  "session:play": Roles.ADMINISTRATOR,
  "session:close": Roles.ADMINISTRATOR,
  "session:removeRecord": Roles.ADMINISTRATOR,

  "namespace:create": Roles.OBSERVER,
  "namespace:leave": Roles.OBSERVER,
  "namespace:rename": Roles.ADMINISTRATOR,
  "namespace:addMember": Roles.ADMINISTRATOR,
  "namespace:editMember": Roles.ADMINISTRATOR,
  "namespace:removeMember": Roles.ADMINISTRATOR,
  "namespace:updateSessionRecording": Roles.ADMINISTRATOR,
  "namespace:delete": Roles.OWNER,

  "tag:edit": Roles.OPERATOR,
  "tag:remove": Roles.OPERATOR,
  "tag:deviceCreate": Roles.OPERATOR,
  "tag:update": Roles.OPERATOR,

  "webEndpoint:create": Roles.ADMINISTRATOR,
  "webEndpoint:delete": Roles.ADMINISTRATOR,

  "connector:add": Roles.ADMINISTRATOR,
  "connector:edit": Roles.ADMINISTRATOR,
  "connector:remove": Roles.ADMINISTRATOR,

  "firewall:create": Roles.ADMINISTRATOR,
  "firewall:edit": Roles.ADMINISTRATOR,
  "firewall:remove": Roles.ADMINISTRATOR,

  "publicKey:create": Roles.ADMINISTRATOR,
  "publicKey:edit": Roles.ADMINISTRATOR,
  "publicKey:remove": Roles.ADMINISTRATOR,

  "billing:subscribe": Roles.OWNER,
  "billing:unsubscribe": Roles.OWNER,

  "notification:view": Roles.OPERATOR,

  "apiKey:create": Roles.ADMINISTRATOR,
  "apiKey:delete": Roles.ADMINISTRATOR,
} as const;

type Action = keyof typeof permissions;

const hasPermission = (action: Action) => {
  const { role } = useAuthStore();
  const roleLevel = Roles[role.toUpperCase()] || 0;
  const requiredLevel = permissions[action];

  return roleLevel >= requiredLevel;
};

export default hasPermission;
