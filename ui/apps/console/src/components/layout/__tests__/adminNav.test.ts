import { describe, it, expect, beforeEach } from "vitest";
import { setEdition } from "@/tests/edition";
import { adminNavSectionTitle } from "../adminNav";

describe("adminNavSectionTitle", () => {
  beforeEach(() => setEdition("enterprise"));

  it.each([
    ["/admin/instance", undefined],
    ["/admin/users", "Accounts"],
    ["/admin/namespaces", "Accounts"],
    ["/admin/settings/authentication", "Instance"],
    ["/admin/instance-api-keys", "Instance"],
    ["/admin/license", "Instance"],
    ["/admin/settings/telemetry", "Instance"],
  ])("places %s under %s", (route, title) => {
    expect(adminNavSectionTitle(route)).toBe(title);
  });

  it("offers telemetry on cloud, which has no authentication page", () => {
    setEdition("cloud");

    expect(adminNavSectionTitle("/admin/settings/telemetry")).toBe("Instance");
    expect(adminNavSectionTitle("/admin/settings/authentication")).toBeUndefined();
  });
});
