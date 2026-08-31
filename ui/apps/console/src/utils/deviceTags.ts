import type { Device } from "@/client";

/** A device whose tags are flattened to plain names, the shape the console renders. */
export type TaggedDevice = Omit<Device, "tags"> & { tags: string[] };

/**
 * Normalizes a device's tags to plain strings. The API has returned tags both as objects and as
 * bare strings across versions, and a missing field as well, so every caller downstream would
 * otherwise repeat the same three-way check.
 */
export function normalizeDeviceTags(device: Device): TaggedDevice {
  return {
    ...device,
    tags: Array.isArray(device.tags)
      ? device.tags.map((t) =>
          typeof t === "object" && t !== null && "name" in t
            ? t.name
            : String(t),
        )
      : [],
  };
}
