import { describe, it, expect } from "vitest";
import { buildFilter } from "../useDevices";

describe("buildFilter", () => {
  describe("search only", () => {
    it("encodes a name OR custom_fields filter", () => {
      const result = JSON.parse(atob(buildFilter("my-device", [])));
      expect(result).toEqual([
        { type: "operator", params: { name: "or" } },
        { type: "property", params: { name: "name", operator: "contains", value: "my-device" } },
        { type: "operator", params: { name: "or" } },
        { type: "property", params: { name: "custom_fields", operator: "contains", value: "my-device" } },
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
        { type: "operator", params: { name: "or" } },
        { type: "property", params: { name: "name", operator: "contains", value: "srv" } },
        { type: "operator", params: { name: "or" } },
        { type: "property", params: { name: "custom_fields", operator: "contains", value: "srv" } },
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
