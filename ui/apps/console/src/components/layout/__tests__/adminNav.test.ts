import { describe, it, expect } from "vitest";
import { adminNavSectionTitle } from "../adminNav";

describe("adminNavSectionTitle", () => {
  it.each([
    ["/admin/dashboard", undefined],
    ["/admin/users", "Accounts"],
    ["/admin/namespaces", "Accounts"],
    ["/admin/devices", "Resources"],
    ["/admin/sessions", "Resources"],
    ["/admin/firewall-rules", "Resources"],
    ["/admin/settings/authentication", "Instance"],
    ["/admin/instance-api-keys", "Instance"],
    ["/admin/license", "Instance"],
  ])("places %s under %s", (route, title) => {
    expect(adminNavSectionTitle(route)).toBe(title);
  });
});
