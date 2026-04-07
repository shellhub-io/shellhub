import { describe, it, expect } from "vitest";
import { buildFilter } from "../useDevices";

describe("buildFilter", () => {
  describe("search only", () => {
    it("encodes a name filter", () => {
      const result = JSON.parse(atob(buildFilter("my-device", [])));
      expect(result).toEqual([
        { type: "property", params: { name: "name", operator: "contains", value: "my-device" } },
      ]);
    });
  });

  describe("tags only", () => {
    it("encodes a tags filter", () => {
      const result = JSON.parse(atob(buildFilter("", ["web", "prod"])));
      expect(result).toEqual([
        { type: "property", params: { name: "tags.name", operator: "contains", value: ["web", "prod"] } },
      ]);
    });
  });

  describe("search and tags combined", () => {
    it("encodes both filters in the same array", () => {
      const result = JSON.parse(atob(buildFilter("srv", ["prod"])));
      expect(result).toEqual([
        { type: "property", params: { name: "name", operator: "contains", value: "srv" } },
        { type: "property", params: { name: "tags.name", operator: "contains", value: ["prod"] } },
      ]);
    });
  });

  describe("empty inputs", () => {
    it("returns an empty filter array when both are empty", () => {
      const result = JSON.parse(atob(buildFilter("", [])));
      expect(result).toEqual([]);
    });
  });
});
