import { describe, it, expect, beforeEach, vi } from "vitest";

// recentDevicesStore reads the active tenant from authStore; stub it with a
// mutable ref so each test can control which tenant is recording.
const { tenantRef } = vi.hoisted(() => ({
  tenantRef: { current: null as string | null },
}));

vi.mock("@/stores/authStore", () => ({
  useAuthStore: {
    getState: () => ({ tenant: tenantRef.current }),
  },
}));

import { useRecentDevicesStore } from "@/stores/recentDevicesStore";
import { useTerminalStore } from "@/stores/terminalStore";

const TENANT = "tenant-a";

const recentsFor = (tenant: string) =>
  useRecentDevicesStore.getState().byTenant[tenant] ?? [];

describe("recentDevicesStore", () => {
  beforeEach(() => {
    tenantRef.current = TENANT;
    useRecentDevicesStore.setState({ byTenant: {} });
    useTerminalStore.setState({ sessions: [], reconnectTarget: null });
    localStorage.clear();
  });

  it("records devices under the active tenant, most-recent-first", () => {
    const { record } = useRecentDevicesStore.getState();
    record("dev-1", "web-01");
    record("dev-2", "db-01");

    const recents = recentsFor(TENANT);
    expect(recents.map((d) => d.uid)).toEqual(["dev-2", "dev-1"]);
    expect(recents[0].name).toBe("db-01");
    expect(typeof recents[0].connectedAt).toBe("string");
  });

  it("dedupes a re-recorded device by moving it to the front", () => {
    const { record } = useRecentDevicesStore.getState();
    record("dev-1", "web-01");
    record("dev-2", "db-01");
    record("dev-1", "web-01");

    const recents = recentsFor(TENANT);
    expect(recents.map((d) => d.uid)).toEqual(["dev-1", "dev-2"]);
    expect(recents.filter((d) => d.uid === "dev-1")).toHaveLength(1);
  });

  it("caps each tenant's list at the store limit", () => {
    const { record } = useRecentDevicesStore.getState();
    for (let i = 0; i < 15; i++) record(`dev-${i}`, `name-${i}`);

    const recents = recentsFor(TENANT);
    expect(recents).toHaveLength(10);
    expect(recents[0].uid).toBe("dev-14"); // newest kept
    expect(recents.at(-1)?.uid).toBe("dev-5"); // oldest kept
  });

  it("keeps each tenant's list isolated", () => {
    const { record } = useRecentDevicesStore.getState();
    record("dev-1", "web-01");
    tenantRef.current = "tenant-b";
    record("dev-2", "db-01");

    expect(recentsFor("tenant-a").map((d) => d.uid)).toEqual(["dev-1"]);
    expect(recentsFor("tenant-b").map((d) => d.uid)).toEqual(["dev-2"]);
  });

  it("no-ops when there is no active tenant", () => {
    tenantRef.current = null;
    useRecentDevicesStore.getState().record("dev-1", "web-01");
    expect(useRecentDevicesStore.getState().byTenant).toEqual({});
  });

  it("records the device when a terminal opens (the connect choke point)", () => {
    useTerminalStore.getState().open({
      deviceUid: "dev-9",
      deviceName: "edge-01",
      username: "root",
      password: "",
    });

    expect(recentsFor(TENANT).map((d) => d.uid)).toEqual(["dev-9"]);
  });
});
