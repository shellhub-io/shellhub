import { describe, it, expect } from "vitest";
import { hasPermission, isValidRole, type Permission, type Role } from "../permission";
import { rolePermissions } from "../permissions.generated";

const roles = Object.keys(rolePermissions) as Role[];
const everyPermission = [
  ...new Set<Permission>(Object.values(rolePermissions).flat()),
];

describe("isValidRole", () => {
  it.each(roles)("accepts %s", (role) => {
    expect(isValidRole(role)).toBe(true);
  });

  it.each([null, "", "admin", "superuser", "OWNER", "constructor", "toString", "hasOwnProperty"])(
    "rejects %j",
    (role) => {
      expect(isValidRole(role)).toBe(false);
    },
  );
});

describe("hasPermission", () => {
  it.each([null, "admin", "constructor"])("grants %j nothing", (role) => {
    everyPermission.forEach((permission) =>
      expect(hasPermission(role, permission)).toBe(false),
    );
  });

  it.each(roles)("grants %s exactly what the authorizer gives it", (role) => {
    const held: readonly Permission[] = rolePermissions[role];
    everyPermission.forEach((permission) =>
      expect(hasPermission(role, permission)).toBe(held.includes(permission)),
    );
  });

  it("keeps device connections from observers", () => {
    expect(hasPermission("observer", "DeviceConnect")).toBe(false);
  });

  it("keeps namespace deletion from administrators", () => {
    expect(hasPermission("administrator", "NamespaceDelete")).toBe(false);
    expect(hasPermission("owner", "NamespaceDelete")).toBe(true);
  });

  it("requires every permission in a list", () => {
    expect(hasPermission("operator", ["TagCreate", "TagDelete"])).toBe(true);
    expect(hasPermission("operator", ["TagCreate", "DeviceRemove"])).toBe(false);
  });
});
