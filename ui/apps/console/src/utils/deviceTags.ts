import type { Device } from "@/client";

/** A device whose tags are flattened to plain names, the shape the console renders. */
export type TaggedDevice = Omit<Device, "tags"> & { tags: string[] };

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
