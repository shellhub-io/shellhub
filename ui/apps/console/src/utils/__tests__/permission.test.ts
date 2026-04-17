import { describe, it, expect } from "vitest";
import { hasPermission, isValidRole, type Role, type Action } from "../permission";

/* ─── Action sets by minimum required role ─── */

const OBSERVER_ACTIONS: Action[] = [
  "device:connect", "device:details",
  "session:details",
  "namespace:create", "namespace:leave",
];

const OPERATOR_ACTIONS: Action[] = [
  "device:add", "device:accept", "device:reject", "device:rename",
  "tag:edit", "tag:remove", "tag:deviceCreate", "tag:update",
  "notification:view",
];

const ADMINISTRATOR_ACTIONS: Action[] = [
  "device:remove",
  "session:play", "session:close", "session:removeRecord",
  "namespace:rename",
  "namespace:addMember", "namespace:editMember", "namespace:removeMember",
  "namespace:editInvitation", "namespace:cancelInvitation",
  "namespace:updateSessionRecording", "namespace:editBanner",
  "namespace:updateAllowPassword", "namespace:updateAllowPublicKey",
  "namespace:updateAllowRoot", "namespace:updateAllowEmptyPasswords",
  "namespace:updateAllowTTY", "namespace:updateAllowTcpForwarding",
  "namespace:updateAllowWebEndpoints", "namespace:updateAllowSFTP",
  "namespace:updateAllowAgentForwarding",
  "publicKey:create", "publicKey:edit", "publicKey:remove",
  "firewall:create", "firewall:edit", "firewall:remove",
  "webEndpoint:create", "webEndpoint:delete",
  "connector:add", "connector:edit", "connector:remove",
  "apiKey:create", "apiKey:edit", "apiKey:delete",
];

const OWNER_ACTIONS: Action[] = [
  "device:choose",
  "namespace:delete",
  "billing:subscribe", "billing:unsubscribe",
];

const ALL_ACTIONS: Action[] = [
  ...OBSERVER_ACTIONS,
  ...OPERATOR_ACTIONS,
  ...ADMINISTRATOR_ACTIONS,
  ...OWNER_ACTIONS,
];

/* ─── isValidRole ─── */

describe("isValidRole", () => {
  it("returns true for each of the four valid roles", () => {
    const validRoles: Role[] = ["observer", "operator", "administrator", "owner"];
    validRoles.forEach((role) => expect(isValidRole(role)).toBe(true));
  });

  it("returns false for null", () => {
    expect(isValidRole(null)).toBe(false);
  });

  it("returns false for empty string", () => {
    expect(isValidRole("")).toBe(false);
  });

  it("returns false for common invalid strings", () => {
    ["admin", "superuser", "user", "OWNER"].forEach((s) =>
      expect(isValidRole(s)).toBe(false),
    );
  });

  it("returns false for prototype-chain keys", () => {
    // 'in' would return true for these; hasOwnProperty must not
    ["constructor", "toString", "hasOwnProperty"].forEach((s) =>
      expect(isValidRole(s)).toBe(false),
    );
  });
});

/* ─── hasPermission ─── */

describe("hasPermission", () => {
  describe("null or invalid role", () => {
    it("denies all actions when role is null", () => {
      ALL_ACTIONS.forEach((action) =>
        expect(hasPermission(null, action)).toBe(false),
      );
    });

    it("denies all actions when role is an invalid string", () => {
      expect(hasPermission("admin", "device:connect")).toBe(false);
      expect(hasPermission("superuser", "namespace:delete")).toBe(false);
    });
  });

  describe("observer role (level 1)", () => {
    it("allows observer-level actions", () => {
      OBSERVER_ACTIONS.forEach((action) =>
        expect(hasPermission("observer", action)).toBe(true),
      );
    });

    it("denies operator-level actions", () => {
      OPERATOR_ACTIONS.forEach((action) =>
        expect(hasPermission("observer", action)).toBe(false),
      );
    });

    it("denies administrator-level actions", () => {
      ADMINISTRATOR_ACTIONS.forEach((action) =>
        expect(hasPermission("observer", action)).toBe(false),
      );
    });

    it("denies owner-level actions", () => {
      OWNER_ACTIONS.forEach((action) =>
        expect(hasPermission("observer", action)).toBe(false),
      );
    });
  });

  describe("operator role (level 2)", () => {
    it("allows observer-level actions", () => {
      OBSERVER_ACTIONS.forEach((action) =>
        expect(hasPermission("operator", action)).toBe(true),
      );
    });

    it("allows operator-level actions", () => {
      OPERATOR_ACTIONS.forEach((action) =>
        expect(hasPermission("operator", action)).toBe(true),
      );
    });

    it("denies administrator-level actions", () => {
      ADMINISTRATOR_ACTIONS.forEach((action) =>
        expect(hasPermission("operator", action)).toBe(false),
      );
    });

    it("denies owner-level actions", () => {
      OWNER_ACTIONS.forEach((action) =>
        expect(hasPermission("operator", action)).toBe(false),
      );
    });
  });

  describe("administrator role (level 3)", () => {
    it("allows observer-level actions", () => {
      OBSERVER_ACTIONS.forEach((action) =>
        expect(hasPermission("administrator", action)).toBe(true),
      );
    });

    it("allows operator-level actions", () => {
      OPERATOR_ACTIONS.forEach((action) =>
        expect(hasPermission("administrator", action)).toBe(true),
      );
    });

    it("allows administrator-level actions", () => {
      ADMINISTRATOR_ACTIONS.forEach((action) =>
        expect(hasPermission("administrator", action)).toBe(true),
      );
    });

    it("denies owner-level actions", () => {
      OWNER_ACTIONS.forEach((action) =>
        expect(hasPermission("administrator", action)).toBe(false),
      );
    });
  });

  describe("owner role (level 4)", () => {
    it("allows all actions", () => {
      ALL_ACTIONS.forEach((action) =>
        expect(hasPermission("owner", action)).toBe(true),
      );
    });
  });

  describe("boundary conditions — minimum required role", () => {
    it.each<[Action, Role, Role]>([
      ["device:accept", "observer", "operator"],
      ["device:remove", "operator", "administrator"],
      ["namespace:delete", "administrator", "owner"],
      ["billing:subscribe", "administrator", "owner"],
    ])("%s: %s denied, %s allowed", (action, denied, allowed) => {
      expect(hasPermission(denied, action)).toBe(false);
      expect(hasPermission(allowed, action)).toBe(true);
    });
  });
});
