import { describe, it, expect } from "vitest";
import type { Device } from "@/client";
import { normalizeDeviceTags } from "@/utils/deviceTags";

function makeDevice(tags: unknown): Device {
  return { uid: "device-1", name: "edge-01", tags } as unknown as Device;
}

describe("normalizeDeviceTags", () => {
  it("flattens tag objects down to their names", () => {
    const device = makeDevice([{ name: "production" }, { name: "edge" }]);

    expect(normalizeDeviceTags(device).tags).toEqual(["production", "edge"]);
  });

  it("keeps tags that already arrive as plain strings", () => {
    const device = makeDevice(["production", "edge"]);

    expect(normalizeDeviceTags(device).tags).toEqual(["production", "edge"]);
  });

  it.each([
    ["undefined", undefined],
    ["null", null],
    ["a non-array value", "production"],
  ])("returns an empty tag list when tags are %s", (_label, tags) => {
    expect(normalizeDeviceTags(makeDevice(tags)).tags).toEqual([]);
  });

  it("leaves every other device field untouched", () => {
    const device = makeDevice([{ name: "production" }]);

    const normalized = normalizeDeviceTags(device);

    expect(normalized.uid).toBe("device-1");
    expect(normalized.name).toBe("edge-01");
  });
});
