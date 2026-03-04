import { describe, it, expect } from "vitest";
import { hasAnyDevices } from "../stats";
import { Stats } from "@/types/stats";

const emptyStats: Stats = {
  registered_devices: 0,
  pending_devices: 0,
  rejected_devices: 0,
  online_devices: 0,
  active_sessions: 0,
};

describe("hasAnyDevices", () => {
  it("returns false for null", () => {
    expect(hasAnyDevices(null)).toBe(false);
  });

  it("returns false when all counts are zero", () => {
    expect(hasAnyDevices(emptyStats)).toBe(false);
  });

  it("returns true when registered_devices > 0", () => {
    expect(hasAnyDevices({ ...emptyStats, registered_devices: 1 })).toBe(true);
  });

  it("returns true when pending_devices > 0", () => {
    expect(hasAnyDevices({ ...emptyStats, pending_devices: 1 })).toBe(true);
  });

  it("returns true when rejected_devices > 0", () => {
    expect(hasAnyDevices({ ...emptyStats, rejected_devices: 1 })).toBe(true);
  });
});
